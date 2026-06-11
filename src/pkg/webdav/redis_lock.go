package webdav

import (
	"context"
	"encoding/json"
	"fmt"
	"myobj/src/config"
	"myobj/src/pkg/logger"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/net/webdav"
)

const (
	lockKeyPrefix    = "myobj:webdav:lock:"
	defaultLockTTL   = 1 * time.Hour
	maxLockTTL       = 24 * time.Hour
)

// unlockScript Lua 脚本：原子性地验证 token 并删除锁
// KEYS[1] = 锁的 Redis key
// ARGV[1] = 期望的 token
// 返回: 1 = 成功删除, 0 = token 不匹配或锁不存在
var unlockScript = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
    return redis.call("DEL", KEYS[1])
end
return 0
`)

// redisLock 存储在 Redis 中的锁信息
type redisLock struct {
	Token     string        `json:"token"`
	Root      string        `json:"root"`
	Duration  time.Duration `json:"duration"`
	OwnerXML  string        `json:"owner_xml"`
	ZeroDepth bool          `json:"zero_depth"`
	CreatedAt time.Time     `json:"created_at"`
}

// RedisLockSystem 基于 Redis 的分布式锁系统
type RedisLockSystem struct {
	client *redis.Client
}

// NewRedisLockSystem 创建基于 Redis 的分布式锁系统
func NewRedisLockSystem(client *redis.Client) *RedisLockSystem {
	return &RedisLockSystem{client: client}
}

// NewRedisClientFromConfig 从配置创建 Redis 客户端
func NewRedisClientFromConfig(cfg *config.Cache) (*redis.Client, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}
	return redisClient, nil
}

// lockKey 生成锁的 Redis key
func lockKey(token string) string {
	return lockKeyPrefix + token
}

// generateToken 生成唯一的锁 token
func generateToken() string {
	return "opaquelocktoken:" + uuid.New().String()
}

// calcLockTTL 计算锁的 TTL
func calcLockTTL(duration time.Duration) time.Duration {
	if duration <= 0 {
		// 负数或零表示无限期锁，使用默认 TTL
		return defaultLockTTL
	}
	if duration > maxLockTTL {
		return maxLockTTL
	}
	return duration
}

// Create 创建一个新的锁
func (rls *RedisLockSystem) Create(now time.Time, details webdav.LockDetails) (string, error) {
	ctx := context.Background()
	token := generateToken()
	ttl := calcLockTTL(details.Duration)

	lock := &redisLock{
		Token:     token,
		Root:      details.Root,
		Duration:  details.Duration,
		OwnerXML:  details.OwnerXML,
		ZeroDepth: details.ZeroDepth,
		CreatedAt: now,
	}

	data, err := json.Marshal(lock)
	if err != nil {
		return "", fmt.Errorf("failed to marshal lock data: %w", err)
	}

	// SET NX EX：仅当 key 不存在时设置，保证原子性
	ok, err := rls.client.SetNX(ctx, lockKey(token), data, ttl).Result()
	if err != nil {
		return "", fmt.Errorf("redis SET NX failed: %w", err)
	}
	if !ok {
		// 理论上不会发生（UUID 唯一），但作为安全检查
		return "", webdav.ErrLocked
	}

	logger.LOG.Debug("WebDAV 锁已创建",
		"token", token,
		"root", details.Root,
		"ttl", ttl.String(),
	)

	return token, nil
}

// Refresh 刷新锁的 TTL
func (rls *RedisLockSystem) Refresh(now time.Time, token string, duration time.Duration) (webdav.LockDetails, error) {
	ctx := context.Background()
	key := lockKey(token)

	// 获取当前锁数据
	data, err := rls.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return webdav.LockDetails{}, webdav.ErrNoSuchLock
		}
		return webdav.LockDetails{}, fmt.Errorf("redis GET failed: %w", err)
	}

	var lock redisLock
	if err := json.Unmarshal(data, &lock); err != nil {
		return webdav.LockDetails{}, fmt.Errorf("failed to unmarshal lock data: %w", err)
	}

	// 计算新的 TTL（从当前时间开始）
	newTTL := calcLockTTL(duration)

	// 更新锁数据中的 Duration 和 CreatedAt
	lock.Duration = duration
	lock.CreatedAt = now
	updatedData, err := json.Marshal(lock)
	if err != nil {
		return webdav.LockDetails{}, fmt.Errorf("failed to marshal lock data: %w", err)
	}

	// 原子性地更新锁数据和 TTL
	// 使用 SET EX 覆盖写入（锁已存在，SET NX 不适用）
	if err := rls.client.Set(ctx, key, updatedData, newTTL).Err(); err != nil {
		return webdav.LockDetails{}, fmt.Errorf("redis SET failed: %w", err)
	}

	logger.LOG.Debug("WebDAV 锁已刷新",
		"token", token,
		"new_ttl", newTTL.String(),
	)

	return webdav.LockDetails{
		Root:      lock.Root,
		Duration:  duration,
		OwnerXML:  lock.OwnerXML,
		ZeroDepth: lock.ZeroDepth,
	}, nil
}

// Unlock 解锁（使用 Lua 脚本保证原子性）
func (rls *RedisLockSystem) Unlock(now time.Time, token string) error {
	ctx := context.Background()
	key := lockKey(token)

	// 使用 Lua 脚本原子性地验证 token 并删除锁
	result, err := unlockScript.Run(ctx, rls.client, []string{key}, token).Int()
	if err != nil {
		if err == redis.Nil {
			return webdav.ErrNoSuchLock
		}
		return fmt.Errorf("redis unlock script failed: %w", err)
	}

	if result == 0 {
		// token 不匹配或锁不存在
		// 检查 key 是否存在以区分错误类型
		exists, err := rls.client.Exists(ctx, key).Result()
		if err != nil {
			return fmt.Errorf("redis EXISTS failed: %w", err)
		}
		if exists == 0 {
			return webdav.ErrNoSuchLock
		}
		return webdav.ErrForbidden
	}

	logger.LOG.Debug("WebDAV 锁已释放", "token", token)
	return nil
}

// Confirm 确认锁的有效性
// 允许锁定不存在的文件（permissiveLockSystem 语义）
func (rls *RedisLockSystem) Confirm(now time.Time, name0, name1 string, conditions ...webdav.Condition) (func(), error) {
	if len(conditions) == 0 {
		// 无锁条件：允许访问未锁定的资源
		return func() {}, nil
	}

	ctx := context.Background()

	// 检查每个条件的 token 是否有效
	for _, cond := range conditions {
		if cond.Token == "" {
			continue
		}

		key := lockKey(cond.Token)
		data, err := rls.client.Get(ctx, key).Bytes()
		if err != nil {
			if err == redis.Nil {
				// token 无效，继续检查下一个条件集
				// WebDAV handler 会尝试其他条件集
				return nil, webdav.ErrConfirmationFailed
			}
			return nil, fmt.Errorf("redis GET failed: %w", err)
		}

		var lock redisLock
		if err := json.Unmarshal(data, &lock); err != nil {
			return nil, fmt.Errorf("failed to unmarshal lock data: %w", err)
		}

		// 验证锁是否覆盖请求的资源
		if !lockCoversResource(lock.Root, name0) || !lockCoversResource(lock.Root, name1) {
			return nil, webdav.ErrConfirmationFailed
		}
	}

	// 所有条件验证通过
	return func() {}, nil
}

// lockCoversResource 检查锁是否覆盖指定的资源
// 锁覆盖：资源路径等于锁根路径，或是锁根路径的子路径
func lockCoversResource(lockRoot, resource string) bool {
	if resource == "" {
		return true
	}
	if lockRoot == resource {
		return true
	}
	// 检查 resource 是否是 lockRoot 的子路径
	// lockRoot = "/dir/" 覆盖 resource = "/dir/file.txt"
	if len(resource) > len(lockRoot) && resource[:len(lockRoot)] == lockRoot {
		return true
	}
	return false
}

// Close 关闭 Redis 客户端连接
func (rls *RedisLockSystem) Close() error {
	if rls.client != nil {
		return rls.client.Close()
	}
	return nil
}

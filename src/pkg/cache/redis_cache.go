package cache

import (
	"context"
	"fmt"
	"myobj/src/config"
	"myobj/src/pkg/logger"
	"time"

	"github.com/redis/go-redis/v9"
)

const redisKeyPrefix = "myobj:"

type RedisCache struct {
	red *redis.Client
}

func NewRedisCache(cfg *config.Cache) (*RedisCache, error) {
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
	return &RedisCache{
		red: redisClient,
	}, nil
}

func (r *RedisCache) prefixedKey(key string) string {
	return redisKeyPrefix + key
}

func (r *RedisCache) Get(key string) (any, error) {
	val, err := r.red.Get(context.Background(), r.prefixedKey(key)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("key %s not found", key)
		}
		return nil, err
	}
	return val, nil
}

func (r *RedisCache) Set(key string, value any, expire int) error {
	return r.red.Set(context.Background(), r.prefixedKey(key), value, time.Duration(expire)*time.Second).Err()
}
func (r *RedisCache) Delete(key string) error {
	return r.red.Del(context.Background(), r.prefixedKey(key)).Err()
}
func (r *RedisCache) Stop() {
	err := r.red.Close()
	if err != nil {
		logger.LOG.Error("关闭缓存连接失败", "error", err)
		return
	}
}
func (r *RedisCache) Clear() {
	ctx := context.Background()
	var cursor uint64
	for {
		keys, nextCursor, err := r.red.Scan(ctx, cursor, redisKeyPrefix+"*", 100).Result()
		if err != nil {
			break
		}
		if len(keys) > 0 {
			r.red.Del(ctx, keys...)
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
}

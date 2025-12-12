package cache

import (
	"context"
	"fmt"
	"myobj/src/config"
	"myobj/src/pkg/logger"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	red *redis.Client
}

func NewRedisCache(cfg *config.Cache) *RedisCache {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})
	ctx := context.Background()
	pone, err := redisClient.Ping(ctx).Result() // 发送Ping命令检查连接
	if err != nil {
		logger.LOG.Error("Redis连接失败", "error", err, "pone", pone)
		panic(err) // 如果连接失败，抛出panic
	}
	return &RedisCache{
		red: redisClient,
	}
}

func (r *RedisCache) Get(key string) (any, error) {
	val := r.red.Get(context.Background(), key).Val()
	return val, nil
}

func (r *RedisCache) Set(key string, value any, expire int) error {
	return r.red.Set(context.Background(), key, value, time.Duration(expire)*time.Second).Err()
}
func (r *RedisCache) Delete(key string) error {
	return r.red.Del(context.Background(), key).Err()
}
func (r *RedisCache) Stop() {
	err := r.red.Close()
	if err != nil {
		logger.LOG.Error("关闭缓存连接失败", "error", err)
		return
	}
}
func (r *RedisCache) Clear() {
	r.red.FlushDB(context.Background())
}

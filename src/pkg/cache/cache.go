package cache

import (
	"myobj/src/config"
	"myobj/src/pkg/logger"
)

// Cache 缓存接口定义
type Cache interface {
	Get(key string) (any, error)
	Set(key string, value any, expire int) error
	Delete(key string) error
	Stop()
	Clear()
}

func InitCache() Cache {
	cfg := config.GetConfig().Cache
	if cfg.Type == "redis" {
		logger.LOG.Info("使用 Redis 缓存")
		return NewRedisCache(&cfg)
	}
	return NewLocalCache()
}

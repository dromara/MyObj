package util

import (
	"fmt"
	"myobj/src/pkg/cache"

	"github.com/goccy/go-json"
)

// CacheGetJSON 从缓存获取并反序列化 JSON。
// 兼容 LocalCache 返回对象指针和 RedisCache 返回 JSON 字符串两种情况。
func CacheGetJSON(c cache.Cache, key string, target interface{}) error {
	data, err := c.Get(key)
	if err != nil {
		return err
	}
	switch v := data.(type) {
	case string:
		return json.Unmarshal([]byte(v), target)
	default:
		// LocalCache 可能直接返回对象指针，尝试 JSON 序列化再反序列化以统一处理
		bytes, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("cache value marshal failed: %w", err)
		}
		return json.Unmarshal(bytes, target)
	}
}

// CacheSetJSON 将对象序列化为 JSON 并存入缓存
func CacheSetJSON(c cache.Cache, key string, value interface{}, ttl int) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.Set(key, string(bytes), ttl)
}

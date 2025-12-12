package cache

import (
	"fmt"
	"sync"
	"time"
)

// LocalCacheData 缓存数据项
type LocalCacheData struct {
	data   any
	expire time.Time
}

// LocalCache 本地缓存实现
type LocalCache struct {
	data            map[string]LocalCacheData
	lock            sync.RWMutex
	nextExpire      time.Time     // 下一个最近的过期时间
	stopChan        chan struct{} // 停止信号
	ticker          *time.Ticker  // 定时清理器
	cleanupInterval time.Duration // 定时清理间隔
}

// NewLocalCache 创建一个新的本地缓存实例
// cleanupInterval: 定时清理间隔，默认为1分钟
func NewLocalCache(cleanupInterval ...time.Duration) *LocalCache {
	interval := 1 * time.Minute
	if len(cleanupInterval) > 0 && cleanupInterval[0] > 0 {
		interval = cleanupInterval[0]
	}

	local := &LocalCache{
		data:            make(map[string]LocalCacheData),
		stopChan:        make(chan struct{}),
		cleanupInterval: interval,
	}

	// 启动定时清理协程
	go local.startCleanupRoutine()

	return local
}

// startCleanupRoutine 启动后台定时清理协程
func (c *LocalCache) startCleanupRoutine() {
	c.ticker = time.NewTicker(c.cleanupInterval)
	defer c.ticker.Stop()

	for {
		select {
		case <-c.ticker.C:
			c.cleanupExpired() // 定时删除
		case <-c.stopChan:
			return
		}
	}
}

// Get 获取缓存数据（惰性删除：访问时检查过期）
func (c *LocalCache) Get(key string) (any, error) {
	c.lock.RLock()
	data, ok := c.data[key]
	c.lock.RUnlock()

	if !ok {
		return nil, fmt.Errorf("key %s not found", key)
	}

	// 惰性删除：检查是否过期
	if time.Now().After(data.expire) {
		c.lock.Lock()
		delete(c.data, key)
		c.lock.Unlock()
		return nil, fmt.Errorf("key %s expired", key)
	}

	return data.data, nil
}

// Set 设置缓存数据，智能判断到期时间并调整清理策略
// expire: 过期时间（秒）
func (c *LocalCache) Set(key string, value any, expire int) error {
	expireTime := time.Now().Add(time.Duration(expire) * time.Second)

	c.lock.Lock()
	defer c.lock.Unlock()

	c.data[key] = LocalCacheData{
		data:   value,
		expire: expireTime,
	}

	// 智能判断：如果新添加的缓存过期时间早于当前记录的最近过期时间，更新并调整定时器
	if c.nextExpire.IsZero() || expireTime.Before(c.nextExpire) {
		c.nextExpire = expireTime
		c.adjustCleanupTiming(expireTime)
	}

	return nil
}

// Delete 删除指定key的缓存
func (c *LocalCache) Delete(key string) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	delete(c.data, key)
	return nil
}

// cleanupExpired 定时清理过期缓存
func (c *LocalCache) cleanupExpired() {
	c.lock.Lock()
	defer c.lock.Unlock()

	now := time.Now()
	var nextExpire time.Time

	// 遍历所有缓存项，删除过期项并找出下一个最近的过期时间
	for key, item := range c.data {
		if now.After(item.expire) {
			// 删除过期项
			delete(c.data, key)
		} else {
			// 更新下一个最近的过期时间
			if nextExpire.IsZero() || item.expire.Before(nextExpire) {
				nextExpire = item.expire
			}
		}
	}

	c.nextExpire = nextExpire
}

// adjustCleanupTiming 智能调整清理时机
// 如果新的过期时间很近，可以触发一次即时清理
func (c *LocalCache) adjustCleanupTiming(expireTime time.Time) {
	// 如果新的过期时间在下一个清理周期之前，且距离现在很近（比如小于清理间隔的一半）
	timeUntilExpire := time.Until(expireTime)
	if timeUntilExpire > 0 && timeUntilExpire < c.cleanupInterval/2 {
		// 在一个新的goroutine中延迟执行清理，不阻塞Set操作
		go func(d time.Duration) {
			time.Sleep(d)
			c.cleanupExpired()
		}(timeUntilExpire)
	}
}

// Stop 停止缓存的后台清理协程
func (c *LocalCache) Stop() {
	close(c.stopChan)
}

// Size 返回当前缓存中的项数
func (c *LocalCache) Size() int {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return len(c.data)
}

// Clear 清空所有缓存
func (c *LocalCache) Clear() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.data = make(map[string]LocalCacheData)
	c.nextExpire = time.Time{}
}

package cloudsync

import (
	"fmt"
	"myobj/src/pkg/cloudsync/internal"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

const listCacheTTL = 30 * time.Second

type listCacheEntry struct {
	files     []CloudFile
	total     int
	expiresAt time.Time
}

var (
	listCacheMu sync.RWMutex
	listCache   = make(map[string]*listCacheEntry)
	listFlight  singleflight.Group
)

// ListFilesCached 带短期缓存与请求合并的目录列表
func ListFilesCached(provider CloudProvider, providerID, credential, pdirFid string, page, size int) ([]CloudFile, int, error) {
	key := fmt.Sprintf("%s:%s:%s:%d:%d", providerID, internal.HashCredential(credential), pdirFid, page, size)

	listCacheMu.RLock()
	if entry, ok := listCache[key]; ok && time.Now().Before(entry.expiresAt) {
		files := append([]CloudFile(nil), entry.files...)
		total := entry.total
		listCacheMu.RUnlock()
		return files, total, nil
	}
	listCacheMu.RUnlock()

	v, err, _ := listFlight.Do(key, func() (interface{}, error) {
		listCacheMu.RLock()
		if entry, ok := listCache[key]; ok && time.Now().Before(entry.expiresAt) {
			listCacheMu.RUnlock()
			return entry, nil
		}
		listCacheMu.RUnlock()

		files, total, err := provider.ListFiles(pdirFid, page, size)
		if err != nil {
			return nil, err
		}

		entry := &listCacheEntry{
			files:     append([]CloudFile(nil), files...),
			total:     total,
			expiresAt: time.Now().Add(listCacheTTL),
		}
		listCacheMu.Lock()
		listCache[key] = entry
		listCacheMu.Unlock()
		return entry, nil
	})
	if err != nil {
		return nil, 0, err
	}
	entry := v.(*listCacheEntry)
	return append([]CloudFile(nil), entry.files...), entry.total, nil
}

// InvalidateListCache 清除指定 Provider 凭据的列表缓存
func InvalidateListCache(providerID, credential string) {
	prefix := providerID + ":" + internal.HashCredential(credential) + ":"
	listCacheMu.Lock()
	for k := range listCache {
		if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			delete(listCache, k)
		}
	}
	listCacheMu.Unlock()
}

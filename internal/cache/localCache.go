package cache

import (
	"SnapLink/internal/custom_err"
	cache2 "SnapLink/pkg/cache"
	"github.com/dgraph-io/ristretto"
	"github.com/pkg/errors"
	"github.com/zhufuyi/sponge/pkg/logger"
	"sync"
)

var instanceLocalCache struct {
	cache *ristretto.Cache
	once  sync.Once
}

// LocalCache 单例模式设置本地缓存
func LocalCache() cache2.ILocalCache {
	instanceLocalCache.once.Do(func() {
		var err error
		instanceLocalCache.cache, err = ristretto.NewCache(&ristretto.Config{
			NumCounters: 1e7,     // number of keys to track frequency of (10M).
			MaxCost:     1 << 30, // maximum cost of bfCache (1GB).
			BufferItems: 64,      // number of keys per Get buffer.
		})
		if err != nil {
			logger.Panic(errors.Wrap(custom_err.ErrCacheInitFailed, "LocalCache").Error())
		}
	})
	return instanceLocalCache.cache
}

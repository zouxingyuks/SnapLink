package cache

import (
	"github.com/dgraph-io/ristretto"
	"github.com/pkg/errors"
	"github.com/zhufuyi/sponge/pkg/logger"
	"sync"
)

var instance struct {
	cache *ristretto.Cache
	once  sync.Once
}

// LocalCache 单例模式设置本地缓存
func LocalCache() *ristretto.Cache {
	instance.once.Do(func() {
		var err error
		instance.cache, err = ristretto.NewCache(&ristretto.Config{
			NumCounters: 1e7,     // number of keys to track frequency of (10M).
			MaxCost:     1 << 30, // maximum cost of cache (1GB).
			BufferItems: 64,      // number of keys per Get buffer.
		})
		if err != nil {
			logger.Panic(errors.Wrap(err, "init LocalCache failed").Error())
		}
	})
	return instance.cache
}

package hyperloglog

import (
	"SnapLink/internal/config"
	"context"
	"github.com/redis/go-redis/v9"
	"sync"
	"time"
)

var defaultInstance struct {
	cache Cache
	once  sync.Once
}

func DefaultCache() Cache {

	defaultInstance.once.Do(func() {
		opt := config.Get().PFRedis
		defaultInstance.cache = NewCache(redis.NewClient(&redis.Options{
			Addr:         opt.Addr,
			Password:     opt.Password,
			DB:           opt.DB,
			ReadTimeout:  time.Duration(opt.ReadTimeout) * time.Second,
			WriteTimeout: time.Duration(opt.WriteTimeout) * time.Second,
			DialTimeout:  time.Duration(opt.DialTimeout) * time.Second,
		}))
	})
	return defaultInstance.cache
}

// PFAdd 添加元素
func PFAdd(ctx context.Context, key string, values ...any) error {
	return DefaultCache().PFAdd(ctx, key, values...)
}

// PFCount 获取基数
func PFCount(ctx context.Context, keys ...string) (int64, error) {
	return DefaultCache().PFCount(ctx, keys...)

}

// PFMerge 合并多个hyperloglog
func PFMerge(ctx context.Context, destKey string, sourceKeys ...string) error {
	return DefaultCache().PFMerge(ctx, destKey, sourceKeys...)
}

// Delete 删除
func Delete(ctx context.Context, key string) error {
	return DefaultCache().Delete(ctx, key)
}

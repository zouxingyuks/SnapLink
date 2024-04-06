package cache

import (
	"SnapLink/internal/config"
	"SnapLink/internal/custom_err"
	cache2 "SnapLink/pkg/cache"
	"context"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"github.com/zhufuyi/sponge/pkg/logger"
	"sync"
	"time"
)

const (
	// BloomFilterCachePrefixKey  布隆过滤器缓存前缀
	BloomFilterCachePrefixKey = "BFCache"
)

var instanceBloomFilterCache = new(bfCache)

type bfCache struct {
	bfCache cache2.BFCache
	once    sync.Once
}

func BFCache() *bfCache {
	instanceBloomFilterCache.once.Do(func() {
		opt := config.Get().BFRedis
		var err error
		instanceBloomFilterCache.bfCache, err = cache2.NewBloomFilterCache(redis.NewClient(&redis.Options{
			Addr:         opt.Addr,
			Password:     opt.Password,
			DB:           opt.DB,
			ReadTimeout:  time.Duration(opt.ReadTimeout) * time.Second,
			WriteTimeout: time.Duration(opt.WriteTimeout) * time.Second,
			DialTimeout:  time.Duration(opt.DialTimeout) * time.Second,
		}), cache2.NewKeyGenerator(BloomFilterCachePrefixKey))
		if err != nil {
			logger.Panic(errors.Wrap(custom_err.ErrCacheInitFailed, "BloomFilterCache").Error())
		}
	})
	return instanceBloomFilterCache
}

// BFCreate 创建布隆过滤器
func (c *bfCache) BFCreate(ctx context.Context, key string, errorRate float64, capacity int) error {
	return c.bfCache.Create(ctx, key, errorRate, capacity)
}

// BFAdd 添加值
func (c *bfCache) BFAdd(ctx context.Context, key string, value string) error {
	return c.bfCache.Add(ctx, key, value)
}

// BFMAdd 批量添加多个值
func (c *bfCache) BFMAdd(ctx context.Context, key string, values ...string) error {
	return c.bfCache.MAdd(ctx, key, values...)
}

// BFExists 检查值是否存在
func (c *bfCache) BFExists(ctx context.Context, key string, value string) (bool, error) {
	return c.bfCache.Exists(ctx, key, value)
}

// BFMExists 批量检查多个值是否存在
func (c *bfCache) BFMExists(ctx context.Context, key string, values ...string) ([]bool, error) {
	return c.bfCache.MExists(ctx, key, values...)
}

// BFRename 重命名
func (c *bfCache) BFRename(ctx context.Context, key string, newKey string) error {
	return c.bfCache.Rename(ctx, key, newKey)
}

// BFDelete 删除布隆过滤器
func (c *bfCache) BFDelete(ctx context.Context, key string) error {
	return c.bfCache.Delete(ctx, key)
}

// BFInfo 获取布隆过滤器的信息
func (c *bfCache) BFInfo(ctx context.Context, key string) (items, capacity int64, err error) {
	return c.bfCache.Info(ctx, key)
}

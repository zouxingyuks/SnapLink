package cache

import (
	"SnapLink/internal/custom_err"
	"SnapLink/internal/model"
	cache2 "SnapLink/pkg/cache"
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"github.com/zhufuyi/sponge/pkg/logger"
	"strconv"
	"sync"
	"time"
)

const (
	// ShortLinkGroupCountExpireTime expire time
	ShortLinkGroupCountExpireTime = 10 * time.Minute
	ShortLinkGroupCountPrefix     = "sl_count"
	EmptyShortLinkGroupCount      = -1
)

var shortLinkGroupCountInstance = new(struct {
	*shortLinkGroupCountCache
	sync.Once
})

func ShortLinkGroupCountCache() *shortLinkGroupCountCache {
	shortLinkGroupCountInstance.Once.Do(func() {
		var err error
		if shortLinkGroupCountInstance.shortLinkGroupCountCache, err = NewShortLinkGroupCountCache(model.GetRedisCli(), nil); err != nil {
			logger.Panic(errors.Wrap(ErrInitCacheFailed, "ShortLinkGroupCountCache").Error())
		}
	})
	return shortLinkGroupCountInstance.shortLinkGroupCountCache
}

// shortLinkGroupCountCache define a cache struct
type shortLinkGroupCountCache struct {
	kvCache cache2.IKVCache
}

// NewShortLinkGroupCountCache new a cache
func NewShortLinkGroupCountCache(client *redis.Client, localCache cache2.ILocalCache) (*shortLinkGroupCountCache, error) {
	var err error
	cache := new(shortLinkGroupCountCache)
	cache.kvCache, err = cache2.NewKVCache(client, cache2.NewKeyGenerator(ShortLinkGroupCountPrefix), localCache)
	return cache, err
}

// Get 获取分组下的短链接数量
func (c *shortLinkGroupCountCache) Get(ctx context.Context, gid string) (int64, error) {
	value, err := c.kvCache.Get(ctx, gid)
	if errors.Is(err, cache2.ErrCacheNotFound) {
		return 0, custom_err.ErrCacheNotFound
	}
	if value == cache2.EmptyValue {
		return 0, nil
	}
	count, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, err
	}
	if count == EmptyShortLinkGroupCount {
		return 0, custom_err.ErrCacheNotFound
	}
	return count, err

}

// Set 设置分组下的短链接数量
func (c *shortLinkGroupCountCache) Set(ctx context.Context, gid string, count int64) error {
	value := strconv.Itoa(int(count))
	return c.kvCache.Set(ctx, gid, value, ShortLinkGroupCountExpireTime)
}

// SetCacheWithNotFound 设置空值来防御缓存穿透
func (c *shortLinkGroupCountCache) SetCacheWithNotFound(ctx context.Context, gid string) error {
	return c.kvCache.SetEmpty(ctx, gid, ShortLinkGroupCountExpireTime)
}

// Del 删除数据键值
func (c *shortLinkGroupCountCache) Del(ctx context.Context, gid string) error {
	return c.kvCache.Del(ctx, gid)
}

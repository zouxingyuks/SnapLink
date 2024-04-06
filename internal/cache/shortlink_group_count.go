package cache

import (
	"SnapLink/internal/custom_err"
	"SnapLink/internal/model"
	cache2 "SnapLink/pkg/cache"
	"context"
	"github.com/pkg/errors"
	"github.com/zhufuyi/sponge/pkg/logger"
	"strconv"
	"sync"
	"time"
)

const (
	ShortLinkGroupCountExpireTime = 10 * time.Minute
	ShortLinkGroupCountPrefix     = "sl_count"
)

var (
	shortLinkGroupCountInstance = new(shortLinkGroupCountCache)
	emptyCount                  = int64(0)
)

func ShortLinkGroupCountCache() *shortLinkGroupCountCache {
	shortLinkGroupCountInstance.once.Do(func() {
		var err error
		if shortLinkGroupCountInstance.kvCache, err = cache2.NewKVCache(model.GetRedisCli(), cache2.NewKeyGenerator(ShortLinkGroupCountPrefix), LocalCache()); err != nil {
			logger.Panic(errors.Wrap(custom_err.ErrCacheInitFailed, "ShortLinkGroupCountCache").Error())
		}
	})
	return shortLinkGroupCountInstance
}

// shortLinkGroupCountCache define a cache struct
type shortLinkGroupCountCache struct {
	kvCache cache2.IKVCache
	once    sync.Once
}

// Get 获取分组下的短链接数量
func (c *shortLinkGroupCountCache) Get(ctx context.Context, gid string) (int64, error) {
	value, err := c.kvCache.Get(ctx, gid)
	if errors.Is(err, cache2.ErrKVCacheNotFound) {
		return emptyCount, custom_err.ErrCacheNotFound
	}
	if value == cache2.EmptyValue {
		return emptyCount, nil
	}
	count, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return emptyCount, errors.Wrap(custom_err.ErrCacheGetFailed, err.Error())
	}
	return count, nil

}

// Set 设置分组下的短链接数量
func (c *shortLinkGroupCountCache) Set(ctx context.Context, gid string, count int64) error {
	value := strconv.Itoa(int(count))
	if err := c.kvCache.Set(ctx, gid, value, ShortLinkGroupCountExpireTime); err != nil {
		return errors.Wrap(custom_err.ErrCacheSetFailed, err.Error())
	}
	return nil
}

// SetCacheWithNotFound 设置空值来防御缓存穿透
func (c *shortLinkGroupCountCache) SetCacheWithNotFound(ctx context.Context, gid string) error {
	if err := c.kvCache.SetCacheWithNotFound(ctx, gid, ShortLinkGroupCountExpireTime); err != nil {
		return errors.Wrap(custom_err.ErrCacheSetFailed, err.Error())
	}
	return nil
}

// Del 删除数据键值
func (c *shortLinkGroupCountCache) Del(ctx context.Context, gid string) error {
	if err := c.kvCache.Del(ctx, gid); err != nil {
		return errors.Wrap(custom_err.ErrCacheDelFailed, err.Error())
	}
	return nil
}

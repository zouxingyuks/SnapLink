package cache

import (
	"SnapLink/internal/custom_err"
	"SnapLink/internal/model"
	cache2 "SnapLink/pkg/cache"
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/zhufuyi/sponge/pkg/logger"
	"sync"
	"time"
)

const (
	RedirectsExpireTime    = 10 * time.Minute
	RedirectCachePrefixKey = "redirect"
)

var redirectInstance = new(redirectsCache)

func Redirect() *redirectsCache {
	redirectInstance.once.Do(func() {
		var err error
		if redirectInstance.kvCache, err = cache2.NewKVCache(model.GetRedisCli(), cache2.NewKeyGenerator(RedirectCachePrefixKey), LocalCache()); err != nil {
			logger.Panic(errors.Wrap(custom_err.ErrCacheInitFailed, "RedirectsCache").Error())
		}
	})
	return redirectInstance
}

var emptyRedirect = new(model.Redirect)

// redirectsCache define a cache struct
type redirectsCache struct {
	kvCache cache2.IKVCache
	once    sync.Once
}

// Set write to cache
func (c *redirectsCache) Set(ctx context.Context, uri string, redirect *model.Redirect, ttl time.Duration) error {
	jsonBytes, err := json.Marshal(redirect)
	if err != nil {
		return errors.Wrap(custom_err.ErrCacheSetFailed, err.Error())
	}
	if err = c.kvCache.Set(ctx, uri, string(jsonBytes), ttl); err != nil {
		return errors.Wrap(custom_err.ErrCacheSetFailed, err.Error())
	}
	return nil
}

// Get 获取缓存
func (c *redirectsCache) Get(ctx context.Context, uri string) (*model.Redirect, error) {
	value, err := c.kvCache.Get(ctx, uri)
	if errors.Is(err, cache2.ErrCacheNotFound) {
		return nil, custom_err.ErrCacheNotFound
	}
	if value == cache2.EmptyValue {
		return emptyRedirect, nil
	}
	redirect := new(model.Redirect)
	err = json.Unmarshal([]byte(value), redirect)
	if err != nil {
		return nil, errors.Wrap(custom_err.ErrCacheGetFailed, err.Error())
	}
	return redirect, nil
}

// Del 删除缓存
func (c *redirectsCache) Del(ctx context.Context, uri string) error {
	if err := c.kvCache.Del(ctx, uri); err != nil {
		return errors.Wrap(custom_err.ErrCacheDelFailed, err.Error())
	}
	return nil
}

// SetCacheWithNotFound 设置不存在的缓存，以防止缓存穿透，默认过期时间 10 分钟
func (c *redirectsCache) SetCacheWithNotFound(ctx context.Context, uri string) error {
	if err := c.kvCache.SetEmpty(ctx, uri, RedirectsExpireTime); err != nil {
		return errors.Wrap(custom_err.ErrCacheSetFailed, err.Error())
	}
	return nil
}

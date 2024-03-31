package cache

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-redis/redis/v8"
	"time"

	"SnapLink/internal/model"
)

const (
	// cache prefix key for redirects
	redirectsCachePrefixKey = "redirects:"
	// RedirectsExpireTime 默认过期时间
	RedirectsExpireTime = 10 * time.Minute
	// RedirectsNeverExpireTime 永不过期
	RedirectsNeverExpireTime = 0
)

// RedirectsCache slCache interface
type RedirectsCache interface {
	Set(ctx context.Context, uri string, info *model.Redirect, duration time.Duration) error
	Get(ctx context.Context, uri string) (*model.Redirect, error)
	SetCacheWithNotFound(ctx context.Context, uri string) error
}

var emptyRedirectBytes = []byte(`{"uri":"","original_url":""}`)

// redirectsCache define a cache struct
type redirectsCache struct {
	client *redis.Client
}

// NewRedirectsCache new a cache
func NewRedirectsCache(cacheType *model.CacheType) RedirectsCache {
	return &redirectsCache{
		client: cacheType.Rdb,
	}
}

// GetRedirectsCacheKey cache key
func (c *redirectsCache) GetRedirectsCacheKey(uri string) string {
	return redirectsCachePrefixKey + uri
}

// Set write to cache
func (c *redirectsCache) Set(ctx context.Context, uri string, redirect *model.Redirect, duration time.Duration) error {
	if redirect == nil || uri == "" {
		return nil
	}
	key := c.GetRedirectsCacheKey(uri)
	jsonBytes, _ := json.Marshal(redirect)
	err := c.client.Set(ctx, key, jsonBytes, duration).Err()
	return err
}

// Get 获取缓存
func (c *redirectsCache) Get(ctx context.Context, uri string) (*model.Redirect, error) {
	key := c.GetRedirectsCacheKey(uri)
	jsonBytes, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, model.ErrCacheNotFound
		}
		return nil, err
	}
	redirect := new(model.Redirect)
	err = json.Unmarshal(jsonBytes, redirect)
	if err != nil {
		return nil, err
	}
	return redirect, nil
}

// SetCacheWithNotFound 设置不存在的缓存，以防止缓存穿透，默认过期时间 10 分钟
func (c *redirectsCache) SetCacheWithNotFound(ctx context.Context, uri string) error {
	key := c.GetRedirectsCacheKey(uri)
	err := c.client.Set(ctx, key, emptyRedirectBytes, RedirectsExpireTime).Err()
	if err != nil {
		return err
	}
	return nil
}

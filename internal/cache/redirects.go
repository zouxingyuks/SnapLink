package cache

import (
	"SnapLink/internal/model"
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"time"
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

// NewRedirectsCache 新建短链接的缓存
func NewRedirectsCache(client *redis.Client) (RedirectsCache, error) {
	var err error
	cache := &redirectsCache{
		client: client,
	}

	if err != nil {
		return nil, err
	}
	return cache, nil
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

	// 设置本地缓存
	_ = LocalCache().Set(key, jsonBytes, 1)
	// 设置分布式缓存
	err := c.client.Set(ctx, key, jsonBytes, duration).Err()
	return err
}

// Get 获取缓存
func (c *redirectsCache) Get(ctx context.Context, uri string) (*model.Redirect, error) {
	key := c.GetRedirectsCacheKey(uri)
	redirect := new(model.Redirect)
	var (
		jsonBytes []byte
		err       error
		ok        bool
	)

	//从本地缓存查询
	if data, exist := LocalCache().Get(key); exist {
		jsonBytes, ok = (data).([]byte)
	}
	//如果本地缓存没查到
	if !ok {
		//从分布式缓存查询
		jsonBytes, err = c.client.Get(ctx, key).Bytes()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				return nil, model.ErrCacheNotFound
			}
			return nil, err
		}
		//更新数据到本地缓存
		_ = LocalCache().Set(key, jsonBytes, 1)
	}

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

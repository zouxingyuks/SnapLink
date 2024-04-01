package cache

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"strconv"
	"time"

	"SnapLink/internal/model"
)

const (
	// ShortLinkExpireTime expire time
	ShortLinkExpireTime = 10 * time.Minute
	emptySLCount        = -1
)

var _ ShortLinkCache = (*shortLinkCache)(nil)

// ShortLinkCache cache interface
type ShortLinkCache interface {
	GetCount(ctx context.Context, gid string) (int64, error)
	SetCacheWithNotFound(ctx context.Context, gid string) error
	SetCount(ctx context.Context, gid string, count int64) error
}

// shortLinkCache define a cache struct
type shortLinkCache struct {
	client *redis.Client
}

// NewShortLinkCache new a cache
func NewShortLinkCache(client *redis.Client) (ShortLinkCache, error) {
	cache := &shortLinkCache{
		client: client,
	}
	return cache, nil
}

// SetCount 设置分组下的短链接数量
func (c *shortLinkCache) SetCount(ctx context.Context, gid string, count int64) error {
	cacheKey := makeSLGroupKey(gid)
	return c.client.Set(ctx, cacheKey, count, ShortLinkExpireTime).Err()
}

// GetCount 获取分组下的短链接数量
func (c *shortLinkCache) GetCount(ctx context.Context, gid string) (int64, error) {
	cacheKey := makeSLGroupKey(gid)
	str, err := c.client.Get(ctx, cacheKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, model.ErrCacheNotFound
		}
		return 0, err
	}
	count, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, err
	}
	if count == emptySLCount {
		return 0, model.ErrCacheNotFound
	}
	return count, err
}

// SetCacheWithNotFound 设置空值来防御缓存穿透
func (c *shortLinkCache) SetCacheWithNotFound(ctx context.Context, gid string) error {
	cacheKey := makeSLGroupKey(gid)
	return c.client.Set(ctx, cacheKey, emptySLCount, ShortLinkExpireTime).Err()
}

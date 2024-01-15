package cache

import (
	"context"
	"strings"
	"time"

	"SnapLink/internal/model"

	"github.com/zhufuyi/sponge/pkg/cache"
	"github.com/zhufuyi/sponge/pkg/encoding"
	"github.com/zhufuyi/sponge/pkg/utils"
)

const (
	// cache prefix key, must end with a colon
	shortLinkCachePrefixKey = "shortLink:"
	// ShortLinkExpireTime expire time
	ShortLinkExpireTime = 10 * time.Minute
)

var _ ShortLinkCache = (*shortLinkCache)(nil)

// ShortLinkCache cache interface
type ShortLinkCache interface {
	Set(ctx context.Context, id uint64, data *model.ShortLink, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.ShortLink, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ShortLink, error)
	MultiSet(ctx context.Context, data []*model.ShortLink, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetCacheWithNotFound(ctx context.Context, id uint64) error
}

// shortLinkCache define a cache struct
type shortLinkCache struct {
	cache cache.Cache
}

// NewShortLinkCache new a cache
func NewShortLinkCache(cacheType *model.CacheType) ShortLinkCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""
	var c cache.Cache
	if strings.ToLower(cacheType.CType) == "redis" {
		c = cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.ShortLink{}
		})
	} else {
		c = cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.ShortLink{}
		})
	}

	return &shortLinkCache{
		cache: c,
	}
}

// GetShortLinkCacheKey cache key
func (c *shortLinkCache) GetShortLinkCacheKey(id uint64) string {
	return shortLinkCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *shortLinkCache) Set(ctx context.Context, id uint64, data *model.ShortLink, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetShortLinkCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *shortLinkCache) Get(ctx context.Context, id uint64) (*model.ShortLink, error) {
	var data *model.ShortLink
	cacheKey := c.GetShortLinkCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *shortLinkCache) MultiSet(ctx context.Context, data []*model.ShortLink, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetShortLinkCacheKey(uint64(v.ID))
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *shortLinkCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.ShortLink, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetShortLinkCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.ShortLink)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.ShortLink)
	for _, id := range ids {
		val, ok := itemMap[c.GetShortLinkCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *shortLinkCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetShortLinkCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetCacheWithNotFound set empty cache
func (c *shortLinkCache) SetCacheWithNotFound(ctx context.Context, id uint64) error {
	cacheKey := c.GetShortLinkCacheKey(id)
	err := c.cache.SetCacheWithNotFound(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

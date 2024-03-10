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
	tUserCachePrefixKey = "tUser:"
	// TUserExpireTime expire time
	TUserExpireTime = 5 * time.Minute
)

var _ TUserCache = (*tUserCache)(nil)

// TUserCache cache interface
type TUserCache interface {
	Set(ctx context.Context, id uint64, data *model.TUser, duration time.Duration) error
	Get(ctx context.Context, id uint64) (*model.TUser, error)
	MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.TUser, error)
	MultiSet(ctx context.Context, data []*model.TUser, duration time.Duration) error
	Del(ctx context.Context, id uint64) error
	SetCacheWithNotFound(ctx context.Context, id uint64) error
}

// tUserCache define a cache struct
type tUserCache struct {
	cache cache.Cache
}

// NewTUserCache new a cache
func NewTUserCache(cacheType *model.CacheType) TUserCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.TUser{}
		})
		return &tUserCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.TUser{}
		})
		return &tUserCache{cache: c}
	}

	return nil // no cache
}

// GetTUserCacheKey cache key
func (c *tUserCache) GetTUserCacheKey(id uint64) string {
	return tUserCachePrefixKey + utils.Uint64ToStr(id)
}

// Set write to cache
func (c *tUserCache) Set(ctx context.Context, id uint64, data *model.TUser, duration time.Duration) error {
	if data == nil || id == 0 {
		return nil
	}
	cacheKey := c.GetTUserCacheKey(id)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *tUserCache) Get(ctx context.Context, id uint64) (*model.TUser, error) {
	var data *model.TUser
	cacheKey := c.GetTUserCacheKey(id)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *tUserCache) MultiSet(ctx context.Context, data []*model.TUser, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetTUserCacheKey(v.ID)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *tUserCache) MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.TUser, error) {
	var keys []string
	for _, v := range ids {
		cacheKey := c.GetTUserCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.TUser)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[uint64]*model.TUser)
	for _, id := range ids {
		val, ok := itemMap[c.GetTUserCacheKey(id)]
		if ok {
			retMap[id] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *tUserCache) Del(ctx context.Context, id uint64) error {
	cacheKey := c.GetTUserCacheKey(id)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetCacheWithNotFound set empty cache
func (c *tUserCache) SetCacheWithNotFound(ctx context.Context, id uint64) error {
	cacheKey := c.GetTUserCacheKey(id)
	err := c.cache.SetCacheWithNotFound(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

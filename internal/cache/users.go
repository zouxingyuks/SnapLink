package cache

import (
	"context"
	"strings"
	"time"

	"SnapLink/internal/model"

	"github.com/zhufuyi/sponge/pkg/cache"
	"github.com/zhufuyi/sponge/pkg/encoding"
)

const (
	// cache prefix key, must end with a colon
	usersCachePrefixKey = "users:"
	// UsersExpireTime expire time
	UsersExpireTime = 10 * time.Minute
)

var _ UsersCache = (*usersCache)(nil)

// UsersCache cache interface
type UsersCache interface {
	Set(ctx context.Context, username string, data *model.Users, duration time.Duration) error
	Get(ctx context.Context, username string) (*model.Users, error)
	MultiGet(ctx context.Context, usernames []string) (map[string]*model.Users, error)
	MultiSet(ctx context.Context, data []*model.Users, duration time.Duration) error
	Del(ctx context.Context, username string) error
	SetCacheWithNotFound(ctx context.Context, username string) error
}

// usersCache define a cache struct
type usersCache struct {
	cache cache.Cache
}

// NewUsersCache new a cache
func NewUsersCache(cacheType *model.CacheType) UsersCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""
	var c cache.Cache
	if strings.ToLower(cacheType.CType) == "redis" {
		c = cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.Users{}
		})
	} else {
		c = cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.Users{}
		})
	}

	return &usersCache{
		cache: c,
	}
}

// GetUsersCacheKey cache key
func (c *usersCache) GetUsersCacheKey(username string) string {
	return usersCachePrefixKey + username
}

// Set write to cache
func (c *usersCache) Set(ctx context.Context, username string, data *model.Users, duration time.Duration) error {
	if data == nil || username == "" {
		return nil
	}
	cacheKey := c.GetUsersCacheKey(username)
	err := c.cache.Set(ctx, cacheKey, data, duration)
	if err != nil {
		return err
	}
	return nil
}

// Get cache value
func (c *usersCache) Get(ctx context.Context, username string) (*model.Users, error) {
	var data *model.Users
	cacheKey := c.GetUsersCacheKey(username)
	err := c.cache.Get(ctx, cacheKey, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MultiSet multiple set cache
func (c *usersCache) MultiSet(ctx context.Context, data []*model.Users, duration time.Duration) error {
	valMap := make(map[string]interface{})
	for _, v := range data {
		cacheKey := c.GetUsersCacheKey(v.Username)
		valMap[cacheKey] = v
	}

	err := c.cache.MultiSet(ctx, valMap, duration)
	if err != nil {
		return err
	}

	return nil
}

// MultiGet multiple get cache, return key in map is id value
func (c *usersCache) MultiGet(ctx context.Context, usernames []string) (map[string]*model.Users, error) {
	var keys []string
	for _, v := range usernames {
		cacheKey := c.GetUsersCacheKey(v)
		keys = append(keys, cacheKey)
	}

	itemMap := make(map[string]*model.Users)
	err := c.cache.MultiGet(ctx, keys, itemMap)
	if err != nil {
		return nil, err
	}

	retMap := make(map[string]*model.Users)
	for _, username := range usernames {
		val, ok := itemMap[c.GetUsersCacheKey(username)]
		if ok {
			retMap[username] = val
		}
	}

	return retMap, nil
}

// Del delete cache
func (c *usersCache) Del(ctx context.Context, username string) error {
	cacheKey := c.GetUsersCacheKey(username)
	err := c.cache.Del(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

// SetCacheWithNotFound set empty cache
func (c *usersCache) SetCacheWithNotFound(ctx context.Context, username string) error {
	cacheKey := c.GetUsersCacheKey(username)
	err := c.cache.SetCacheWithNotFound(ctx, cacheKey)
	if err != nil {
		return err
	}
	return nil
}

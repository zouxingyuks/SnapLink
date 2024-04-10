package cache

import (
	"SnapLink/internal/custom_err"
	"SnapLink/internal/model"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/zhufuyi/sponge/pkg/logger"
	cache2 "github.com/zouxingyuks/common_pkg/cache"
	"sync"
	"time"
)

import (
	"context"
)

const (
	// cache prefix key, must end with a colon
	shortLinkGroupsCachePrefixKey = "groups"
	// ShortLinkGroupExpireTime expire time
	ShortLinkGroupExpireTime = 1 * time.Hour
)

var (
	slGroupInstance      = new(shortLinkGroupsCache)
	emptyShortLinkGroups = make([]*model.ShortLinkGroup, 0)
)

func SLGroup() *shortLinkGroupsCache {
	slGroupInstance.once.Do(func() {
		var err error
		keyGen := cache2.NewKeyGenerator(shortLinkGroupsCachePrefixKey)
		if slGroupInstance.kvCache, err = cache2.NewKVCache(model.GetRedisCli(), nil, cache2.WithKeyGen(keyGen)); err != nil {
			logger.Panic(errors.Wrap(custom_err.ErrCacheInitFailed, "ShortLinkGroupCache").Error())
		}
	})
	return slGroupInstance
}

// shortLinkGroupsCache define a bfCache struct
type shortLinkGroupsCache struct {
	kvCache cache2.IKVCache
	once    sync.Once
}

// Set 设置用户的 groups 缓存
func (c *shortLinkGroupsCache) Set(ctx context.Context, username string, groups []*model.ShortLinkGroup) error {
	jsonBytes, err := json.Marshal(groups)
	if err != nil {
		return errors.Wrap(custom_err.ErrCacheSetFailed, err.Error())
	}
	if err := c.kvCache.Set(ctx, username, string(jsonBytes), ShortLinkGroupExpireTime); err != nil {
		return errors.Wrap(custom_err.ErrCacheSetFailed, err.Error())
	}
	return nil
}

// SetCacheWithNotFound 设置用户的 groups 缓存为空
func (c *shortLinkGroupsCache) SetCacheWithNotFound(ctx context.Context, username string) error {
	if err := c.kvCache.SetCacheWithNotFound(ctx, username, ShortLinkGroupCountExpireTime); err != nil {
		return errors.Wrap(custom_err.ErrCacheSetFailed, err.Error())
	}
	return nil
}

// Get 获取用户的 groups 缓存
func (c *shortLinkGroupsCache) Get(ctx context.Context, username string) ([]*model.ShortLinkGroup, error) {
	value, err := c.kvCache.Get(ctx, username)
	if errors.Is(err, cache2.ErrKVCacheNotFound) {
		return emptyShortLinkGroups, custom_err.ErrCacheNotFound
	}
	if value == cache2.KVCacheEmptyValue {
		return emptyShortLinkGroups, nil
	}

	groups := make([]*model.ShortLinkGroup, 0)

	err = json.Unmarshal([]byte(value), &groups)
	if err != nil {
		return nil, err
	}

	if err != nil {
		return emptyShortLinkGroups, errors.Wrap(custom_err.ErrCacheGetFailed, err.Error())
	}
	return groups, nil
}

// Del 删除用户的 hash 缓存
func (c *shortLinkGroupsCache) Del(ctx context.Context, username string) error {
	if err := c.kvCache.Del(ctx, username); err != nil {
		return errors.Wrap(custom_err.ErrCacheDelFailed, err.Error())
	}
	return nil
}

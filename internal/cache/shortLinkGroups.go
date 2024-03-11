package cache

import (
	"SnapLink/internal/model"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"time"
)

import (
	"context"
)

// shortLinkGroup 更多是进行大量查询,因此使用 hash 缓存
const (
	// cache prefix key, must end with a colon
	shortLinkGroupsCachePrefixKey = "groups:"
	// ShortLinkGroupExpireTime expire time
	ShortLinkGroupExpireTime = 1 * time.Hour
	shortLinkGroupCacheEmpty = "empty"
)

var _ ShortLinkGroupCache = (*shortLinkGroupsCache)(nil)

// ShortLinkGroupCache cache interface
type ShortLinkGroupCache interface {
	HSet(ctx context.Context, username string, group *model.ShortLinkGroup) error
	HMSet(ctx context.Context, username string, groups []*model.ShortLinkGroup) error
	HSetEmpty(ctx context.Context, username string) error
	HGetALL(ctx context.Context, username string) ([]*model.ShortLinkGroup, error)
	HDel(ctx context.Context, username string, gids ...string) error
	UpdateOrders(ctx context.Context, username string, gids []string, orders []int) error
}

// shortLinkGroupsCache define a cache struct
type shortLinkGroupsCache struct {
	client *redis.Client
}

// NewShortLinkGroupCache new a cache
func NewShortLinkGroupCache(cacheType *model.CacheType) ShortLinkGroupCache {
	return &shortLinkGroupsCache{
		client: cacheType.Rdb,
	}
}

// makeSLGroupKey 获取用户的 group 缓存 key
func (c *shortLinkGroupsCache) makeSLGroupKey(username string) string {
	return shortLinkGroupsCachePrefixKey + username
}

// HSet 设置用户的 hash 缓存
func (c *shortLinkGroupsCache) HSet(ctx context.Context, username string, group *model.ShortLinkGroup) error {
	key := c.makeSLGroupKey(username)
	bytes, err := json.Marshal(group)
	if err != nil {
		return err
	}
	err = c.client.HSet(ctx, key, group.Gid, string(bytes)).Err()
	if err != nil {
		return err
	}
	return nil
}

// HMSet 设置用户的 hash 缓存
func (c *shortLinkGroupsCache) HMSet(ctx context.Context, username string, groups []*model.ShortLinkGroup) error {
	data := make(map[string]interface{})
	for _, group := range groups {
		bytes, err := json.Marshal(group)
		if err != nil {
			return err
		}
		data[group.Gid] = string(bytes)
	}
	key := c.makeSLGroupKey(username)
	err := c.client.HMSet(ctx, key, data).Err()
	return err
}

// HSetEmpty 设置用户的 hash 缓存为空
func (c *shortLinkGroupsCache) HSetEmpty(ctx context.Context, username string) error {
	key := c.makeSLGroupKey(username)
	return c.client.HSet(ctx, key, shortLinkGroupCacheEmpty, "").Err()
}

// HGetALL 获取用户的 hash 缓存
func (c *shortLinkGroupsCache) HGetALL(ctx context.Context, username string) ([]*model.ShortLinkGroup, error) {
	key := c.makeSLGroupKey(username)
	data, err := c.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	groups := make([]*model.ShortLinkGroup, 0, len(data))
	for k, v := range data {
		if k == shortLinkGroupCacheEmpty {
			// 当且仅当只存在一个空值时，返回空
			if len(data) == 1 {
				return nil, nil
			}
			continue
		}
		var group model.ShortLinkGroup
		err = json.Unmarshal([]byte(v), &group)
		if err != nil {
			return nil, err
		}
		groups = append(groups, &group)
	}
	return groups, nil
}

// HDel 删除用户的 hash 中的某个 key
func (c *shortLinkGroupsCache) HDel(ctx context.Context, username string, gids ...string) error {
	key := c.makeSLGroupKey(username)
	return c.client.HDel(ctx, key, gids...).Err()
}

// Del 删除用户的 hash 缓存
func (c *shortLinkGroupsCache) Del(ctx context.Context, username string) error {
	key := c.makeSLGroupKey(username)
	return c.client.Del(ctx, key).Err()
}

// UpdateOrders 更新用户的 hash 缓存中的排序
func (c *shortLinkGroupsCache) UpdateOrders(ctx context.Context, username string, gids []string, orders []int) error {
	pipeline := c.client.Pipeline()
	key := c.makeSLGroupKey(username)
	for i, gid := range gids {
		pipeline.HSet(ctx, key, gid, orders[i])
	}
	_, err := pipeline.Exec(ctx)
	return err
}

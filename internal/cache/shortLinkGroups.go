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
	ADD(ctx context.Context, username string, group *model.ShortLinkGroup) error
	MADD(ctx context.Context, username string, groups []*model.ShortLinkGroup) error
	SetEmpty(ctx context.Context, username string) error
	GetALL(ctx context.Context, username string) ([]*model.ShortLinkGroup, error)
	Del(ctx context.Context, username string) error
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
func makeSLGroupKey(username string) string {
	return shortLinkGroupsCachePrefixKey + username
}

// ADD 添加用户的 group 缓存
func (c *shortLinkGroupsCache) ADD(ctx context.Context, username string, group *model.ShortLinkGroup) error {
	key := makeSLGroupKey(username)
	bytes, err := json.Marshal(group)
	if err != nil {
		return err
	}
	err = c.client.ZAdd(ctx, key, &redis.Z{
		Score:  float64(group.SortOrder),
		Member: string(bytes),
	}).Err()
	if err != nil {
		return err
	}
	return nil
}

// MADD 批量添加用户的 group 缓存
func (c *shortLinkGroupsCache) MADD(ctx context.Context, username string, groups []*model.ShortLinkGroup) error {
	key := makeSLGroupKey(username)
	pipeline := c.client.Pipeline()
	for _, group := range groups {
		bytes, err := json.Marshal(group)
		if err != nil {
			return err
		}
		pipeline.ZAdd(ctx, key, &redis.Z{
			Score:  float64(group.SortOrder),
			Member: bytes,
		})
	}
	_, err := pipeline.Exec(ctx)
	return err
}

// SetEmpty 设置用户的 hash 缓存为空
func (c *shortLinkGroupsCache) SetEmpty(ctx context.Context, username string) error {
	key := makeSLGroupKey(username)
	return c.client.ZAddXX(ctx, key, &redis.Z{
		Score:  0,
		Member: shortLinkGroupCacheEmpty,
	}).Err()
}

// GetALL 获取用户的 hash 缓存
func (c *shortLinkGroupsCache) GetALL(ctx context.Context, username string) ([]*model.ShortLinkGroup, error) {
	key := makeSLGroupKey(username)
	data, err := c.client.ZRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, err
	}
	groups := make([]*model.ShortLinkGroup, 0, len(data))
	for _, v := range data {
		if v == shortLinkGroupCacheEmpty {
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

// Del 删除用户的 hash 缓存
func (c *shortLinkGroupsCache) Del(ctx context.Context, username string) error {
	key := makeSLGroupKey(username)
	return c.client.Del(ctx, key).Err()
}

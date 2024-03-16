package hyperloglog

import (
	"context"
	"github.com/redis/go-redis/v9"
)

const (
	// CachePrefixKey hyperloglog 缓存前缀
	CachePrefixKey = "hyperloglog:"
)

var _ Cache = (*hyperLogLogCache)(nil)

type Cache interface {
	// PFAdd 添加元素
	PFAdd(ctx context.Context, key string, values ...any) error
	// PFCount 获取基数
	PFCount(ctx context.Context, keys ...string) (int64, error)
	// PFMerge 合并多个hyperloglog
	PFMerge(ctx context.Context, destKey string, sourceKeys ...string) error
	// Delete 删除
	Delete(ctx context.Context, key string) error
}
type hyperLogLogCache struct {
	client *redis.Client
}

func NewCache(client *redis.Client) Cache {
	return &hyperLogLogCache{client: client}
}
func makePFKey(key string) string {
	return CachePrefixKey + key
}

// PFAdd 添加元素
func (h *hyperLogLogCache) PFAdd(ctx context.Context, key string, values ...any) error {
	return h.client.PFAdd(ctx, makePFKey(key), values...).Err()
}

// PFCount 获取基数
func (h *hyperLogLogCache) PFCount(ctx context.Context, keys ...string) (int64, error) {
	l := len(keys)
	for i := 0; i < l; i++ {
		keys[i] = makePFKey(keys[i])
	}
	return h.client.PFCount(ctx, keys...).Result()
}

// PFMerge 合并多个hyperloglog
func (h *hyperLogLogCache) PFMerge(ctx context.Context, destKey string, sourceKeys ...string) error {
	l := len(sourceKeys)
	for i := 0; i < l; i++ {
		sourceKeys[i] = makePFKey(sourceKeys[i])
	}
	return h.client.PFMerge(ctx, makePFKey(destKey), sourceKeys...).Err()
}

// Delete 删除
func (h *hyperLogLogCache) Delete(ctx context.Context, key string) error {
	return h.client.Del(ctx, makePFKey(key)).Err()
}

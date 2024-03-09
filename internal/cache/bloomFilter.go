package cache

import (
	"context"
	"github.com/redis/go-redis/v9"
)

const (
	// cache prefix key, must end with a colon
	BloomFilterCachePrefixKey = "BloomFilterCache:"
)

var _ BloomFilterCache = (*bloomFilterCache)(nil)

type BloomFilterCache interface {
	Create(ctx context.Context, key string, errorRate float64, capacity int) error
	Add(ctx context.Context, key string, value string) error
	MAdd(ctx context.Context, key string, values ...string) error

	Exists(ctx context.Context, key string, value string) (bool, error)
	MExists(ctx context.Context, key string, values ...string) ([]bool, error)

	Delete(ctx context.Context, key string) error
	Rename(ctx context.Context, key string, newKey string) error
}
type bloomFilterCache struct {
	client *redis.Client
}

func NewBloomFilterCache(client *redis.Client) BloomFilterCache {
	return &bloomFilterCache{client: client}
}

// Create 创建布隆过滤器
// capacity: 期望插入的元素数量
// errorRate: 期望的错误率
// expansion: 扩容倍数
// nonScaling: 是否禁止自动扩容
func (b *bloomFilterCache) Create(ctx context.Context, key string, errorRate float64, capacity int) error {
	return b.client.BFReserveWithArgs(ctx, makeBFKey(key), &redis.BFReserveOptions{
		Capacity:   int64(capacity),
		Error:      errorRate,
		Expansion:  2,
		NonScaling: false,
	}).Err()
}

// Add 添加值
func (b *bloomFilterCache) Add(ctx context.Context, key string, value string) error {
	return b.client.BFAdd(ctx, makeBFKey(key), value).Err()
}

// Exists 检查值是否存在
func (b *bloomFilterCache) Exists(ctx context.Context, key string, value string) (bool, error) {
	return b.client.BFExists(ctx, makeBFKey(key), value).Result()
}

// MAdd 批量添加多个值
func (b *bloomFilterCache) MAdd(ctx context.Context, key string, values ...string) error {
	pipeline := b.client.Pipeline()
	for _, value := range values {
		pipeline.BFAdd(ctx, makeBFKey(key), value)
	}
	_, err := pipeline.Exec(ctx)
	return err

}

// MExists 批量检查多个值是否存在
func (b *bloomFilterCache) MExists(ctx context.Context, key string, values ...string) ([]bool, error) {
	pipeline := b.client.Pipeline()
	for _, value := range values {
		pipeline.BFExists(ctx, makeBFKey(key), value)
	}
	cmds, err := pipeline.Exec(ctx)
	if err != nil {
		return nil, err
	}
	results := make([]bool, len(cmds))
	for i, cmd := range cmds {
		result, err := cmd.(*redis.BoolCmd).Result()
		if err != nil {
			return nil, err
		}
		results[i] = result
	}
	return results, nil
}

// Delete 删除布隆过滤器
func (b *bloomFilterCache) Delete(ctx context.Context, key string) error {
	return b.client.Del(ctx, makeBFKey(key)).Err()
}

// Rename 重命名
func (b *bloomFilterCache) Rename(ctx context.Context, key string, newKey string) error {
	return b.client.Rename(ctx, makeBFKey(key), makeBFKey(newKey)).Err()
}

func makeBFKey(key string) string {
	return BloomFilterCachePrefixKey + key
}

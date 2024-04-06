package cache

import (
	"context"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

type BFCache interface {
	Create(ctx context.Context, key string, errorRate float64, capacity int) error
	Add(ctx context.Context, key string, value string) error
	MAdd(ctx context.Context, key string, values ...string) error

	Exists(ctx context.Context, key string, value string) (bool, error)
	MExists(ctx context.Context, key string, values ...string) ([]bool, error)

	Delete(ctx context.Context, key string) error
	Rename(ctx context.Context, key string, newKey string) error
	Info(ctx context.Context, key string) (items, capacity int64, err error)
}
type bloomFilterCache struct {
	client *redis.Client
	keyGen KeyGenerator
}

func NewBloomFilterCache(client *redis.Client, keyGen KeyGenerator) (BFCache, error) {
	cache := &bloomFilterCache{
		client: client,
		keyGen: keyGen,
	}
	return cache, nil
}

// Create 创建布隆过滤器
// capacity: 期望插入的元素数量
// errorRate: 期望的错误率
// expansion: 扩容倍数
// nonScaling: 是否禁止自动扩容
func (b *bloomFilterCache) Create(ctx context.Context, key string, errorRate float64, capacity int) error {
	key = b.keyGen(key)
	if err := b.client.BFReserveWithArgs(ctx, key, &redis.BFReserveOptions{
		Capacity:   int64(capacity),
		Error:      errorRate,
		Expansion:  2,
		NonScaling: false,
	}).Err(); err != nil {
		return errors.Wrap(ErrBFCacheCreateFailed, err.Error())
	}
	return nil
}

// Add 添加值
func (b *bloomFilterCache) Add(ctx context.Context, key string, value string) error {
	key = b.keyGen(key)
	if err := b.client.BFAdd(ctx, key, value).Err(); err != nil {
		return errors.Wrap(ErrBFCacheAddFailed, err.Error())
	}
	return nil
}

// Exists 检查值是否存在
func (b *bloomFilterCache) Exists(ctx context.Context, key string, value string) (bool, error) {
	key = b.keyGen(key)
	exist, err := b.client.BFExists(ctx, key, value).Result()
	if err != nil {
		return false, errors.Wrap(ErrBFCacheExistsFailed, err.Error())
	}
	return exist, nil
}

// MAdd 批量添加多个值
func (b *bloomFilterCache) MAdd(ctx context.Context, key string, values ...string) error {
	key = b.keyGen(key)

	pipeline := b.client.Pipeline()
	for _, value := range values {
		pipeline.BFAdd(ctx, key, value)
	}
	_, err := pipeline.Exec(ctx)
	if err != nil {
		return errors.Wrap(ErrBFCacheMAddFailed, err.Error())
	}
	return nil

}

// MExists 批量检查多个值是否存在
func (b *bloomFilterCache) MExists(ctx context.Context, key string, values ...string) ([]bool, error) {
	key = b.keyGen(key)

	pipeline := b.client.Pipeline()
	for _, value := range values {
		pipeline.BFExists(ctx, key, value)
	}
	cmds, err := pipeline.Exec(ctx)
	if err != nil {
		return nil, errors.Wrap(ErrBFCacheMExistsFailed, err.Error())
	}
	results := make([]bool, len(cmds))
	for i, cmd := range cmds {
		result, err := cmd.(*redis.BoolCmd).Result()
		if err != nil {
			return nil, errors.Wrap(ErrBFCacheMExistsFailed, err.Error())
		}
		results[i] = result
	}
	return results, nil
}

// Delete 删除布隆过滤器
func (b *bloomFilterCache) Delete(ctx context.Context, key string) error {
	key = b.keyGen(key)
	if err := b.client.Del(ctx, key).Err(); err != nil {
		return errors.Wrap(ErrBFCacheDelFailed, err.Error())
	}
	return nil
}

// Rename 重命名
func (b *bloomFilterCache) Rename(ctx context.Context, key string, newKey string) error {
	key = b.keyGen(key)
	newKey = b.keyGen(newKey)
	if err := b.client.Rename(ctx, key, newKey).Err(); err != nil {
		return errors.Wrap(ErrBFCacheRenameFailed, err.Error())
	}
	return nil
}

// Info 获取布隆过滤器的信息
// items: 已插入元素数量
// capacity: 布隆过滤器容量

func (b *bloomFilterCache) Info(ctx context.Context, key string) (items, capacity int64, err error) {
	key = b.keyGen(key)

	info, err := b.client.BFInfo(ctx, key).Result()
	if err != nil {
		return 0, 0, errors.Wrap(ErrBFCacheInfoFailed, err.Error())
	}
	return info.ItemsInserted, info.Capacity, nil
}

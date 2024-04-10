package cache

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"golang.org/x/sync/singleflight"
	"time"
)

const EmptyValue = "kvCache empty value"
const defaultTTL = time.Minute * 10

func makeRandTTL(ttl time.Duration) time.Duration {
	if ttl <= 0 {
		return 0
	}
	return ttl + time.Duration(time.Now().UnixNano()%int64(ttl))
}

// IKVCache 多级 KV 缓存接口
type IKVCache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	SetCacheWithNotFound(ctx context.Context, key string, ttl time.Duration) error
	Del(ctx context.Context, key string) error
}

// kvCache define a cache struct
type kvCache struct {
	client     *redis.Client
	sfg        *singleflight.Group
	keyGen     KeyGenerator
	localCache ILocalCache
}

// NewKVCache new a KVCache
func NewKVCache(client *redis.Client, keyGen KeyGenerator, localCache ILocalCache) (IKVCache, error) {
	cache := &kvCache{
		client:     client,
		sfg:        new(singleflight.Group),
		keyGen:     keyGen,
		localCache: localCache,
	}
	return cache, nil
}

// Get 获取分组下的短链接数量
func (c *kvCache) Get(ctx context.Context, key string) (string, error) {
	key = c.keyGen(key)
	if c.localCache != nil {
		if value, exist := c.localCache.Get(key); exist {
			return value.(string), nil
		}
	}
	value, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", ErrKVCacheNotFound
		}
		return "", err
	}
	if c.localCache != nil {
		// 将数据写入本地缓存
		ttl := makeRandTTL(defaultTTL)
		if !c.localCache.SetWithTTL(key, value, 2, ttl) {
			return "", errors.Wrap(ErrKVCacheSetLocalCacheFailed, fmt.Sprintf("key: %s, value: %s", key, value))
		}
	}
	return value, nil
}

// Set 设置分组下的短链接数量
// 此处均使用 errors.Wrap 进行封装，以快速定位与分辨错误
func (c *kvCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	key = c.keyGen(key)
	//做统一处理,以适配 redis 的 ttl 规则
	ttl = makeRandTTL(ttl)
	if c.localCache != nil {
		// 如果 ttl 小于 0，视为永不过期
		if ttl == 0 {
			c.localCache.Set(key, value, 3) // 永不过期的值应该是权重较高的
		} else {
			if !c.localCache.SetWithTTL(key, value, 2, ttl) {
				return errors.Wrap(ErrKVCacheSetLocalCacheFailed, fmt.Sprintf("key: %s, value: %s", key, value))
			}
		}
	}
	if err := c.client.Set(ctx, key, value, ttl).Err(); err != nil {
		return errors.Wrap(ErrKVCacheSetRedisFailed, err.Error())
	}
	return nil
}

// SetCacheWithNotFound 空值防御机制
func (c *kvCache) SetCacheWithNotFound(ctx context.Context, key string, ttl time.Duration) error {
	ttl = makeRandTTL(ttl)
	return c.Set(ctx, key, EmptyValue, ttl)
}

// Del 删除数据键值
func (c *kvCache) Del(ctx context.Context, key string) error {
	key = c.keyGen(key)
	_, err, shared := c.sfg.Do(key, func() (interface{}, error) {
		c.localCache.Del(key)
		return nil, c.client.Del(ctx, key).Err()
	})
	if err != nil {
		return errors.Wrap(ErrKVCacheDelFailed, err.Error())
	}
	// 如果是共享删除的情况，为了避免时间间隔内进行存在数据获取延迟，进行二次删除
	if shared {
		// 此处再次使用 singleflight 来进行合并
		_, err, _ = c.sfg.Do(key, func() (interface{}, error) {
			return nil, c.client.Del(ctx, key).Err()
		})
		if err != nil {
			c.localCache.Del(key)
			return errors.Wrap(ErrKVCacheDelFailed, err.Error())
		}
	}
	return nil
}

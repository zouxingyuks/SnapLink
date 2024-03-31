package go_redis_bloomfilter

import (
	"context"
	"github.com/go-redis/redis/v8"
)

type BloomFilter struct {
	client *redis.Client
}

func NewBloomFilter(client *redis.Client) *BloomFilter {
	return &BloomFilter{client}
}

// ADD 添加元素
// err 错误信息
func (b *BloomFilter) ADD(ctx context.Context, key string, value string) (err error) {
	count, err := b.client.Do(ctx, "BF.ADD", key, value).Result()
	if err != nil {
		return err
	}
	//如果元素已经存在，则返回错误
	if count.(int64) == 0 {
		return ErrElementExists

	}
	return nil
}

// MADD 添加多个元素
// nums 返回添加的元素个数,如果元素已经存在则返回0
// err 错误信息
func (b *BloomFilter) MADD(ctx context.Context, key string, values ...string) (nums []int64, err error) {
	args := make([]any, len(values)+2)
	args[0] = "BF.MADD"
	args[1] = key
	for i, v := range values {
		args[i+2] = v
	}
	counts, err := b.client.Do(ctx, args...).Result()
	results := []int64{}
	l := len(counts.([]interface{}))
	for i := 0; i < l; i++ {
		results = append(results, counts.([]interface{})[i].(int64))
	}
	return results, err
}

// EXIST 判断元素是否存在
// exist 返回元素是否存在
// err 错误信息
func (b *BloomFilter) EXIST(ctx context.Context, key string, value string) (exist bool, err error) {
	count, err := b.client.Do(ctx, "BF.EXISTS", key, value).Result()
	if count.(int64) == 1 {
		exist = true
	}
	return exist, err
}

// MEXIST 判断多个元素是否存在
func (b *BloomFilter) MEXIST(ctx context.Context, key string, values ...string) (results []int64, err error) {
	args := make([]any, len(values)+2)
	args[0] = "BF.MEXISTS"
	args[1] = key
	counts, err := b.client.Do(ctx, args...).Result()
	l := len(counts.([]interface{}))
	for i := 0; i < l; i++ {
		results = append(results, counts.([]interface{})[i].(int64))
	}
	return results, err
}

// EXISTOrADD 判断元素是否存在，如果不存在则添加
// exist 返回元素是否存在
// err 错误信息
func (b *BloomFilter) EXISTOrADD(ctx context.Context, key string, value string) (exist bool, err error) {
	exist, err = b.EXIST(ctx, key, value)
	//如果不存在,且此错误非预定义的错误，则返回错误
	if err != nil {
		return false, err
	}
	if !exist {
		err = b.ADD(ctx, key, value)
		return false, err
	}
	return true, nil

}

// RESERVE 创建一个布隆过滤器
// message 返回创建的布隆过滤器的信息
func (b *BloomFilter) RESERVE(ctx context.Context, key string, errorRate float64, capacity int) error {
	err := b.client.Do(ctx, "BF.RESERVE", key, errorRate, capacity).Err()
	return err
}

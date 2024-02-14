package cache

import (
	"SnapLink/internal/model"
	go_redis_bloomfilter "SnapLink/pkg/go-redis-bloomfilter"
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"io"
)

type LinkStatsCache struct {
	client      *redis.Client
	bloomFilter *go_redis_bloomfilter.BloomFilter
}

func NewLinkStatsCache(cacheType *model.CacheType) *LinkStatsCache {
	return &LinkStatsCache{
		client:      cacheType.Rdb,
		bloomFilter: go_redis_bloomfilter.NewBloomFilter(cacheType.Rdb),
	}
}

// Set 设置缓存
func (l *LinkStatsCache) Set(ctx context.Context, values map[string]any) error {
	for key, value := range values {
		_, err := l.client.HSet(ctx, key, value).Result()
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("redis hset error, key: %s, value: %v", key, value))
		}
	}
	return nil
}

func (l *LinkStatsCache) Get(ctx context.Context, originalUrl string, date string, hour int) (*model.LinkAccessStat, error) {
	m, err := l.client.HGetAll(ctx, makeKey(originalUrl, date, hour)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			l.client.Set(ctx, makeKey(originalUrl, date, hour), 0, 0)
		}
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, errors.Wrap(MarshalTypeError, err.Error())
	}
	var linkAccessStat model.LinkAccessStat
	err = json.Unmarshal(bytes, &linkAccessStat)
	if err != nil {
		return nil, errors.Wrap(UnmarshalTypeError, err.Error())
	}
	return &linkAccessStat, nil
}

// GetByDateHour 根据日期和小时获取所有的数据，用于进行持久化和统计
func (l *LinkStatsCache) GetByDateHour(ctx context.Context, date string, hour int) ([]*model.LinkAccessStat, error) {
	//获取所有相同日期和小时的数据
	keys, err := l.client.SMembers(ctx, makeSetKey(date, hour)).Result()
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("redis smembers error, key: %s", makeSetKey(date, hour)))
	}
	datas := make([]*model.LinkAccessStat, 0)
	for _, key := range keys {
		//遍历所有数据
		m, err := l.client.HGetAll(ctx, key).Result()
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("redis hgetall error, key: %s", key))
		}
		bytes, err := json.Marshal(m)
		if err != nil {
			return nil, errors.Wrap(MarshalTypeError, err.Error())
		}
		var linkAccessStat model.LinkAccessStat
		err = json.Unmarshal(bytes, &linkAccessStat)
		if err != nil {
			return nil, errors.Wrap(UnmarshalTypeError, err.Error())
		}
		datas = append(datas, &linkAccessStat)
	}
	return datas, nil

}

// ExistOrAddIP 判断IP是否存在，如果不存在则添加
// exist: 是否存在
// err: 错误。如果存在或是添加成功则返回nil
func (l *LinkStatsCache) existOrAddIP(ctx context.Context, originalUrl string, date string, hour int, ip string) (bool, error) {
	bKey := makeBloomFilterKey(originalUrl, date, hour, "ip")
	exist, err := l.bloomFilter.EXISTOrADD(ctx, bKey, ip)
	if err != nil {
		return false, err
	}
	return exist, nil
}

// ExistOrAddLocation 判断Location是否存在，如果不存在则添加
func (l *LinkStatsCache) existOrAddLocation(ctx context.Context, originalUrl string, date string, hour int, location string) (bool, error) {
	bKey := makeBloomFilterKey(originalUrl, date, hour, "location")

	exist, err := l.bloomFilter.EXISTOrADD(ctx, bKey, location)
	//本身出现不可预料的错误
	if err != nil {
		return false, err
	}
	return exist, nil
}

// ExistOrAddUA  判断UA是否存在，如果不存在则添加
func (l *LinkStatsCache) existOrAddUA(ctx context.Context, originalUrl string, date string, hour int, ua string) (bool, error) {
	bKey := makeBloomFilterKey(originalUrl, date, hour, "ua")
	exist, err := l.bloomFilter.EXISTOrADD(ctx, bKey, ua)
	if err != nil {
		return false, err
	}
	return exist, nil
}

// UpdatePv 更新PV
func (l *LinkStatsCache) UpdatePv(ctx context.Context, originalUrl string, date string, hour int) error {
	_, err := l.client.HIncrBy(ctx, makeRecordKey(originalUrl, date, hour), "pv", 1).Result()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("redis hincrby error, key: %s", makeKey(originalUrl, date, hour)))
	}
	return nil
}

// UpdateUv 更新UV
// 更新UV
func (l *LinkStatsCache) UpdateUv(ctx context.Context, originalUrl string, date string, hour int) error {
	_, err := l.client.HIncrBy(ctx, makeRecordKey(originalUrl, date, hour), "uv", 1).Result()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("redis hincrby error, key: %s", makeKey(originalUrl, date, hour)))
	}
	return nil
}

// UpdateUip 更新Uip
func (l *LinkStatsCache) UpdateUip(ctx context.Context, originalUrl string, date string, hour int, uip string) error {

	exist, err := l.existOrAddIP(ctx, originalUrl, date, hour, uip)
	//本身出现不可预料的错误
	if err != nil {
		return err
	}
	//如果存在则不更新
	if exist {
		return nil
	}
	_, err = l.client.HIncrBy(ctx, makeRecordKey(originalUrl, date, hour), "uip", 1).Result()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("redis hincrby error, key: %s", makeKey(originalUrl, date, hour)))
	}
	return nil

}

// makeKey 生成用于统计表的键
// 参数 originalUrl 是原始的 URL，
// date 是日期（格式：yyyyMMdd），
// hour 是小时（24小时制）。
// 返回值是格式化后的键。
func makeKey(originalUrl string, date string, hour int) string {
	urlHash := ""
	if originalUrl != "" {
		// 对原始URL进行SHA1哈希处理，以保证键的长度固定
		hash := sha1.New()
		io.WriteString(hash, originalUrl)
		urlHash = fmt.Sprintf("%x", hash.Sum(nil))
	}
	// 使用|作为不同部分的分隔符，生成键
	key := fmt.Sprintf("%s:%02d:%s", date, hour, urlHash)
	return key
}

// makeSetKey 生成用于集合的键
// 参数 date 是日期（格式：yyyyMMdd），hour 是小时（24小时制）。
// 返回值是格式化后的键。
func makeSetKey(date string, hour int) string {
	return fmt.Sprintf("%s:%02d", date, hour)
}

// makeBloomFilterKey 生成用于布隆过滤器的键
// 参数 originalUrl 是原始的 URL，date 是日期（格式：yyyyMMdd），hour 是小时（24小时制）。
// 返回值是格式化后的键。
func makeBloomFilterKey(originalUrl string, date string, hour int, name string) string {
	return fmt.Sprintf("%s:bloomFilter:%s", makeKey(originalUrl, date, hour), name)
}

// makeRecordKey 生成用于记录的键
// 参数 originalUrl 是原始的 URL，date 是日期（格式：yyyyMMdd），hour 是小时（24小时制）。
// 返回值是格式化后的键。
func makeRecordKey(originalUrl string, date string, hour int) string {
	return fmt.Sprintf("%s:record", makeKey(originalUrl, date, hour))
}

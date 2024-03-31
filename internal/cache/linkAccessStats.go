package cache

import (
	"SnapLink/internal/model"
	"SnapLink/pkg/go-redis-bloomfilter"
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"io"
	"strconv"
	"time"
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

// GetAllUri 获取所有的uri
func (l *LinkStatsCache) GetAllUri(ctx context.Context, date string, hour int) ([]string, error) {
	keys, err := l.client.SMembers(ctx, fmt.Sprintf("%s:%02d:uris", date, hour)).Result()
	if err != nil {
		//todo 优化此处的错误处理
		return nil, errors.Wrap(err, fmt.Sprintf("redis smembers error, key: %s", fmt.Sprintf("%s:%02d:uris", date, hour)))
	}
	return keys, nil
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

// UpdatePv 更新PV
func (l *LinkStatsCache) UpdatePv(ctx context.Context, uri string, date string, hour int) error {
	setsKey := fmt.Sprintf("%s:%02d:uris", date, hour)
	isNew := false
	if l.client.Exists(ctx, setsKey).Val() == 0 {
		isNew = true
	}
	if err := l.client.SAdd(ctx, setsKey, uri).Err(); err != nil {
		//todo 优化此处的错误处理
		return errors.Wrap(err, fmt.Sprintf("redis set error, key: %s", fmt.Sprintf("%s:%02d:uris", date, hour)))
	}
	//使用集合来进行统计目前的uris
	if isNew {
		//设置过期时间
		if err := l.client.ExpireAt(ctx, setsKey, expireAt(date, hour)).Err(); err != nil {
			//todo 优化此处的错误处理
			return errors.Wrap(err, fmt.Sprintf("redis expire error, key: %s", fmt.Sprintf("%s:%02d:uris", date, hour)))
		}
	}
	//开始记录统计信息
	staticKey := makeStaticKey(uri, date, hour)
	if l.client.Exists(ctx, staticKey).Val() == 0 {
		//设置过期时间
		isNew = true
	}
	_, err := l.client.HIncrBy(ctx, staticKey, "pv", 1).Result()

	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("redis hincrby error, key: %s", makeKey(uri, date, hour)))
	}
	if isNew {
		l.client.ExpireAt(ctx, staticKey, expireAt(date, hour))

	}
	return nil
}

// UpdateUv 更新UV
// 更新UV
func (l *LinkStatsCache) UpdateUv(ctx context.Context, uri string, date string, hour int) error {
	_, err := l.client.HIncrBy(ctx, makeStaticKey(uri, date, hour), "uv", 1).Result()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("redis hincrby error, key: %s", makeKey(uri, date, hour)))
	}
	return nil
}

// UpdateIp 更新Uip
func (l *LinkStatsCache) UpdateIp(ctx context.Context, uri string, date string, hour int, ip string) error {
	//基于布隆过滤器的去重
	bKey := makeBloomFilterKey(uri, date, hour, "ip")
	exist, err := l.bloomFilter.EXISTOrADD(ctx, bKey, ip)
	if err != nil {
		return err
	}
	//本身出现不可预料的错误
	if err != nil {
		return err
	}
	//如果存在则不更新
	if exist {
		return nil
	}
	_, err = l.client.HIncrBy(ctx, makeStaticKey(uri, date, hour), "uip", 1).Result()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("redis hincrby error, key: %s", makeKey(uri, date, hour)))
	}
	return nil

}

// UpdateLocation 更新Location
// 统计地理位置
func (l *LinkStatsCache) UpdateLocation(ctx context.Context, uri string, date string, hour int, cityCode string) error {
	isNew := false
	key := makeHashKey(uri, date, hour, "locations")
	//存在性判断
	if l.client.Exists(ctx, key).Val() == 0 {
		isNew = true
	}
	_, err := l.client.HIncrBy(ctx, key, cityCode, 1).Result()
	if isNew {
		l.client.ExpireAt(ctx, key, expireAt(date, hour))

	}
	return err
}

// UpdateUA 更新UA
func (l *LinkStatsCache) UpdateUA(ctx context.Context, uri string, date string, hour int, browser, device string) error {
	isNew := false
	browserKey := makeHashKey(uri, date, hour, "browsers")
	deviceKey := makeHashKey(uri, date, hour, "devices")
	//存在性判断
	if l.client.Exists(ctx, browserKey).Val() == 0 {
		isNew = true
	}
	_, err := l.client.HIncrBy(ctx, browserKey, browser, 1).Result()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("redis hincrby error, key: %s", browserKey))

	}
	_, err = l.client.HIncrBy(ctx, deviceKey, device, 1).Result()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("redis hincrby error, key: %s", deviceKey))

	}
	if isNew {
		l.client.ExpireAt(ctx, browserKey, expireAt(date, hour))
		l.client.ExpireAt(ctx, deviceKey, expireAt(date, hour))
	}
	return nil
}

// GetStatisticByDateHour 从缓存中获取统计数据
func (l *LinkStatsCache) GetStatisticByDateHour(ctx context.Context, uri string, date string, hour int) (*model.LinkAccessStatistic, error) {
	var err error
	static := new(model.LinkAccessStatistic)
	static.Date = date
	static.Hour = hour
	static.URI = uri
	data := l.client.HGetAll(ctx, makeStaticKey(uri, date, hour)).Val()
	if len(data) == 0 {
		//todo 设计错误
		return nil, errors.New("no data")
	}
	static.Pv, _ = strconv.ParseInt(data["pv"], 10, 64)
	static.Uv, _ = strconv.ParseInt(data["uv"], 10, 64)
	static.Uip, _ = strconv.ParseInt(data["uip"], 10, 64)
	static.Datetime = fmt.Sprintf("%s %02d:00:00", date, hour)
	//获取地理位置
	locations := l.client.HGetAll(ctx, makeHashKey(uri, date, hour, "locations")).Val()
	static.Regions, err = json.Marshal(locations)
	if err != nil {
		return nil, err
	}
	ips := l.client.HGetAll(ctx, makeBloomFilterKey(uri, date, hour, "ips")).Val()
	static.IPs, err = json.Marshal(ips)
	if err != nil {
		return nil, err
	}
	devices := l.client.HGetAll(ctx, makeBloomFilterKey(uri, date, hour, "devices")).Val()
	static.Devices, err = json.Marshal(devices)
	if err != nil {
		return nil, err
	}

	return static, nil
}

// makeKey 生成用于统计表的键
// 参数 uri 是原始的 URL，
// date 是日期（格式：yyyyMMdd），
// hour 是小时（24小时制）。
// 返回值是格式化后的键。
func makeKey(uri string, date string, hour int) string {
	urlHash := ""
	if uri != "" {
		// 对原始URL进行SHA1哈希处理，以保证键的长度固定
		hash := sha1.New()
		io.WriteString(hash, uri)
		urlHash = fmt.Sprintf("%x", hash.Sum(nil))
	}
	// 使用|作为不同部分的分隔符，生成键
	key := fmt.Sprintf("%s:%02d:%s", date, hour, urlHash)
	return key
}

// makeSetKey 生成用于集合的键
// 参数 date 是日期（格式：yyyyMMdd），hour 是小时（24小时制）。
// 返回值是格式化后的键。
func makeHashKey(uri string, date string, hour int, name string) string {
	return fmt.Sprintf("%s:hash:%s", makeKey(uri, date, hour), name)
}

// makeBloomFilterKey 生成用于布隆过滤器的键
// 参数 uri 是原始的 URL，date 是日期（格式：yyyyMMdd），hour 是小时（24小时制）。
// 返回值是格式化后的键。
func makeBloomFilterKey(uri string, date string, hour int, name string) string {
	return fmt.Sprintf("%s:bloomFilterCache:%s", makeKey(uri, date, hour), name)
}

// makeStaticKey 生成用于记录的键
// 参数 uri 是原始的 URL，date 是日期（格式：yyyyMMdd），hour 是小时（24小时制）。
// 返回值是格式化后的键。
func makeStaticKey(uri string, date string, hour int) string {
	return fmt.Sprintf("%s:static", makeKey(uri, date, hour))
}

func expireAt(date string, hour int) time.Time {
	//下一个小时的10分钟
	expireAtTime, _ := time.Parse("2006-01-02 15:04:05", fmt.Sprintf("%s %02d:10:00", date, hour))
	expireAtTime = expireAtTime.Add(1 * time.Hour)
	return expireAtTime
}

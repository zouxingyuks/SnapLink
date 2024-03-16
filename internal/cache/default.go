package cache

import (
	"SnapLink/internal/config"
	"context"
	"github.com/redis/go-redis/v9"
	"sync"
	"time"
)

type Cache struct {
	client *redis.Client
}

func NewCache(client *redis.Client) *Cache {
	return &Cache{
		client: client,
	}
}

var defaultInstance struct {
	once  sync.Once
	cache *Cache
}

func Instance() *Cache {
	defaultInstance.once.Do(func() {
		//todo 修改此处
		opt := config.Get().BFRedis
		defaultInstance.cache = NewCache(redis.NewClient(&redis.Options{
			Addr:         opt.Addr,
			Password:     opt.Password,
			DB:           opt.DB,
			ReadTimeout:  time.Duration(opt.ReadTimeout) * time.Second,
			WriteTimeout: time.Duration(opt.WriteTimeout) * time.Second,
			DialTimeout:  time.Duration(opt.DialTimeout) * time.Second,
		}))
	})
	return defaultInstance.cache
}

// Incr
func Incr(ctx context.Context, key string, value int64) error {
	return Instance().client.IncrBy(ctx, key, value).Err()
}

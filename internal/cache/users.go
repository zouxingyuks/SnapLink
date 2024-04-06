package cache

import (
	"strings"
	"time"

	"SnapLink/internal/model"

	"github.com/zhufuyi/sponge/pkg/cache"
	"github.com/zhufuyi/sponge/pkg/encoding"
)

const (
	// bfCache prefix key, must end with a colon
	tUserCachePrefixKey = "tUser:"
	// TUserExpireTime expire time
	TUserExpireTime = 5 * time.Minute
)

var _ TUserCache = (*tUserCache)(nil)

// TUserCache bfCache interface
type TUserCache interface {
	//Set(ctx context.Context, id uint64, data *model.TUser, duration time.Duration) error
	//Get(ctx context.Context, id uint64) (*model.TUser, error)
	//MultiGet(ctx context.Context, ids []uint64) (map[uint64]*model.TUser, error)
	//MultiSet(ctx context.Context, data []*model.TUser, duration time.Duration) error
	//Del(ctx context.Context, id uint64) error
	//SetCacheWithNotFound(ctx context.Context, id uint64) error
}

// tUserCache define a cache struct
type tUserCache struct {
	cache cache.Cache
}

// NewTUserCache new a bfCache
func NewTUserCache(cacheType *model.CacheType) TUserCache {
	jsonEncoding := encoding.JSONEncoding{}
	cachePrefix := ""

	cType := strings.ToLower(cacheType.CType)
	switch cType {
	case "redis":
		c := cache.NewRedisCache(cacheType.Rdb, cachePrefix, jsonEncoding, func() interface{} {
			return &model.TUser{}
		})
		return &tUserCache{cache: c}
	case "memory":
		c := cache.NewMemoryCache(cachePrefix, jsonEncoding, func() interface{} {
			return &model.TUser{}
		})
		return &tUserCache{cache: c}
	}

	return nil // no bfCache
}

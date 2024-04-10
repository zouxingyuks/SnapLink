package cache

import (
	"time"

	"github.com/zhufuyi/sponge/pkg/cache"
)

const (
	// bfCache prefix key, must end with a colon
	TUserCachePrefixKey = "tUser"
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

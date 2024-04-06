package custom_err

import (
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

// 规则 ErrCache+Action+Result
// 返回的错误必须经过封装,以快速定位错误类型与错误位置
var (
	// ErrCacheNotFound No hit cache
	ErrCacheNotFound   = redis.Nil
	ErrCacheInitFailed = errors.New("init cache failed")
	ErrCacheGetFailed  = errors.New("cache get value failed")
	ErrCacheSetFailed  = errors.New("cache set value failed")
	ErrCacheDelFailed  = errors.New("cache del value failed")
	ErrMarshalType     = errors.New("ErrMarshalType")
	ErrUnmarshalType   = errors.New("ErrUnmarshalType")
	ErrDelFailed       = errors.New("del data failed")
)

package custom_err

import (
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

var (
	// ErrCacheNotFound No hit cache
	ErrCacheNotFound = redis.Nil

	// ErrRecordNotFound no records found
	ErrRecordNotFound = gorm.ErrRecordNotFound
)

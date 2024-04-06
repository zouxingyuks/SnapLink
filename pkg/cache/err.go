package cache

import "github.com/pkg/errors"

var (
	ErrDelFailed           = errors.New("del data failed")
	ErrCacheNotFound       = errors.New("cache no found")
	ErrSetLocalCacheFailed = errors.New("set local cache failed")
	ErrSetRedisFailed      = errors.New("set redis failed")
)

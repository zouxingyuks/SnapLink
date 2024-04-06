package cache

import "github.com/pkg/errors"

// kvCache Err
var (
	ErrKVCacheDelFailed           = errors.New("del data failed")
	ErrKVCacheNotFound            = errors.New("cache no found")
	ErrKVCacheSetLocalCacheFailed = errors.New("set local cache failed")
	ErrKVCacheSetRedisFailed      = errors.New("set redis failed")
)

// 错误定义
var (
	ErrBFCacheCreateFailed  = errors.New("bloom filter cache creation failed")
	ErrBFCacheAddFailed     = errors.New("bloom filter cache add failed")
	ErrBFCacheExistsFailed  = errors.New("bloom filter cache existence check failed")
	ErrBFCacheMAddFailed    = errors.New("bloom filter cache multiple add failed")
	ErrBFCacheMExistsFailed = errors.New("bloom filter cache multiple existence check failed")
	ErrBFCacheDelFailed     = errors.New("bloom filter cache delete failed")
	ErrBFCacheRenameFailed  = errors.New("bloom filter cache rename failed")
	ErrBFCacheInfoFailed    = errors.New("bloom filter cache info retrieval failed")
)

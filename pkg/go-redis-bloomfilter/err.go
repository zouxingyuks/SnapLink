package go_redis_bloomfilter

import "github.com/pkg/errors"

var (
	//元素已存在
	ErrElementExists = errors.New("element exists")
)

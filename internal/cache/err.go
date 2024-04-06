package cache

import "github.com/pkg/errors"

var (
	ErrInitCacheFailed = errors.New("init cache failed")
	ErrMarshalType     = errors.New("ErrMarshalType")
	ErrUnmarshalType   = errors.New("ErrUnmarshalType")
)

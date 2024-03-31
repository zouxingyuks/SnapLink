package cache

import "github.com/pkg/errors"

var (
	MarshalTypeError   = errors.New("MarshalTypeError")
	UnmarshalTypeError = errors.New("UnmarshalTypeError")
)

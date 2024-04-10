package custom_err

import "github.com/pkg/errors"

var (
	ErrMarshalType   = errors.New("ErrMarshalType")
	ErrUnmarshalType = errors.New("ErrUnmarshalType")
	ErrDelFailed     = errors.New("del data failed")
)

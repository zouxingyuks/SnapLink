// Package ecode is the package that unifies the definition of http error codes or grpc error codes here.
package ecode

import (
	"github.com/zhufuyi/sponge/pkg/errcode"
)

// http system level error code, error code range 10000~20000
var (
	InvalidParams       = errcode.InvalidParams
	Unauthorized        = errcode.Unauthorized
	InternalServerError = errcode.InternalServerError
	NotFound            = errcode.NotFound
)

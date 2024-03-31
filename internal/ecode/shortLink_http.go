package ecode

import (
	"github.com/zhufuyi/sponge/pkg/errcode"
)

// shortLink business-level http error codes.
// the shortLinkNO value range is 1~100, if the same number appears, it will cause a failure to start the service.
var (
	shortLinkNO       = 85
	shortLinkName     = "shortLink"
	shortLinkBaseCode = errcode.HCode(shortLinkNO)

	ErrUpdateByIDShortLink   = errcode.NewError(shortLinkBaseCode+4, "failed to update "+shortLinkName)
	ErrGetByIDShortLink      = errcode.NewError(shortLinkBaseCode+5, "failed to get "+shortLinkName+" details")
	ErrListByLastIDShortLink = errcode.NewError(shortLinkBaseCode+8, "failed to list by last id "+shortLinkName)
	ErrListShortLink         = errcode.NewError(shortLinkBaseCode+9, "failed to list of "+shortLinkName)
)

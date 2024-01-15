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

	ErrCreateShortLink         = errcode.NewError(shortLinkBaseCode+1, "failed to create "+shortLinkName)
	ErrDeleteByIDShortLink     = errcode.NewError(shortLinkBaseCode+2, "failed to delete "+shortLinkName)
	ErrDeleteByIDsShortLink    = errcode.NewError(shortLinkBaseCode+3, "failed to delete by batch ids "+shortLinkName)
	ErrUpdateByIDShortLink     = errcode.NewError(shortLinkBaseCode+4, "failed to update "+shortLinkName)
	ErrGetByIDShortLink        = errcode.NewError(shortLinkBaseCode+5, "failed to get "+shortLinkName+" details")
	ErrGetByConditionShortLink = errcode.NewError(shortLinkBaseCode+6, "failed to get "+shortLinkName+" details by conditions")
	ErrListByIDsShortLink      = errcode.NewError(shortLinkBaseCode+7, "failed to list by batch ids "+shortLinkName)
	ErrListByLastIDShortLink   = errcode.NewError(shortLinkBaseCode+8, "failed to list by last id "+shortLinkName)
	ErrListShortLink           = errcode.NewError(shortLinkBaseCode+9, "failed to list of "+shortLinkName)
	// error codes are globally unique, adding 1 to the previous error code
)

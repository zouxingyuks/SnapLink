// Package types define the structure of request parameters and respond results in this package
package types

import (
	"time"

	"github.com/zhufuyi/sponge/pkg/mysql/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.

// CreateShortLinkRequest 创建短链接请求参数
type CreateShortLinkRequest struct {
	OriginUrl   string `json:"originUrl" binding:""`
	Domain      string `json:"domain" binding:""`
	Gid         string `json:"gid" binding:""`
	CreateType  string `json:"createType" binding:""`
	ValidTime   string `json:"validTime" binding:""`
	Description string `json:"description" binding:""`
	Enable      int    `json:"enable" binding:""`
	Favicon     string `json:"favicon" binding:""`
}

// UpdateShortLinkByIDRequest request params
type UpdateShortLinkByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	OriginUrl   string    `json:"originUrl" binding:""`
	Domain      string    `json:"domain" binding:""`
	Gid         string    `json:"gid" binding:""`
	CreateType  string    `json:"createType" binding:""`
	ValidTime   time.Time `json:"validTime" binding:""`
	Description string    `json:"description" binding:""`
	Enable      int       `json:"enable" binding:""`
	Favicon     string    `json:"favicon" binding:""`
	Uri         string    `json:"uri" binding:""`
	Clicks      int64     `json:"clicks" binding:""`
}

// ShortLinkObjDetail detail
type ShortLinkObjDetail struct {
	ID string `json:"id"` // convert to string id

	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	OriginUrl   string    `json:"originUrl"`
	Domain      string    `json:"domain"`
	Gid         string    `json:"gid"`
	CreateType  string    `json:"createType"`
	ValidTime   time.Time `json:"validTime"`
	Description string    `json:"description"`
	Enable      int       `json:"enable"`
	Favicon     string    `json:"favicon"`
	Uri         string    `json:"uri"`
	Clicks      int64     `json:"clicks"`
}

// CreateShortLinkRespond only for api docs
type CreateShortLinkRespond struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// UpdateShortLinkByIDRespond only for api docs
type UpdateShortLinkByIDRespond struct {
	Result
}

// GetShortLinkByIDRespond only for api docs
type GetShortLinkByIDRespond struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ShortLink ShortLinkObjDetail `json:"shortLink"`
	} `json:"data"` // return data
}

// DeleteShortLinkByIDRespond only for api docs
type DeleteShortLinkByIDRespond struct {
	Result
}

// DeleteShortLinksByIDsRequest request params
type DeleteShortLinksByIDsRequest struct {
	IDs []uint64 `json:"ids" binding:"min=1"` // id list
}

// DeleteShortLinksByIDsRespond only for api docs
type DeleteShortLinksByIDsRespond struct {
	Result
}

// GetShortLinkByConditionRequest request params
type GetShortLinkByConditionRequest struct {
	query.Conditions
}

// GetShortLinkByConditionRespond only for api docs
type GetShortLinkByConditionRespond struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ShortLink ShortLinkObjDetail `json:"shortLink"`
	} `json:"data"` // return data
}

// ListShortLinksByIDsRequest request params
type ListShortLinksByIDsRequest struct {
	IDs []uint64 `json:"ids" binding:"min=1"` // id list
}

// ListShortLinksByIDsRespond only for api docs
type ListShortLinksByIDsRespond struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ShortLinks []ShortLinkObjDetail `json:"shortLinks"`
	} `json:"data"` // return data
}

// ListShortLinksRequest request params
type ListShortLinksRequest struct {
	query.Params
}

// ListShortLinksRespond only for api docs
type ListShortLinksRespond struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ShortLinks []ShortLinkObjDetail `json:"shortLinks"`
	} `json:"data"` // return data
}

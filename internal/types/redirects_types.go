// Package types define the structure of request parameters and respond results in this package
package types

import (
	"time"

	"github.com/zhufuyi/sponge/pkg/mysql/query"
)

var _ time.Time

// Tip: suggested filling in the binding rules https://github.com/go-playground/validator in request struct fields tag.

// CreateRedirectsRequest request params
type CreateRedirectsRequest struct {
	Gid string `json:"gid" binding:""`
	Uri string `json:"uri" binding:""`
}

// UpdateRedirectsByIDRequest request params
type UpdateRedirectsByIDRequest struct {
	ID uint64 `json:"id" binding:""` // uint64 id

	Gid string `json:"gid" binding:""`
	Uri string `json:"uri" binding:""`
}

// RedirectsObjDetail detail
type RedirectsObjDetail struct {
	ID string `json:"id"` // convert to string id

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Gid       string    `json:"gid"`
	Uri       string    `json:"uri"`
}

// CreateRedirectsRespond only for api docs
type CreateRedirectsRespond struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		ID uint64 `json:"id"` // id
	} `json:"data"` // return data
}

// UpdateRedirectsByIDRespond only for api docs
type UpdateRedirectsByIDRespond struct {
	Result
}

// GetRedirectsByIDRespond only for api docs
type GetRedirectsByIDRespond struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		Redirects RedirectsObjDetail `json:"redirects"`
	} `json:"data"` // return data
}

// DeleteRedirectsByIDRespond only for api docs
type DeleteRedirectsByIDRespond struct {
	Result
}

// DeleteRedirectssByIDsRequest request params
type DeleteRedirectssByIDsRequest struct {
	IDs []uint64 `json:"ids" binding:"min=1"` // id list
}

// DeleteRedirectssByIDsRespond only for api docs
type DeleteRedirectssByIDsRespond struct {
	Result
}

// GetRedirectsByConditionRequest request params
type GetRedirectsByConditionRequest struct {
	query.Conditions
}

// GetRedirectsByConditionRespond only for api docs
type GetRedirectsByConditionRespond struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		Redirects RedirectsObjDetail `json:"redirects"`
	} `json:"data"` // return data
}

// ListRedirectssByIDsRequest request params
type ListRedirectssByIDsRequest struct {
	IDs []uint64 `json:"ids" binding:"min=1"` // id list
}

// ListRedirectssByIDsRespond only for api docs
type ListRedirectssByIDsRespond struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		Redirectss []RedirectsObjDetail `json:"redirectss"`
	} `json:"data"` // return data
}

// ListRedirectssRequest request params
type ListRedirectssRequest struct {
	query.Params
}

// ListRedirectssRespond only for api docs
type ListRedirectssRespond struct {
	Code int    `json:"code"` // return code
	Msg  string `json:"msg"`  // return information description
	Data struct {
		Redirectss []RedirectsObjDetail `json:"redirectss"`
	} `json:"data"` // return data
}

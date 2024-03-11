// Package types define the structure of request parameters and respond results in this package
package types

// CreateShortLinkGroupRequest 创建短链接分组请求参数
type CreateShortLinkGroupRequest struct {
	Name string `json:"name" binding:"required"`
}

// UpdateShortLinkGroupByGIDRequest 更新短链接分组请求参数
type UpdateShortLinkGroupByGIDRequest struct {
	Gid  string `json:"gid" binding:"required"`
	Name string `json:"name" binding:"required"`
}

// UpdateShortLinkGroupSortOrderRequest 更新短链接分组排序请求参数
type UpdateShortLinkGroupSortOrderRequest struct {
	Gid       string `json:"gid" binding:"required"`
	SortOrder int    `json:"sortOrder" binding:"required"`
}

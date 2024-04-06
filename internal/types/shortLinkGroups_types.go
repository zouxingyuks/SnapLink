package types

import (
	"SnapLink/internal/model"
	"time"
)

// 命名规则： Handler+Action+Res/Req

// ShortLinkGroupCreateReq 创建短链接分组请求参数
type ShortLinkGroupCreateReq struct {
	Name string `json:"name" binding:"required"`
}

// ShortLinkGroupUpdateByGIDReq 更新短链接分组请求参数
type ShortLinkGroupUpdateByGIDReq struct {
	Gid  string `json:"gid" binding:"required"`
	Name string `json:"name" binding:"required"`
}

// ShortLinkGroupUpdateSortOrderReq 更新短链接分组排序请求参数
type ShortLinkGroupUpdateSortOrderReq struct {
	Gid       string `json:"gid" binding:"required"`
	SortOrder int    `json:"sort_order" binding:"required"`
}

// ShortLinkGroupListItem 分组信息(过滤后)
type ShortLinkGroupListItem struct {
	ID        uint      `json:"id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	SortOrder int       `json:"sort_order,omitempty"`
	Gid       string    `json:"gid,omitempty"`
	Name      string    `json:"name,omitempty"`
	Count     int       `json:"count"`
}

// ShortLinkGroupListRes 分组查询响应
type ShortLinkGroupListRes struct {
	Items []*ShortLinkGroupListItem
}

func NewShortLinkGroupListItem(data map[string]any) *ShortLinkGroupListItem {
	group := data["group"].(*model.ShortLinkGroup)
	count := data["count"].(int64)
	res := &ShortLinkGroupListItem{
		ID:        group.ID,
		CreatedAt: group.CreatedAt,
		UpdatedAt: group.UpdatedAt,
		SortOrder: group.SortOrder,
		Gid:       group.Gid,
		Name:      group.Name,
		Count:     int(count),
	}
	return res
}

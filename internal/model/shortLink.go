package model

import (
	"fmt"
	"gorm.io/gorm"
	"time"
)

type ShortLink struct {
	ID            uint `gorm:"primarykey"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index:uri_deleted"`
	OriginUrl     string         `gorm:"type:nvarchar(255);column:origin_url;comment:'原始链接';not null" json:"origin_url"`
	Domain        string         `gorm:"type:nvarchar(50);column:domain;comment:'域名';" json:"domain"`
	Gid           string         `gorm:"column:gid;comment:'组id';not null;index:idx_gid_uri" json:"gid"`
	CreatedType   int            `gorm:"column:created_type;comment:'创建类型';not null" json:"created_type"`
	ValidDateType int            `gorm:"column:valid_date_type;comment:'有效时间类型';not null" json:"valid_date_type"`
	ValidTime     time.Time      `gorm:"column:valid_time;comment:'有效时间';default:0" json:"valid_time"`
	Description   string         `gorm:"column:description;type:text;comment:'描述'" json:"description"`
	Enable        int            `gorm:"column:enable;type:tinyint(1);comment:'是否启用';default:1" json:"enable"`
	Favicon       string         `gorm:"column:favicon;comment:'网站图标';default:''"`
	Uri           string         `gorm:"type:nvarchar(255);column:uri;comment:'生成短链接的uri';not null;index:idx_gid_uri;index:uri_deleted" json:"uri"`
}

// TName 对应的分表表名
func (s ShortLink) TName() string {
	id := hash(s.Gid)
	return fmt.Sprintf("%s%d", ShortLinkPrefix, id%ShortLinkShardingNum)
}

package model

import (
	"gorm.io/gorm"
	"strconv"
	"time"
)

type ShortLink struct {
	gorm.Model
	OriginUrl   string    `gorm:"type:nvarchar(255);column:origin_url;commit:'原始链接';not null" json:"origin_url"`
	Domain      string    `gorm:"type:nvarchar(255);column:domain;commit:'域名';not null"`
	Gid         string    `gorm:"column:gid;commit:'组id';not null" json:"gid"`
	CreateType  string    `gorm:"column:create_type;commit:'创建类型';not null" json:"create_type"`
	ValidTime   time.Time `gorm:"column:valid_time;commit:'有效时间';default:0" json:"valid_time"`
	Description string    `gorm:"column:description;type:text;commit:'描述'" json:"description"`
	Enable      int       `gorm:"column:enable;type:tinyint(1);commit:'是否启用';default:1" json:"enable"`
	Favicon     string    `gorm:"column:favicon;commit:'网站图标';default:''"`
	Uri         string    `gorm:"type:nvarchar(255);column:uri;commit:'生成短链接的uri';not null"`
	Clicks      int64     `gorm:"column:clicks;commit:'点击次数';default:0"`
}

func (s ShortLink) TableName() string {
	gid, err := strconv.Atoi(s.Gid)
	if err != nil {
		return "short_link_0"
	}
	return "short_link_" + strconv.Itoa(gid%25)
}

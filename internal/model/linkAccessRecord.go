package model

import (
	"fmt"
	"time"
)

type LinkAccessRecord struct {
	ID          uint64     `gorm:"column:id;type:bigint(20) unsigned;primary_key;AUTO_INCREMENT" json:"id"`
	Datetime    string     `gorm:"column:date;type:DATETIME;commit:'访问时间'"`
	DeletedAt   *time.Time `gorm:"column:deleted_at;type:datetime(3)" json:"-"`
	OriginalURL string     `gorm:"column:originalURL;type:varchar(255)" json:"originalurl"`
	Gid         int        `gorm:"column:gid;type:int(11)" json:"gid"`
	IP4         string     `gorm:"column:ip4;type:varchar(20);commit:'ipv4地址'" json:"ip4"`
	//IP6         string     `gorm:"column:ip6;type:varchar(50);commit:'ipv6地址'" json:"ip6"`
	Device    string `gorm:"column:device;type:varchar(20);commit:'设备类型'" json:"device"`
	UserAgent string `gorm:"column:userAgent;type:varchar(50);commit:'UA'" json:"userAgent"`
	Explorer  string `gorm:"column:explorer;type:varchar(20);commit:'浏览器类型'" json:"explorer"`
	Network   string `gorm:"column:network;type:varchar(20);commit:'网络类型'" json:"network"`
	Hour      int    `gorm:"-"`
	Date      string `gorm:"-"`
}

func (l LinkAccessRecord) TName() string {
	return fmt.Sprintf("link_access_statistic_%d", l.Gid%25)
}

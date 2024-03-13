package model

import (
	"fmt"
	"time"
)

type LinkAccessRecord struct {
	ID          uint64     `gorm:"column:id;type:bigint(20) unsigned;primary_key;AUTO_INCREMENT" json:"id"`
	URI         string     `gorm:"column:uri;type:varchar(255);comment:'访问链接';index:idx_uri" json:"uri"`
	Datetime    string     `gorm:"column:datetime;type:DATETIME;comment:'访问时间';index:idx_date" json:"datetime"`
	DeletedAt   *time.Time `gorm:"column:deleted_at;type:datetime(3)" json:"-"`
	OriginalURL string     `gorm:"column:originalURL;type:varchar(255)" json:"originalurl"`
	Gid         int        `gorm:"column:gid;type:int(11)" json:"gid"`
	IP4         string     `gorm:"column:ip4;type:varchar(20);comment:'ipv4地址'" json:"ip4"`
	IP6         string     `gorm:"column:ip6;type:varchar(50);comment:'ipv6地址'" json:"ip6"`
	Device      string     `gorm:"column:device;type:varchar(20);comment:'设备类型'" json:"device"`
	UserAgent   string     `gorm:"column:userAgent;type:text;comment:'UA'" json:"userAgent"`
	Browser     string     `gorm:"column:browser;type:varchar(20);comment:'浏览器'" json:"browser"`
	Network     string     `gorm:"column:network;type:varchar(20);comment:'网络类型'" json:"network"`
	Local       string     `gorm:"column:local;type:varchar(20);comment:'地区'" json:"local"`
	Hour        int        `gorm:"-"`
	Date        string     `gorm:"-"`
}

func (l LinkAccessRecord) TName() string {
	//todo 分库分表设计
	return fmt.Sprintf("link_access_records")
}

package model

import (
	"fmt"
	"gorm.io/gorm"
	"time"
)

// LinkAccessRecord 访问记录表
type LinkAccessRecord struct {
	ID          uint      `gorm:"primarykey"`
	CreatedAt   time.Time `gorm:"uniqueIndex:uidx_uri_requestID_date,priority:2"`
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
	URI         string         `gorm:"column:uri;type:varchar(255);comment:'访问链接';uniqueIndex:uidx_uri_requestID_date,priority:1" json:"uri"`
	Gid         string         ` gorm:"column:gid;type:varchar(50);comment:'gid';index:idx_gid" json:"gid"`
	IP4         string         `gorm:"column:ip4;type:varchar(20);comment:'ipv4地址'" json:"ip4"`
	IP6         string         `gorm:"column:ip6;type:varchar(50);comment:'ipv6地址'" json:"ip6"`
	OriginalURL string         `gorm:"column:originalURL;type:varchar(255)" json:"originalurl"`
	UserAgent   string         `gorm:"column:userAgent;type:varchar(255);comment:'用户代理'" json:"userAgent"`
	Device      string         `gorm:"column:device;type:varchar(20);comment:'设备类型'" json:"device"`
	OS          string         `gorm:"column:os;type:varchar(20);comment:'操作系统'" json:"os"`
	Browser     string         `gorm:"column:browser;type:varchar(20);comment:'浏览器'" json:"browser"`
	Network     string         `gorm:"column:network;type:varchar(20);comment:'网络类型'" json:"network"`
	Local       string         `gorm:"column:local;type:nvarchar(20);comment:'地区'" json:"local"`
	Date        string         `gorm:"column:date;type:varchar(10);comment:'日期';index:idx_date" json:"date"`
	RequestID   string         `gorm:"column:requestID;type:varchar(50);comment:'请求ID';uniqueIndex:uidx_uri_requestID_date,priority:3" json:"requestID"`
	Hour        int            `gorm:"-"`
}

// TName 对应的分表表名
// 因为查询记录的时候通常是查询某个链接的访问记录，所以这里的分表规则是根据URI进行分表
func (l LinkAccessRecord) TName() string {
	id := hash(l.URI)
	return fmt.Sprintf("link_access_record_%d", id%LinkAccessRecordShardingNum)
}

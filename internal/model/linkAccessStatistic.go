package model

import (
	"gorm.io/gorm"
)
import "gorm.io/datatypes"

type LinkAccessStatistic struct {
	gorm.Model `json:"-"`
	URI        string         `gorm:"column:uri;type:varchar(255);commit:'访问链接';uniqueIndex:idx_query;index:idx_uri_date" json:"uri"`
	Pv         int64          `gorm:"column:pv;type:bigint(20)" json:"pv"`
	Uv         int64          `gorm:"column:uv;type:bigint(20)" json:"uv"`
	Uip        int64          `gorm:"column:uip;type:bigint(20)" json:"uip"`
	Datetime   string         `gorm:"column:datetime;type:datetime;uniqueIndex:idx_query" json:"datetime"`
	Date       string         `gorm:"column:date;type:date;index:idx_date;index:idx_uri_date" json:"date"`
	Hour       int            `gorm:"column:hour;type:int(11)" json:"hour"`
	Weekday    int            `gorm:"column:weekday;type:int(11)" json:"weekday"`
	Regions    datatypes.JSON `gorm:"column:regions;type:json" json:"regions"`
	IPs        datatypes.JSON `gorm:"column:ips;type:json" json:"ips"`
	Devices    datatypes.JSON `gorm:"column:devices;type:json" json:"devices"`
}

func (l LinkAccessStatistic) TName() string {
	return "link_access_statistics"
}

type LinkAccessStatisticDay struct {
	URI      string `gorm:"uri" json:"uri"`
	TodayPv  int64  `gorm:"today_pv" json:"today_pv"`
	TodayUv  int64  `gorm:"today_uv" json:"today_uv"`
	TodayUip int64  `gorm:"today_uip" json:"today_uip"`
	Date     string `gorm:"date" json:"date"`
}

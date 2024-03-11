package model

import (
	"fmt"
	"gorm.io/gorm"
	"time"
)

type ShortLinkGroup struct {
	ID        uint           `gorm:"primarykey" json:"-"`
	CreatedAt time.Time      ` json:"-"`
	UpdatedAt time.Time      ` json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index:idx" json:"-"`
	SortOrder int            `gorm:"column:sort_order;type:int;NOT NULL;default:0;index:idx;commit:'排序标识'" json:"sort_order"`
	Gid       string         `gorm:"column:gid;NOT NULL;commit:'分组 id';index:idx" json:"gid"`
	Name      string         `gorm:"column:name;type:varchar(50);NOT NULL;commit:'分组名'" json:"name"`
	CUsername string         `gorm:"column:c_username;type:varchar(50);NOT NULL;commit:'创建人';index:idx" json:"cUser"`
}

// TName 根据创建人进行分表
func (s ShortLinkGroup) TName() string {
	id := hash(s.CUsername)
	return fmt.Sprintf("short_link_group_%d", id%SLGroupShardingNum)
}

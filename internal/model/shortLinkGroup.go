package model

import (
	"fmt"
	"gorm.io/gorm"
)

type ShortLinkGroup struct {
	gorm.Model
	SortOrder   int    `gorm:"column:sort_order;type:int;NOT NULL;default:0" json:"sort_order"`
	Description string `gorm:"column:description;type:text" json:"description"`
	Enable      bool   `gorm:"column:enable;type:tinyint(1);default:1" json:"enable"`
	Gid         int    `gorm:"column:gid;NOT NULL;uniqueIndex:idx_name_c_userid" json:"gid"`
	Name        string `gorm:"column:name;type:varchar(50);NOT NULL" json:"name"`
	CUserId     string `gorm:"column:c_user_id;type:varchar(50);NOT NULL;uniqueIndex:idx_name_c_userid" json:"cUser"`
}

func (s ShortLinkGroup) TableName() string {
	return fmt.Sprintf("short_link_group_%d", s.ID%16)
}

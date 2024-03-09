package model

import (
	"fmt"
	"time"
)

type TUser struct {
	ID           uint64     `gorm:"column:id;type:bigint(20);primary_key;AUTO_INCREMENT" json:"id"`
	Username     string     `gorm:"column:username;type:varchar(256);commit:'用户名'" json:"username"`
	Password     string     `gorm:"column:password;type:varchar(512);commit:'密码'" json:"password"`
	RealName     string     `gorm:"column:real_name;type:varchar(256);commit:'真实姓名'" json:"realName"`
	Phone        string     `gorm:"column:phone;type:varchar(128);commit:'手机号'" json:"phone"`
	Mail         string     `gorm:"column:mail;type:varchar(512);commit:'邮箱'" json:"mail"`
	DeletionTime int64      `gorm:"column:deletion_time;type:bigint(20);commit:'注销时间戳'" json:"deletionTime"` //
	CreateTime   *time.Time `gorm:"column:create_time;type:datetime;commit:'创建时间'" json:"createTime"`        //
	UpdateTime   *time.Time `gorm:"column:update_time;type:datetime;commit:'修改时间'" json:"updateTime"`        //
	DelFlag      int        `gorm:"column:del_flag;type:tinyint(1);commit:'删除标识'" json:"delFlag"`            //  0：未删除 1：已删除
}

// TName 基于 Username 进行分库分表
func (u *TUser) TName() string {
	//对 username 进行取模分表
	id := hash(u.Username)
	return fmt.Sprintf("t_user%d", id%16)
}

// hash 计算字符串的哈希值
func hash(s string) uint64 {
	h := uint64(0)
	for i := 0; i < len(s); i++ {
		h = 31*h + uint64(s[i])
	}
	return h
}

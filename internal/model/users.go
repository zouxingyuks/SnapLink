package model

import (
	"fmt"
	"gorm.io/gorm"
)

type TUser struct {
	gorm.Model `json:"-"`
	Username   string `gorm:"column:username;type:nvarchar(20);comment:'用户名';uniqueIndex" json:"username"`
	Password   string `gorm:"column:password;type:varchar(80);comment:'密码'" json:"password"`
	RealName   string `gorm:"column:real_name;type:nvarchar(20);comment:'真实姓名'" json:"realName"`
	Phone      string `gorm:"column:phone;type:varchar(20);comment:'手机号';uniqueIndex" json:"phone"`
	Mail       string `gorm:"column:mail;type:varchar(50);comment:'邮箱';uniqueIndex" json:"mail"`
}

// TName 基于 Username 进行分库分表
func (u *TUser) TName() string {
	//对 username 进行取模分表
	id := hash(u.Username)
	return fmt.Sprintf("%s%d", TUserPrefix, id%TUserShardingNum)
}

// hash 计算字符串的哈希值
func hash(s string) uint64 {
	h := uint64(0)
	for i := 0; i < len(s); i++ {
		h = 31*h + uint64(s[i])
	}
	return h
}

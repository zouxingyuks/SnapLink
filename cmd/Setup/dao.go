package main

import (
	"SnapLink/internal/model"
	"fmt"
	"gorm.io/gorm"
	"sync"
)

type generateTables func(db *gorm.DB)

var fns = []generateTables{
	generateShortLink,
	generateShortLinkGroup,
	generateTUser,
	generateRedirect,
	generateLinkAccessRecord,
	generateLinkAccessStatistic,
}

func daoInit() {
	DB := model.GetDB()
	// 自动迁移模式
	DB.AutoMigrate(
		model.Config{},
		// 在此处填入需要迁移的数据类型
	)
	wg := new(sync.WaitGroup)
	//有分表的话，需要在此处添加分表的迁移
	for _, fn := range fns {
		wg.Add(1)
		go func(db *gorm.DB) {
			fn(db)
			wg.Done()
		}(DB)
	}
	wg.Wait()

}

func generateShortLinkGroup(db *gorm.DB) {
	shortLinkGroup := model.ShortLinkGroup{}
	for i := 0; i < model.SLGroupShardingNum; i++ {
		tName := fmt.Sprintf("t_link_group%d", i)
		// 先判断是否存在表，不存在则创建
		if !db.Migrator().HasTable(tName) {
			db.Table(tName).AutoMigrate(shortLinkGroup)
		}
	}
}

// 分表 TUser
func generateTUser(db *gorm.DB) {
	tUser := model.TUser{}
	for i := 0; i < model.TUserShardingNum; i++ {
		// 先判断是否存在表，不存在则创建
		tName := fmt.Sprintf("%s%d", model.TUserPrefix, i)
		if !db.Migrator().HasTable(tName) {
			db.Table(tName).AutoMigrate(tUser)
		}
	}

}

// 分表 shortLink
func generateShortLink(db *gorm.DB) {
	shortLink := model.ShortLink{}
	for i := 0; i < model.ShortLinkShardingNum; i++ {
		// 先判断是否存在表，不存在则创建
		tName := fmt.Sprintf("%s%d", model.ShortLinkPrefix, i)
		if !db.Migrator().HasTable(tName) {
			db.Table(tName).AutoMigrate(shortLink)
		}
	}
}

// 分表 redirect
func generateRedirect(db *gorm.DB) {
	redirectInfo := model.Redirect{}
	for i := 0; i < model.RedirectShardingNum; i++ {
		// 先判断是否存在表，不存在则创建
		tName := fmt.Sprintf("redirect_%d", i)
		if !db.Migrator().HasTable(tName) {
			db.Table(tName).AutoMigrate(redirectInfo)
		}
	}
}

// 分表 LinkAccessRecord
func generateLinkAccessRecord(db *gorm.DB) {
	linkAccessRecord := model.LinkAccessRecord{}
	for i := 0; i < model.LinkAccessRecordShardingNum; i++ {
		// 先判断是否存在表，不存在则创建
		tName := fmt.Sprintf("link_access_record_%d", i)
		if !db.Migrator().HasTable(tName) {
			db.Table(tName).AutoMigrate(linkAccessRecord)
		}
	}
}

// 分表 LinkAccessStatistic
func generateLinkAccessStatistic(db *gorm.DB) {
	linkAccessStatistic := model.LinkAccessStatistic{}
	for i := 0; i < model.LinkAccessStatisticShardingNum; i++ {
		// 先判断是否存在表，不存在则创建
		tName := fmt.Sprintf("link_access_statistic_%d", i)
		if !db.Migrator().HasTable(tName) {
			db.Table(tName).AutoMigrate(linkAccessStatistic)
		}
	}
}

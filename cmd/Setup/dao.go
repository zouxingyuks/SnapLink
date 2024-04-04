package main

import (
	"SnapLink/internal/model"
	"fmt"
	"github.com/zhufuyi/sponge/pkg/logger"
	"gorm.io/gorm"
	"sync"
)

type generateTables func(db *gorm.DB)

var fns = []generateTables{
	generateTableFunc(model.ShortLinkGroup{}, model.SLGroupPrefix, model.SLGroupShardingNum),
	generateTableFunc(model.TUser{}, model.TUserPrefix, model.TUserShardingNum),
	generateTableFunc(model.ShortLink{}, model.ShortLinkPrefix, model.ShortLinkShardingNum),
	generateTableFunc(model.Redirect{}, model.RedirectPrefix, model.RedirectShardingNum),
	//generateTableFunc(model.LinkAccessRecord{}, model.LinkAccessRecordPrefix, model.LinkAccessRecordShardingNum),
	//generateTableFunc(model.LinkAccessStatistic{}, model.LinkAccessStatisticPrefix, model.LinkAccessStatisticShardingNum),
}

func daoInit() {
	DB := model.GetDB()
	DB.AutoMigrate(
	// 在此处填入需要迁移的数据类型
	)

	wg := new(sync.WaitGroup)
	for _, fn := range fns {
		wg.Add(1)
		go func(fn generateTables, db *gorm.DB) {
			defer wg.Done()
			fn(db)
		}(fn, DB)
	}
	wg.Wait()
}

func generateTableFunc(table interface{}, prefix string, shardingNum int) generateTables {
	return func(db *gorm.DB) {
		existTables := make(map[string]struct{}, shardingNum)
		query := `SELECT table_name FROM information_schema.tables WHERE table_schema = ? AND table_type = 'BASE TABLE' AND table_name LIKE ?`
		rows, err := db.Raw(query, db.Migrator().CurrentDatabase(), prefix+"%").Rows()
		if err != nil {
			logger.Panic(err.Error())
			return
		}
		defer rows.Close()

		for rows.Next() {
			var tableName string
			if err := rows.Scan(&tableName); err != nil {
				logger.Panic(err.Error())
				return
			}
			existTables[tableName] = struct{}{}
		}
		for i := 0; i < shardingNum; i++ {
			tableName := fmt.Sprintf("%s-%d", prefix, i)
			if _, ok := existTables[tableName]; !ok {
				if err := db.Table(tableName).AutoMigrate(table); err != nil {
					fmt.Printf("Failed to migrate table %s: %v\n", tableName, err)
				}
			}
		}
	}
}

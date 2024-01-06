package db

import (
	"github.com/pkg/errors"
	"go-ssas/service/conf"
	"gorm.io/gorm"
)

func DB() *gorm.DB {
	dbInstance.Do(func() {
		var err error
		//在此处进行修改数据库的连接以及数据库类型
		dbInstance.DB, err = connMysql(
			conf.DataBase().User,
			conf.DataBase().Password,
			conf.DataBase().Host,
			conf.DataBase().Port,
			conf.DataBase().Name,
			conf.DataBase().Charset,
		)
		if err != nil {
			panic(errors.Wrap(err, "connect to database failed"))
		}
		initDao(dbInstance.DB)
	})
	return dbInstance.Session(&gorm.Session{NewDB: false})
}
func migration(db *gorm.DB) {
	// 自动迁移模式
	db.AutoMigrate(
	// 在此处填入需要迁移的数据类型

	)

}

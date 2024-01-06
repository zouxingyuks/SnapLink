package db

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"sync"
	"time"
)

var dbInstance = new(struct {
	*gorm.DB
	sync.Once
})

// 连接MySQL数据库
//
//go:generate go get -u gorm.io/driver/mysql
func connMysql(user, password, host string, port int, name, charset string) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local", user, password, host, port, name, charset)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	return db, err
}

// 连接SQLite数据库
//
//go:generate go get -u gorm.io/driver/sqlite
func connSQLite(filePath string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(filePath), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	return db, err
}

// initDao 连接数据库
func initDao(db *gorm.DB) {
	// 设置连接池
	//设置连接池
	sqlDB, _ := db.DB()
	//设置空闲连接池中的最大连接数
	sqlDB.SetMaxIdleConns(10)
	//设置打开的最大连接数
	sqlDB.SetMaxOpenConns(100)
	//设置连接的最大可复用时间
	sqlDB.SetConnMaxLifetime(time.Hour)
	//执行迁移
	migration(db)
}

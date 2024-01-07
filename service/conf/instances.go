package conf

import (
	"github.com/pkg/errors"
	"sync"
)

var defaultConfig = map[string]any{
	"system": system{
		Env: "dev",
	},
	"database": database{
		User:       "",
		Password:   "",
		Host:       "",
		Name:       "",
		Port:       3306,
		UnixSocket: false,
	},
}

// SystemConfig 系统配置
type system struct {
	Env  string
	Host string
}

var systemInstance = new(struct {
	sync.Once
	*system
})

func System() *system {
	systemInstance.Do(
		func() {
			err := config().Sub("system").Unmarshal(&systemInstance.system)
			if err != nil {
				panic(errors.New("init systemConfig...failed").Error())
			}
		})
	return systemInstance.system

}

type database struct {
	User       string
	Password   string
	Host       string
	Name       string
	Port       int
	UnixSocket bool
}

var databaseInstance = new(struct {
	sync.Once
	*database
})

func DataBase() *database {
	databaseInstance.Do(
		func() {
			err := config().Sub("database").Unmarshal(&databaseInstance.database)
			if err != nil {
				panic(errors.New("init databaseConfig...failed").Error())
			}
		})
	return databaseInstance.database

}

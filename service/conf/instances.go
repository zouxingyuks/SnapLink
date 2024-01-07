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
	"redis": redis{
		Network:  "tcp",
		Server:   "localhost:6379",
		PoolSize: 10,
		DB:       0,
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

type redis struct {
	Network  string
	Server   string
	User     string
	Password string
	PoolSize int
	DB       int
}

var redisInstance = new(struct {
	sync.Once
	*redis
})

func Redis() *redis {
	redisInstance.Do(
		func() {
			err := config().Sub("redis").Unmarshal(&redisInstance.redis)
			if err != nil {
				panic(errors.New("init redisConfig...failed").Error())
			}
		})
	return redisInstance.redis
}

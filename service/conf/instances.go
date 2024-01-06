package conf

import (
	"github.com/pkg/errors"
	"sync"
)

var defaultConfig = map[string]any{
	"system": system{
		Env: "dev",
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

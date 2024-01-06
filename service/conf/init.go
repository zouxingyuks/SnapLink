package conf

import (
	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"log"
	"os"
	"path"
	"sync"
)

var configInstance = new(struct {
	*viper.Viper
	sync.Once
})

func config() *viper.Viper {
	configInstance.Once.Do(
		func() {
			configInstance.Viper = viper.New()
			log.Println("init config")
			configInstance.MergeConfigMap(defaultConfig)
			configDir := "./"
			configName := "config"
			configType := "yaml"
			//设置配置文件路径

			//将默认值设置到config中
			configInstance.AddConfigPath(configDir)
			configInstance.SetConfigName(configName)
			configInstance.SetConfigType(configType)

			// 配置文件出错
			if err := configInstance.ReadInConfig(); err != nil {
				// 如果找不到配置文件，则提醒生成配置文件并创建它
				var configFileNotFoundError viper.ConfigFileNotFoundError
				if errors.As(err, &configFileNotFoundError) {
					// 如果 config 目录不存在，则创建它
					if _, err := os.Stat(configDir); os.IsNotExist(err) {
						if err = os.MkdirAll(configDir, 0755); err != nil {
							log.Panic(errors.Wrapf(err, "[error] Failed to create config directory. %s\n", configDir))
						}
					}
					configPath := path.Join(configDir, configName+"."+configType)
					log.Println(errors.Wrapf(err, "[warning] Config file not found. Generating default config file at %s\n", configPath))
					if err := configInstance.WriteConfigAs(configPath); err != nil {
						log.Panic(errors.Wrapf(err, "[error] Failed to generate default config file. %s\n", configPath))
					}
					// 再次读取配置文件
					if err := configInstance.ReadInConfig(); err != nil {
						log.Panic(errors.Wrapf(err, "[error] Failed to read config file. %s\n", configPath))
					}
					panic("请修改配置文件后重启程序")
				}
			}
			configInstance.WatchConfig()
			configInstance.OnConfigChange(func(e fsnotify.Event) {
				ReloadFunc()
			})
		})
	return configInstance.Viper
}

func AddPath(path string) {
	log.Printf("add path %s to config\n", path)
	configInstance.Viper.AddConfigPath(path)
}

// ReloadFunc 此处是配置文件变更后的回调函数
func ReloadFunc() {
	if err := configInstance.ReadInConfig(); err != nil {
		log.Panic(errors.Wrap(err, "Failed to reload config"))
	}
}

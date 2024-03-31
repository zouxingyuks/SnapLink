package service

import (
	"SnapLink/configs"
	config2 "SnapLink/internal/config"
	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/config"
	"github.com/alibaba/sentinel-golang/ext/datasource"
	"github.com/alibaba/sentinel-golang/ext/datasource/file"
	"github.com/alibaba/sentinel-golang/pkg/datasource/nacos"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/pkg/errors"
	"github.com/zhufuyi/sponge/pkg/app"
	"github.com/zhufuyi/sponge/pkg/logger"
)

var _ app.IServer = (*sentinelService)(nil)

const (
	WebAppType = 1
)

type sentinelService struct {
	conf *config.Entity
}

func NewSentinelService() *sentinelService {
	s := new(sentinelService)
	s.conf = config.NewDefaultConfig()
	s.conf.Version = "v1.0"
	s.conf.Sentinel.App = struct {
		Name string
		Type int32
	}{Name: "snaplink", Type: WebAppType}
	return s
}

func (s *sentinelService) Start() error {
	err := sentinel.InitWithConfig(s.conf)
	if err != nil {
		logger.Panic(errors.Wrap(err, "sentinel init failed").Error())
		return err
	}
	// 加载规则
	switch config2.Get().Sentinel.SourceType {
	case "file":
		{
			err = loadRulesFromFile(configs.Path("Sentinel.json"))
			if err != nil {
				logger.Panic(errors.Wrap(err, "加载 sentinel 规则失败").Error())
			}
		}
	case "nacos":
		{
			nacosConfig := config2.Get().Sentinel.Nacos
			//nacos server config
			sc := []constant.ServerConfig{
				{
					ContextPath: nacosConfig.ContextPath,
					Port:        uint64(nacosConfig.Port),
					IpAddr:      nacosConfig.IPAddr,
				},
			}
			//nacos client config
			cc := constant.ClientConfig{
				TimeoutMs: 5000,
			}
			//build nacos config client
			client, err := clients.CreateConfigClient(map[string]interface{}{
				"serverConfigs": sc,
				"clientConfig":  cc,
			})
			if err != nil {
				logger.Panic(errors.Wrap(err, "创建 nacos 客户端失败").Error())
			}
			ds, err := nacos.NewNacosDataSource(client, nacosConfig.Group, nacosConfig.DataID, datasource.NewDefaultPropertyHandler(datasource.FlowRuleJsonArrayParser, datasource.FlowRulesUpdater))
			if err != nil {
				logger.Panic(errors.Wrap(err, "创建 nacos 数据源失败").Error())
			}
			if err = ds.Initialize(); err != nil {
				logger.Panic(errors.Wrap(err, "Fail to initialize nacos data source client, err: %+v").Error())
			}
		}
	default:
		{
			logger.Panic(errors.New("不支持的规则数据源").Error())
		}
	}
	return nil
}

func (s *sentinelService) Stop() error {
	logger.Info("sentinelService stopping...")
	return nil
}

func (s *sentinelService) String() string {
	return "sentinelService"
}

func loadRulesFromFile(filePath string) error {
	// 注册流控规则数据源
	h := datasource.NewDefaultPropertyHandler(datasource.FlowRuleJsonArrayParser, datasource.FlowRulesUpdater)
	ds := file.NewFileDataSource(filePath, h)
	err := ds.Initialize()
	if err != nil {
		return err
	}
	src, err := ds.ReadSource()
	if err != nil {
		return err
	}
	err = ds.Handle(src)
	if err != nil {
		return err
	}
	return nil

}

//
//func loadRulesFromNacos() error {
//	// 注册流控规则数据源
//	h := datasource.NewDefaultPropertyHandler(datasource.FlowRuleJsonArrayParser, datasource.FlowRulesUpdater)
//	ds := nacos.NewNacosDataSource("nacos", "sentinel", "flow", h)
//	err := ds.Initialize()
//	if err != nil {
//		return err
//	}
//	src, err := ds.ReadSource()
//	if err != nil {
//		return err
//	}
//	err = ds.Handle(src)
//	if err != nil {
//		return err
//	}
//	return nil
//}

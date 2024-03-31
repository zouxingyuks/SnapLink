package service

import (
	"SnapLink/internal/config"
	logger2 "SnapLink/internal/logger"
	"SnapLink/internal/messageQueue/rabbitmq"
	"context"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/plugin"
	"github.com/pkg/errors"
	"github.com/zhufuyi/sponge/pkg/app"
	"github.com/zhufuyi/sponge/pkg/logger"
	"net/url"
	"sync"
)

// 本项目中读多写少,适合使用 Cache Aside 策略
// 同时,使用 canal 来去进一步解耦,提高性能

var _ app.IServer = (*cacheASideService)(nil)

type cacheASideService struct {
	subscribers []message.Subscriber
	publishers  []message.Publisher
	ctx         context.Context
	cancel      context.CancelFunc
	mutex       sync.Mutex
	mq          *rabbitmq.MQ
}

// NewCacheASideService 新增旁路缓存更新服务
func NewCacheASideService() app.IServer {
	s := new(cacheASideService)
	s.ctx, s.cancel = context.WithCancel(context.Background())
	var err error
	conf := config.Get().RabbitMQ
	// 初始化 rabbitmq
	// 此处必须使用 url.URL，因为如果直接使用字符串拼接，会导致密码中的特殊字符被转义
	u := url.URL{
		Scheme: "amqp",
		User:   url.UserPassword(conf.User, conf.Password),
		Host:   conf.Addr,
		Path:   "/maxwell",
	}
	s.mq, err = rabbitmq.NewMQ(u.String(), new(logger2.WatermillAdapter))
	if err != nil {
		logger.Panic(err.Error())
		return nil
	}
	return s
}

// option 选项
var cacheASideServiceHandlerMap = []option{
	{
		Name:          "AccessLogHandler",
		SubTopic:      "AccessLog",
		HasPublisher:  false,
		Fn:            nil,
		noPublisherFn: handleAccessLog,
	},
}

type routerParam struct {
	mutex       *sync.Mutex
	subscribers []message.Subscriber
	publishers  []message.Publisher
}

func registerRouter(mq *rabbitmq.MQ, p *routerParam, handlers []option) (*message.Router, error) {
	// 启动消息路由
	router, err := message.NewRouter(message.RouterConfig{}, new(logger2.WatermillAdapter))
	if err != nil {
		return nil, errors.Wrap(err, "start accessWatcherService...failed")
	}
	// 添加信号处理插件
	router.AddPlugin(plugin.SignalsHandler)
	n := len(handlers)
	subscriber, err := mq.NewSubscriber()
	// 添加订阅者,并加入到订阅者列表,以便后续关闭
	p.mutex.Lock()
	p.subscribers = append(p.subscribers, subscriber)
	p.mutex.Unlock()
	if err != nil {
		return nil, errors.Wrap(err, "init accessWatcherService...failed")
	}
	// 取出有发布者和无发布者的处理函数
	Fns := make([]message.HandlerFunc, 0, n)
	NoPublisherFns := make([]message.NoPublishHandlerFunc, 0, n)
	for i := 0; i < n; i++ {
		if handlers[i].HasPublisher {
			Fns = append(Fns, handlers[i].Fn)
		} else {
			NoPublisherFns = append(NoPublisherFns, handlers[i].noPublisherFn)
		}
	}
	// 注册消息处理函数
	for fn := range NoPublisherFns {
		router.AddNoPublisherHandler(handlers[fn].Name, handlers[fn].SubTopic, subscriber, handlers[fn].noPublisherFn)
	}
	if len(Fns) > 0 {
		publisher, err := rabbitmq.NewPublisher()
		if err != nil {
			logger.Panic(errors.Wrap(err, "init accessWatcherService...failed").Error())
		}
		for fn := range Fns {
			router.AddHandler(handlers[fn].Name, handlers[fn].SubTopic, subscriber, handlers[fn].PubTopic, publisher, watcherServiceHandlerMap[fn].Fn)
		}
	}
	return router, nil
}

func (s *cacheASideService) Start() error {
	// 启动数据聚合服务
	go batchCreateRecord(s.ctx)
	p := &routerParam{
		mutex:       &s.mutex,
		subscribers: s.subscribers,
		publishers:  s.publishers,
	}
	for i := 1; i <= 100; i++ {
		go func() {
			router, err := registerRouter(s.mq, p, cacheASideServiceHandlerMap)
			if err != nil {
				logger.Err(errors.Wrap(err, "start accessWatcherService...failed"))
				//todo 继续从此处开些
			}
			if err := router.Run(context.Background()); err != nil {
				logger.Err(errors.Wrap(err, "start accessWatcherService...failed"))
			}
		}()
	}
	return nil
}

func (s *cacheASideService) Stop() error {
	//TODO implement me
	panic("implement me")
}

func (s *cacheASideService) String() string {
	//TODO implement me
	panic("implement me")
}

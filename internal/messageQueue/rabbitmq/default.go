package rabbitmq

import (
	"SnapLink/internal/config"
	logger2 "SnapLink/internal/logger"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/pkg/errors"
	"github.com/zhufuyi/sponge/pkg/logger"
	"net/url"
	"sync"
)

var rabbitmq struct {
	mq   *MQ
	once sync.Once
}

func defaultMQ() *MQ {
	rabbitmq.once.Do(func() {
		var err error
		conf := config.Get().RabbitMQ
		// 初始化 rabbitmq
		// 此处必须使用 url.URL，因为如果直接使用字符串拼接，会导致密码中的特殊字符被转义
		u := url.URL{
			Scheme: "amqp",
			User:   url.UserPassword(conf.User, conf.Password),
			Host:   conf.Addr,
			Path:   conf.VirtualHost,
		}
		rabbitmq.mq, err = NewMQ(u.String(), new(logger2.WatermillAdapter))
		// 需要手动确认消息,保证消息不丢失
		rabbitmq.mq.conf.Consume.NoWait = false
		if err != nil {
			// mq 初始化失败是十分严重的错误，所以这里使用了panic
			logger.Panic(errors.Wrap(err, "init rabbitmq...failed").Error())
		}
	})
	return rabbitmq.mq
}

// NewPublisher 创建发布者
func NewPublisher() (message.Publisher, error) {
	return defaultMQ().NewPublisher()
}

// NewSubscriber 创建订阅者
func NewSubscriber() (message.Subscriber, error) {
	return defaultMQ().NewSubscriber()
}

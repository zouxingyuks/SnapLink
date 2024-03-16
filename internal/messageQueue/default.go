package messageQueue

import (
	"SnapLink/internal/config"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-amqp/pkg/amqp"
	"github.com/pkg/errors"
	"github.com/zhufuyi/sponge/pkg/logger"
	"go.uber.org/zap"
	"net/url"
	"sync"
)

var rocketmq struct {
	mq   *MQ
	once sync.Once
}

func defaultMQ() *MQ {
	rocketmq.once.Do(func() {
		var err error
		conf := config.Get().RocketMQ
		// 初始化 rocketmq
		// 此处必须使用 url.URL，因为如果直接使用字符串拼接，会导致密码中的特殊字符被转义
		u := url.URL{
			Scheme: "amqp",
			User:   url.UserPassword(conf.User, conf.Password),
			Host:   conf.Addr,
			Path:   conf.VirtualHost,
		}
		rocketmq.mq, err = NewMQ(u.String(), new(LoggerAdapter))
		if err != nil {
			// mq 初始化失败是十分严重的错误，所以这里使用了panic
			logger.Panic(errors.Wrap(err, "init rocketmq...failed").Error())
		}
	})
	return rocketmq.mq
}

// NewPublisher 创建发布者
func NewPublisher() (*amqp.Publisher, error) {
	return defaultMQ().NewPublisher()
}

// NewSubscriber 创建订阅者
func NewSubscriber() (*amqp.Subscriber, error) {
	return defaultMQ().NewSubscriber()
}

// LoggerAdapter 日志适配器
type LoggerAdapter struct {
	fields []zap.Field // 用于存储累积的字段
}

// 转换 Watermill LogFields 到 zap.Fields
func (l *LoggerAdapter) toZapFields(fields watermill.LogFields) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields)+len(l.fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	return append(zapFields, l.fields...)
}

func (l *LoggerAdapter) Error(msg string, err error, fields watermill.LogFields) {
	logger.Error(errors.Wrap(err, msg).Error(), l.toZapFields(fields)...)
}

func (l *LoggerAdapter) Info(msg string, fields watermill.LogFields) {
	logger.Info(msg, l.toZapFields(fields)...)
}

func (l *LoggerAdapter) Debug(msg string, fields watermill.LogFields) {
	logger.Debug(msg, l.toZapFields(fields)...)
}

func (l *LoggerAdapter) Trace(msg string, fields watermill.LogFields) {
	// 注意：zap 默认不提供 Trace 级别，这里我们使用 Debug 级别作为替代
	logger.Debug(msg, l.toZapFields(fields)...)
}

// With 方法用于累积字段
func (l *LoggerAdapter) With(fields watermill.LogFields) watermill.LoggerAdapter {
	newFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		newFields = append(newFields, zap.Any(k, v))
	}

	// 返回一个新的 LoggerAdapter 实例，包含了新累积的字段
	return &LoggerAdapter{
		fields: append(l.fields, newFields...),
	}
}

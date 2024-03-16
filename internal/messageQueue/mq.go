package messageQueue

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-amqp/pkg/amqp"
)

type MQ struct {
	conf   amqp.Config
	logger watermill.LoggerAdapter
}

// NewMQ 新建消息队列
// url 格式: amqp://{user}:{password}@{host}:{port}/{virureHost}
func NewMQ(url string, logger watermill.LoggerAdapter) (mq *MQ, err error) {
	mq = new(MQ)
	mq.conf = amqp.NewDurableQueueConfig(url)
	mq.logger = logger
	return mq, nil
}

func (mq *MQ) NewPublisher() (*amqp.Publisher, error) {
	return amqp.NewPublisher(mq.conf, mq.logger)
}

func (mq *MQ) NewSubscriber() (*amqp.Subscriber, error) {
	return amqp.NewSubscriber(mq.conf, mq.logger)
}

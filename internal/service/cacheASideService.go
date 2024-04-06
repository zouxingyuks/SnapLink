package service

import (
	"SnapLink/internal/cache"
	"SnapLink/internal/message_queue/rabbitmq"
	"SnapLink/internal/model"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/pkg/errors"
	"github.com/zhufuyi/sponge/pkg/app"
	"github.com/zhufuyi/sponge/pkg/logger"
	"go.uber.org/zap"
	"strings"
)

// 本项目中读多写少,适合使用 Cache Aside 策略
// 同时,使用 canal 来去进一步解耦,提高性能

var _ app.IServer = (*CacheASideService)(nil)

const (
	// 同时处理旁路更新的消费者数目
	consumerNumber = 10
	// 处理的 SQL 行为
	insertAction = "insert"
	updateAction = "update"
	deleteAction = "delete"
)

var (
	CacheASideServiceName     = "CacheASideService"
	ErrStartCacheASideService = errors.New("Start CacheASideService...failed")
	ErrStopCacheASideService  = errors.New("Stop CacheASideService...failed")
)

type cacheHandler func(ctx context.Context, action string, m map[string]any) error

var (
	cacheHandlerMap = map[string]cacheHandler{
		model.RedirectPrefix: func(ctx context.Context, action string, m map[string]any) error {
			switch action {
			case updateAction:
				return cache.Redirect().Del(ctx, m["uri"].(string))
			case insertAction, deleteAction:
				{
					//处理对应的 gid 数目更新

					return nil
				}
			default:
				return nil
			}
		},
		model.SLGroupPrefix: func(ctx context.Context, action string, m map[string]any) error {
			switch action {
			case insertAction, updateAction:
				return cache.SLGroup().Del(ctx, m["c_username"].(string))
			default:
				return nil
			}
		},
	}
)

type CacheASideService struct {
	subscriber message.Subscriber
	publisher  message.Publisher
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewCacheASideService 新增旁路缓存更新服务
func NewCacheASideService() app.IServer {
	s := new(CacheASideService)
	s.ctx, s.cancel = context.WithCancel(context.Background())
	var err error
	if s.subscriber, err = rabbitmq.NewSubscriber(); err != nil {
		logger.Panic(errors.Wrap(ErrStartCacheASideService, err.Error()).Error())
		return nil
	}
	return s
}

func (s *CacheASideService) Start() error {

	for i := 0; i < consumerNumber; i++ {
		ch, err := s.subscriber.Subscribe(s.ctx, "maxwell")
		if err != nil {
			return errors.Wrap(ErrStartCacheASideService, err.Error())
		}
		go handlerCacheMessage(s.ctx, ch)
	}
	return nil
}

// 此处处理的是 maxwell 生产的标准信息,非标准信息在错误日志中记录并且将数据从 RabbitMQ 中抛弃,
func handlerCacheMessage(ctx context.Context, ch <-chan *message.Message) {
	for msg := range ch {
		// 解析消息载荷
		var payload map[string]any
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			logger.Error(errors.Wrap(err, "Failed to unmarshal message payload").Error(), zap.Any("msg", msg))
			msg.Ack()
			continue
		}

		// 从载荷中获取表名，并处理
		tableName, ok := payload["table"].(string)
		if !ok || tableName == "" {
			logger.Error(errors.New("Invalid or missing 'table' in payload").Error(), zap.Any("msg", msg))
			msg.Ack()
			continue
		}

		// 找到最后一个下划线的索引
		index := strings.LastIndex(tableName, "-")
		var prefix string
		if index == -1 {
			//没有找到的情况就是说明本表没有进行分库分表
			prefix = tableName
		} else {
			// 获取表名前缀
			prefix = tableName[:index]
		}
		data, ok := payload["data"].(map[string]any)
		if !ok {
			logger.Error(fmt.Sprintf("Invalid or missing 'data' in payload for table: %s", tableName), zap.Any("msg", msg))
			msg.Ack()
			continue
		}
		// 只处理配置的表
		if fn, ok := cacheHandlerMap[prefix]; ok {

			if action, ok := payload["type"].(string); ok {
				// 调用处理函数
				if err := fn(ctx, action, data); err != nil {
					// 此处是因为处理失败,因此不抛弃数据
					logger.Error(fmt.Sprintf("Error handling data for table %s: %v", tableName, err), zap.Any("msg", msg))
					msg.Nack()
					continue
				}
			}
		}
		// 确认消息处理成功
		msg.Ack()
		continue
	}
}

func (s *CacheASideService) Stop() error {
	//关闭相关的 订阅者与发布者
	if err := s.subscriber.Close(); err != nil {
		logger.Panic(errors.Wrap(ErrStopCacheASideService, err.Error()).Error())
	}
	s.cancel()
	return nil
}

func (s *CacheASideService) String() string {
	return CacheASideServiceName
}

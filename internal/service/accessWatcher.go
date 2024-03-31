package service

import (
	"SnapLink/internal/cache"
	"SnapLink/internal/cache/hyperloglog"
	"SnapLink/internal/dao"
	logger2 "SnapLink/internal/logger"
	"SnapLink/internal/messageQueue/rabbitmq"
	"SnapLink/internal/model"
	"SnapLink/pkg/ipSearcher"
	"SnapLink/pkg/userAgent"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/plugin"
	"github.com/pkg/errors"
	"github.com/zhufuyi/sponge/pkg/app"
	"github.com/zhufuyi/sponge/pkg/logger"
	"strings"
	"sync"
	"time"
)

var _ app.IServer = (*watcherService)(nil)

type watcherService struct {
	subscribers []message.Subscriber
	publishers  []message.Publisher
	ctx         context.Context
	cancel      context.CancelFunc
	mutex       sync.Mutex
}

type option struct {
	Name          string                       // 名称
	SubTopic      string                       // 订阅主题
	PubTopic      string                       // 发布主题
	HasPublisher  bool                         // 是否有发布者
	Fn            message.HandlerFunc          // 处理函数
	noPublisherFn message.NoPublishHandlerFunc // 无发布者的处理函数
}

var watcherServiceHandlerMap = []option{
	{
		Name:          "AccessLogHandler",
		SubTopic:      "AccessLog",
		HasPublisher:  false,
		Fn:            nil,
		noPublisherFn: handleAccessLog,
	},
}

func NewWatcherService() app.IServer {
	s := new(watcherService)
	s.ctx, s.cancel = context.WithCancel(context.Background())
	return s
}
func (s *watcherService) Start() error {

	// 启动数据聚合服务
	go batchCreateRecord(s.ctx)
	fn := func(s *watcherService) {
		// 启动消息路由
		router, err := message.NewRouter(message.RouterConfig{}, new(logger2.WatermillAdapter))
		if err != nil {
			logger.Err(errors.Wrap(err, "start accessWatcherService...failed"))
		}
		router.AddPlugin(plugin.SignalsHandler)
		n := len(watcherServiceHandlerMap)
		subscriber, err := rabbitmq.NewSubscriber()
		s.mutex.Lock()
		s.subscribers = append(s.subscribers, subscriber)
		s.mutex.Unlock()
		if err != nil {
			logger.Panic(errors.Wrap(err, "init accessWatcherService...failed").Error())
		}

		Fns := make([]message.HandlerFunc, 0, n)
		NoPublisherFns := make([]message.NoPublishHandlerFunc, 0, n)
		for i := 0; i < n; i++ {
			if watcherServiceHandlerMap[i].HasPublisher {
				Fns = append(Fns, watcherServiceHandlerMap[i].Fn)
			} else {
				NoPublisherFns = append(NoPublisherFns, watcherServiceHandlerMap[i].noPublisherFn)
			}
		}
		// 注册消息处理函数
		for fn := range NoPublisherFns {
			router.AddNoPublisherHandler(watcherServiceHandlerMap[fn].Name, watcherServiceHandlerMap[fn].SubTopic, subscriber, watcherServiceHandlerMap[fn].noPublisherFn)
		}
		if len(Fns) > 0 {
			publisher, err := rabbitmq.NewPublisher()
			if err != nil {
				logger.Panic(errors.Wrap(err, "init accessWatcherService...failed").Error())
			}
			for fn := range Fns {
				router.AddHandler(watcherServiceHandlerMap[fn].Name, watcherServiceHandlerMap[fn].SubTopic, subscriber, watcherServiceHandlerMap[fn].PubTopic, publisher, watcherServiceHandlerMap[fn].Fn)
			}
		}
		if err := router.Run(context.Background()); err != nil {
			logger.Err(errors.Wrap(err, "start accessWatcherService...failed"))
		}
	}
	for i := 1; i <= 100; i++ {
		go fn(s)
	}

	return nil
}

func (s *watcherService) Stop() error {
	for _, subscriber := range s.subscribers {
		if err := subscriber.Close(); err != nil {
			return err
		}
	}
	for _, publisher := range s.publishers {
		if err := publisher.Close(); err != nil {
			return err
		}
	}
	s.cancel()
	return nil
}
func (s *watcherService) String() string {
	return "watcherService"
}

// 统计访问数据
func handleAccessLog(msg *message.Message) (err error) {

	// todo 设定单条信息超时时间

	// 处理消息幂等
	// 消息出现重复消费主要有两种情况：
	// 1. 消息发送失败，重试发送
	// 2. 消费者处理失败，重试消费
	// 此处使用消息 ID 作为唯一标识，处理重复消费问题
	var exist bool
	// 重复消费检查
	for i := 0; i < 3; i++ {
		// 设置消息 ID 的幂等锁
		// 由于单条消息的处理时间可能较长，此处设置的锁时间较长
		exist, err = cache.SetNX(context.Background(), makeIdempotentKey(msg.UUID), 1, time.Minute)
		if err != nil {
			//todo 如果此处的缓存已经失效，应该如何处理
			logger.Err(err)
			return err
		}
		// 成功获取到锁
		if !exist {
			break
		}
		time.Sleep(time.Second)
	}
	if exist {
		// 此消息已经处理过
		return nil
	}
	defer func() {
		// 消息处理失败，删除幂等锁
		if err != nil {
			err := cache.Del(context.Background(), makeIdempotentKey(msg.UUID))
			if err != nil {
				logger.Err(err)
				return
			}
		}
	}()
	// 处理消息
	log := new(rabbitmq.AccessLogMessage)
	err = json.Unmarshal(msg.Payload, log)
	if err != nil {
		logger.Error("handle access log: ", logger.Err(err))
		return err
	}
	t, _ := time.Parse("2006-01-02 15:04:05", log.Datetime)
	record := &model.LinkAccessRecord{
		CreatedAt:   t,
		URI:         log.Info.Uri,
		OriginalURL: log.Info.OriginalURL,
		Gid:         log.Info.Gid,
		IP4:         log.IP,
		RequestID:   log.RequestID,
		Date:        t.Format("2006-01-02"),
		Hour:        t.Hour(),
	}
	// 此处的两个函数涉及到网络请求，可以并行处理
	wg := new(sync.WaitGroup)
	wg.Add(2)
	// IP 解析
	go func(wg *sync.WaitGroup) {
		err := parseIP(record.IP4, record)
		if err != nil {
			logger.Error(errors.Wrap(err, "parse ip error").Error())
		}
		wg.Done()
	}(wg)
	// UA 监控
	go func(wg *sync.WaitGroup) {
		err := parseUA(log.Header.Get("User-Agent"), record)
		if err != nil {
			logger.Error(errors.Wrap(err, "parse user agent error").Error())
		}
		wg.Done()
	}(wg)
	ctx := context.Background()

	// 先写入数据库
	// 基于时间与 URI 再做一次幂等性检查
	errChan := make(chan error, 1)
	recordChan <- &recordChanData{
		record:  record,
		errChan: errChan,
	}
	err = <-errChan

	//d := dao.NewAccessRecord(model.GetDB())
	//err = d.Create(ctx, record)
	//if err != nil {
	//	if dao.DuplicateEntry.Is(err) {
	//		// 重复插入
	//		return nil
	//	}
	//	logger.Error(errors.Wrap(err, "create access record error").Error(), zap.Any("record", record))
	//	return err
	//}
	// 统计访问数据
	uid := log.UID
	{
		// 统计PV
		t, _ := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", record.CreatedAt.String())
		timeKey := t.Format("2006-01-02 15")
		pvKey := fmt.Sprintf("%s:%s:pv", record.URI, timeKey)
		if err := cache.Incr(ctx, pvKey, 1); err != nil {
			return err
		}

		// 统计UV
		//todo 如何确定 uid
		uvKey := fmt.Sprintf("%s:%s:uv", record.URI, timeKey)
		if err := hyperloglog.PFAdd(ctx, uvKey, uid); err != nil {
			return err
		}

		// 统计UIP
		uipKey := fmt.Sprintf("%s:%s:uip", record.URI, timeKey)
		if err := hyperloglog.PFAdd(ctx, uipKey, record.IP4); err != nil {
			return err
		}
	}
	wg.Wait()

	// 更新短链接的统计情况
	return nil
}

const idempotentPrefix = "idempotent:"

func makeIdempotentKey(key string) string {
	return idempotentPrefix + key
}

// 用于聚合数据
type recordChanData struct {
	record  *model.LinkAccessRecord
	errChan chan error
}

var recordChan = make(chan *recordChanData, 100)

// 基于管道来去聚合数据
func batchCreateRecord(ctx context.Context) {
	// 每隔 10s 执行一次批量插入
	// 1. 从管道中获取数据
	// 2. 批量插入数据库
	// 3. 清空管道
	ticker := time.NewTicker(1 * time.Second)
	d := dao.NewAccessRecord(model.GetDB())
	logger.Info("start batch create record")
	defer ticker.Stop()
	for {
		// 执行批量插入
		// 从管道中获取数据
		datas := make([]*recordChanData, 0, 10)
		for {
			select {
			case <-ticker.C:
				{
					goto BatchInsert

				}
			case <-ctx.Done():
				{
					return
				}
			case d := <-recordChan:
				datas = append(datas, d)
				if len(datas) >= 10 {
					goto BatchInsert
				}
			}
		}
	BatchInsert:
		// 批量插入数据库
		l := len(datas)
		records := make([]*model.LinkAccessRecord, 0, l)
		errChans := make([]chan error, 0, len(datas))
		for i := 0; i < l; i++ {
			records = append(records, datas[i].record)
			errChans = append(errChans, datas[i].errChan)
		}
		for {
			if len(records) == 0 {
				break
			}
			i, err := d.CreateBatch(context.Background(), records)
			// 全部插入成功
			if err == nil {
				break
			}
			// 部分插入成功,将插入成功的数据重新插入,并且通知错误
			_, _ = d.CreateBatch(context.Background(), records[:i])
			for _, c := range errChans[:i] {
				c <- nil
			}
			//通知错误
			errChans[i] <- err
			// 重新设置 records 与 errChans
			records = records[i+1:]
			errChans = errChans[i+1:]
		}
		time.Sleep(2 * time.Second)
	}
}

// 解析 UA
func parseUA(ua string, record *model.LinkAccessRecord) error {
	record.UserAgent = ua
	info := userAgent.AutoParse(record.UserAgent)
	record.Browser = info.Browser
	record.Device = info.Device
	record.Network = "unknown"
	return nil
}

// 解析 IP,获取地理位置
func parseIP(ip string, record *model.LinkAccessRecord) error {
	info, err := ipSearcher.IPV4(ip)
	if err != nil {
		return err
	}
	record.Local = strings.Join([]string{info.Country, info.Province, info.City}, ".")
	return nil
}

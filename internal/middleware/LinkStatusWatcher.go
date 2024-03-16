package middleware

import (
	"SnapLink/internal/cache"
	"SnapLink/internal/dao"
	"SnapLink/internal/messageQueue"
	"SnapLink/internal/model"
	"context"
	"github.com/ThreeDotsLabs/watermill-amqp/pkg/amqp"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/zhufuyi/sponge/pkg/logger"
	"sync"
	"time"
)

var weekDay = map[string]int{
	"Monday":    0,
	"Tuesday":   1,
	"Wednesday": 2,
	"Thursday":  3,
	"Friday":    4,
	"Saturday":  5,
	"Sunday":    6,
}

func Watcher() gin.HandlerFunc {
	watcherInstance.Once.Do(func() {
		var err error
		watcherInstance.sDao = dao.NewLinkAccessStatisticDao(
			cache.NewLinkStatsCache(model.GetCacheType()))
		watcherInstance.publisher, err = messageQueue.NewPublisher()
		if err != nil {
			logger.Panic(errors.Wrap(err, "init rocketmq...failed").Error())
		}
	})
	return watcherInstance.Do()
}

var watcherInstance watcher

// LinkStatsDao defining the dao interface
type LinkStatsDao interface {
	UpdateIp(ctx context.Context, uri string, date string, hour int, ip string) error
	UpdateUv(ctx context.Context, uri string, date string, hour int) error
	UpdateUA(ctx context.Context, uri string, date string, hour int, browser, device string) error
	UpdatePv(ctx context.Context, uri string, date string, hour int) error
	SaveAccessRecord(ctx context.Context, record *model.LinkAccessRecord) error
	UpdateLocation(ctx context.Context, uri string, date string, hour int, location string) error
}

type watcher struct {
	sDao LinkStatsDao
	sync.Once
	publisher *amqp.Publisher
}

func (w *watcher) Do() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		// 此处是用于监控短链接的访问情况，其成功访问的状态码为302
		if c.Writer.Status() == 302 {
			//todo 发布到消息队列：AccessLog
			info := c.MustGet("info").(*model.Redirect)
			header := c.Request.Header
			ip := c.RemoteIP()
			uid, err := c.Cookie("uid")
			if err != nil {
				uid = uuid.NewString()

				c.SetCookie("uid", uid, 3600, "/", "", false, false)
			}
			err = w.publisher.Publish("AccessLog", messageQueue.NewAccessLogMessage(*info, header, ip, uid, time.Now().String()))
			if err != nil {
				logger.Err(err)
			}
		}
	}
}

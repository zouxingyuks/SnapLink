package middleware

import (
	"SnapLink/internal/message_queue/rabbitmq"
	"SnapLink/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/zhufuyi/sponge/pkg/logger"
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
	publisher, err := rabbitmq.NewPublisher()
	if err != nil {
		logger.Panic(errors.Wrap(err, "init rabbitmq...failed").Error())
	}
	return func(c *gin.Context) {
		c.Next()
		// 此处是用于监控短链接的访问情况，其成功访问的状态码为302
		if c.Writer.Status() == 302 {
			info := c.MustGet("info").(*model.Redirect)
			header := c.Request.Header
			ip := c.RemoteIP()
			uid, err := c.Cookie("uid")
			if err != nil {
				uid = uuid.NewString()

				c.SetCookie("uid", uid, 3600, "/", "", false, false)
			}
			err = publisher.Publish("accessLog", rabbitmq.NewAccessLogMessage(*info, header, c.GetString("request_id"), ip, uid, time.Now().Format("2006-01-02 15:04:05")))
			if err != nil {
				logger.Err(err)
			}
		}
	}
}

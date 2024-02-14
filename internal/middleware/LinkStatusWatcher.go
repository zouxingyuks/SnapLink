package middleware

import (
	"SnapLink/internal/cache"
	"SnapLink/internal/dao"
	"SnapLink/internal/model"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"net/http"
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
		watcherInstance.iDao = dao.NewLinkStatsDao(
			cache.NewLinkStatsCache(model.GetCacheType()))
	})
	return watcherInstance.Do()
}

var watcherInstance watcher

// LinkStatsDao defining the dao interface
type LinkStatsDao interface {
	Get(ctx context.Context, originalUrl string, date string, hour int) (*model.LinkAccessStat, error)
	Set(ctx context.Context, data *model.LinkAccessStat) error
	UpdateUip(ctx context.Context, originalUrl string, date string, hour int, ip string) error
	UpdateUv(ctx context.Context, originalUrl string, date string, hour int) error
	UpdatePv(ctx context.Context, originalUrl string, date string, hour int) error
}

type watcher struct {
	iDao LinkStatsDao
	sync.Once
}

func (w *watcher) Do() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		// 此处是用于监控短链接的访问情况，其成功访问的状态码为302
		if c.Writer.Status() == 302 {
			var (
				err    error
				status *model.LinkAccessStat
			)
			originalURL := c.Writer.Header().Get("Location")

			// 获取目前的短链接统计情况
			{
				// 通过短链接的uri获取短链接的信息，以小时为单位统计
				status, err = w.iDao.Get(c,
					originalURL,
					time.Now().Format("2006-01-02"),
					weekDay[time.Now().Weekday().String()]*24+(time.Now().Hour()))
				status.Date = time.Now().Format("2006-01-02")
				status.Hour = time.Now().Hour()
				if err != nil {
					// todo 记录日志
				}
			}
			// 更新短链接的统计情况
			// pv监控
			{
				w.iDao.UpdatePv(c, originalURL, status.Date, status.Hour)
			}
			// uv监控
			{
				//根据 uv 的存在性与否来判断是否需要更新 uv
				_, err := c.Cookie("uv")
				if errors.Is(err, http.ErrNoCookie) {
					w.iDao.UpdateUv(c, originalURL, status.Date, status.Hour)
				}
			}
			// ip监控&&地理位置监控
			{
				//获取真实 ip
				//todo 优化此获取方法，目前基于gin自带的获取ip的方法来去获取真实的ip
				ip := c.ClientIP()
				err := w.iDao.UpdateUip(c, originalURL, status.Date, status.Hour, ip)
				if err != nil {
					//如果更新出错，则代表redis出错
				}
				//通过ip获取地理位置
				//locationInfos, err := ipSearcher.Do(ip)
				//exist, err := w.iDao.ExistOrAddLocation(c, originalURL, status.Date, status.Hour, locationInfos["city"])
				//if err != nil {
				//	// todo 记录日志
				//	fmt.Println("获取ip失败", err)
				//}
				//if !exist {
				//	//todo 如何对某个区域进行统计
				//	status.Uip++
				//}
			}
			// UA 监控
			{
				//ua := c.GetHeader("User-Agent")

				//if err != nil {
				//	// todo 记录日志
				//	fmt.Println("获取ua失败", err)
				//}
				//if !exist {
				//	//对 ua 进行统计
				//}
			}
			// 更新短链接的统计情况
			//err = watcherInstance.iDao.Set(c, status)
			fmt.Printf("%+v", status)
		}
	}
}

package routers

import (
	"SnapLink/internal/handler"
	"github.com/gin-gonic/gin"
)

type LinkAccessStatisticHandler interface {
	GetStatistic(c *gin.Context)
	GetRecords(c *gin.Context)
	RefreshStatistic(c *gin.Context)
	GetStatisticByDay(c *gin.Context)
}

func init() {
	apiV1RouterFns = append(apiV1RouterFns, func(group *gin.RouterGroup) {
		linkAccessStatisticRouter(group, handler.NewLinkAccessStatisticHandler())
	})
}
func linkAccessStatisticRouter(group *gin.RouterGroup, h LinkAccessStatisticHandler) {
	group = group.Group("/")
	//获取基础访问统计(PV,UV,UIP)
	group.GET("/stats", h.GetStatistic)
	//获取分组短链接监控
	//group.GET("/stats/group", h.GetStatistic)
	//获取单次访问详情
	group.GET("/stats/access-record", h.GetRecords)
	//立刻更新最新的访问统计数据
	group.POST("/linkAccessStatistic/update", h.RefreshStatistic)
	//获取 day 范围统计结果
	group.GET("/linkAccessStatistic/day", h.GetStatisticByDay)
}

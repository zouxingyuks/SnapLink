package routers

import (
	"SnapLink/internal/handler"
	"github.com/gin-gonic/gin"
)

type LinkAccessStatisticHandler interface {
	GetStatistic(c *gin.Context)
	GetRecords(c *gin.Context)
	RefreshStatistic(c *gin.Context)
}

func init() {
	apiV1RouterFns = append(apiV1RouterFns, func(group *gin.RouterGroup) {
		linkAccessStatisticRouter(group, handler.NewLinkAccessStatisticHandler())
	})
}
func linkAccessStatisticRouter(group *gin.RouterGroup, h LinkAccessStatisticHandler) {
	//获取基础访问统计(PV,UV,UIP)
	group.GET("/linkAccessStatistic", h.GetStatistic)
	//获取单次访问详情
	group.GET("/linkAccessStatistic/detail", h.GetRecords)
	//立刻更新最新的访问统计数据
	group.POST("/linkAccessStatistic/update", h.RefreshStatistic)
}

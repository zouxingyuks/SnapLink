package routers

import (
	"SnapLink/internal/handler"
	"github.com/gin-gonic/gin"
	"github.com/zhufuyi/sponge/pkg/gin/middleware"
)

type FixHandler interface {
	RebulidBF(c *gin.Context)
}

func init() {
	apiV1RouterFns = append(apiV1RouterFns, func(group *gin.RouterGroup) {
		fixRouter(group, handler.NewFixHandler())
	})
}

// fixRouter 修复是在生产环境中出现服务异常时的紧急修复接口
func fixRouter(group *gin.RouterGroup, h FixHandler) {
	group = group.Group("/")
	group.Use(middleware.Auth())
	//重建布隆过滤器
	group.GET("/fix/rebuildbf", h.RebulidBF)
}

package routers

import (
	"SnapLink/internal/handler"
	middleware2 "SnapLink/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/zhufuyi/sponge/pkg/gin/middleware"
)

func init() {
	apiV1RouterFns = append(apiV1RouterFns, func(group *gin.RouterGroup) {
		shortLinkRouter(group, handler.NewShortLinkHandler())
	})
}

func shortLinkRouter(group *gin.RouterGroup, h handler.ShortLinkHandler) {
	group = group.Group("/")
	group.Use(middleware.Auth())
	//创建短链接
	group.POST("/shortlink", middleware2.Sentinel("POST /shortlink"), h.Create)
	//批量创建短链接
	group.POST("/shortlink/batch", h.CreateBatch)
	//更新短链接
	group.PUT("/shortlink", h.Update)
	//分页查询短链接
	group.GET("/shortlink/page", h.List)
	//删除短链接
	group.DELETE("/shortlink/:uri", h.Delete)
}

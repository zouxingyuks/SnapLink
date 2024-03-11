package routers

import (
	"SnapLink/internal/handler"
	"github.com/gin-gonic/gin"
	"github.com/zhufuyi/sponge/pkg/gin/middleware"
)

func init() {
	apiV1RouterFns = append(apiV1RouterFns, func(group *gin.RouterGroup) {
		ShortLinkGroupRouter(group, handler.NewShortLinkGroupHandler())
	})
}
func ShortLinkGroupRouter(group *gin.RouterGroup, h handler.ShortLinkGroupHandler) {
	group = group.Group("/group")
	group.Use(middleware.Auth())

	// 创建短链接分组
	group.POST("/", h.Create)
	group.GET("/", h.List)
	group.PUT("/", h.UpdateByGID)
	group.DELETE("/", h.DelByGID)
	group.POST("/sort", h.UpdateSortOrder)
}

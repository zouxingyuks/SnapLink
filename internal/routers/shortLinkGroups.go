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
	group.Use(middleware.Auth())

	// 创建短链接分组
	group.POST("/group", h.Create)
	group.GET("/group", h.List)
	group.PUT("/group", h.UpdateByGID)
	group.DELETE("/group", h.DelByGID)
	group.POST("group/sort", h.UpdateSortOrder)
}

package routers

import (
	"SnapLink/internal/handler"
	"github.com/zhufuyi/sponge/pkg/gin/middleware"

	"github.com/gin-gonic/gin"
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
	group.POST("/create", h.Create)
	////批量创建短链接
	//group.POST("/create/batch", h.CreateBatch)
	//group.DELETE("/shortLink/:id", h.DeleteByID)
	//group.POST("/shortLink/delete/ids", h.DeleteByIDs)
	//group.PUT("/shortLink/:id", h.UpdateByID)
	//group.GET("/shortLink/:id", h.GetByID)
	//group.POST("/shortLink/condition", h.GetByCondition)
	//group.POST("/shortLink/list/ids", h.ListByIDs)
	//group.GET("/shortLink/list", h.ListByLastID)
	//group.POST("/shortLink/list", h.List)
}

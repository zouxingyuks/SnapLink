package routers

import (
	"SnapLink/internal/handler"
	"SnapLink/internal/middleware"
	"github.com/gin-gonic/gin"
)

type RedirectHandler interface {
	Redirect(c *gin.Context)
}

func init() {
	apiV1RouterFns = append(apiV1RouterFns, func(group *gin.RouterGroup) {
		redirectRouter(group, handler.NewRedirectHandler())
	})
}
func redirectRouter(group *gin.RouterGroup, h RedirectHandler) {
	group.Any("/:uri", middleware.Watcher(), h.Redirect)

}

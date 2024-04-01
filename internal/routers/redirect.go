package routers

import (
	"SnapLink/internal/handler"
	"SnapLink/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/zhufuyi/sponge/pkg/logger"
)

type RedirectHandler interface {
	Redirect(c *gin.Context)
}

func init() {
	apiV1RouterFns = append(apiV1RouterFns, func(group *gin.RouterGroup) {
		h, err := handler.NewRedirectHandler()
		if err != nil {
			logger.Panic(errors.Wrap(err, "init redirectHandler failed").Error())
		}
		redirectRouter(group, h)
	})
}
func redirectRouter(group *gin.RouterGroup, h RedirectHandler) {
	group.Any("/:uri", middleware.Watcher(), h.Redirect)

}

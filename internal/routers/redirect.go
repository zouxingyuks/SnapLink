package routers

import (
	"SnapLink/internal/handler"
	"github.com/gin-gonic/gin"
)

func init() {
	apiV1RouterFns = append(apiV1RouterFns, func(group *gin.RouterGroup) {
		redirectRouter(group, handler.NewRedirectHandler())
	})
}
func redirectRouter(group *gin.RouterGroup, h handler.RedirectHandler) {

	group.GET("/:short_uri", h.Redirect)

}

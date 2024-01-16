package routers

import (
	"SnapLink/internal/handler"
	"github.com/pkg/errors"
	"github.com/zhufuyi/sponge/pkg/gin/middleware"
	"github.com/zhufuyi/sponge/pkg/jwt"

	"github.com/gin-gonic/gin"
)

func init() {
	apiV1RouterFns = append(apiV1RouterFns, func(group *gin.RouterGroup) {
		ShortLinkGroupRouter(group, handler.NewShortLinkGroupHandler())
	})
	//todo 修改此处中间件的验证方式
	jwt.Init()

}
func ShortLinkGroupRouter(group *gin.RouterGroup, h handler.ShortLinkGroupHandler) {
	group.GET("/token", GetToken)
	group.Use(middleware.AuthCustom(verify))
	//complete
	group.POST("/ShortLinkGroup", h.Create)
	group.GET("/ShortLinkGroup/list", h.List)
	group.PUT("/ShortLinkGroup", h.UpdateByGID)

	//incomplete
	group.DELETE("/ShortLinkGroup/:id", h.DeleteByID)
	group.POST("/ShortLinkGroup/delete/ids", h.DeleteByIDs)
	group.GET("/ShortLinkGroup/:id", h.GetByID)
	group.POST("/ShortLinkGroup/condition", h.GetByCondition)
}
func verify(claims *jwt.CustomClaims, tokenTail10 string, c *gin.Context) error {
	//todo 验证用户是否有权限
	if claims.Fields["c_user_id"] != "" {
		c.Set("c_user_id", claims.Fields["c_user_id"])
		return nil

	}
	return errors.New("未登录")
}

func GetToken(c *gin.Context) {
	token, _ := jwt.GenerateCustomToken(map[string]interface{}{
		"c_user_id": "123",
		"role":      "admin",
	})
	c.JSON(200, gin.H{
		"token": token,
	})
}

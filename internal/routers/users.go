package routers

import (
	"SnapLink/internal/handler"

	"github.com/gin-gonic/gin"
)

func init() {
	apiV1RouterFns = append(apiV1RouterFns, func(group *gin.RouterGroup) {
		usersRouter(group, handler.NewUsersHandler())
	})
}

// UsersHandler defining the handler interface
type UsersHandler interface {
	GetByUsername(c *gin.Context)
	GetByUsernameDesensitization(c *gin.Context)
	HasUsername(c *gin.Context)
	Register(c *gin.Context)
	UpdateByUsername(c *gin.Context)
	Login(c *gin.Context)
	CheckLogin(c *gin.Context)
	Logout(c *gin.Context)
}

func usersRouter(group *gin.RouterGroup, h UsersHandler) {
	//group.Use(middleware.Auth()) // all of the following routes use jwt authentication

	//根据用户名查找用户信息
	group.GET("/user/:username", h.GetByUsername)
	//根据用户名查找用户无脱敏信息
	group.GET("/actual/user/:username", h.GetByUsernameDesensitization)
	//查询用户名是否可用
	group.GET("/user/has-username", h.HasUsername)
	//注册用户
	group.POST("/user", h.Register)
	//修改用户
	group.PUT("/user", h.UpdateByUsername)
	//用户登录
	group.POST("/user/login", h.Login)
	//检查用户是否登录
	group.GET("/user/check-login", h.CheckLogin)
	//用户登出
	group.DELETE("/user/logout", h.Logout)
}

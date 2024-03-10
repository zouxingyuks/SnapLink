package handler

import (
	"SnapLink/internal/cache"
	"SnapLink/internal/dao"
	"SnapLink/internal/ecode"
	"SnapLink/internal/model"
	"SnapLink/internal/types"
	"SnapLink/pkg/serialize"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/zhufuyi/sponge/pkg/gin/middleware"
)

type UsersHandler struct {
	iDao dao.TUserDao
}

// NewUsersHandler creating the handler interface
func NewUsersHandler() (h *UsersHandler) {
	h = &UsersHandler{
		iDao: dao.NewTUserDao(
			model.GetDB(),
			cache.NewTUserCache(model.GetCacheType()),
		),
	}
	h.makeUsernameBF()
	return
}

func (h *UsersHandler) makeUsernameBF() {
	//todo 此处改为远程配置
	err := cache.Create(context.Background(), "username", 0.001, 1e9)
	// 如果创建失败，说明已经存在，正常结束即可
	// 如果创建成功，说明不存在，需要添加数据
	if err == nil {
		//todo 后面此处尝试改为使用 消息队列配合binlog 来去同步数据

		// 从数据库中获取所有的用户名
		usernames, err := h.iDao.GetAllUserName(context.Background())
		if err != nil {
			panic(errors.Wrap(err, "get all username from db error"))
		}
		// 将所有的用户名添加到布隆过滤器中
		err = cache.MAdd(context.Background(), "username", usernames...)
	}
}

func getTUserUsernameFromPath(c *gin.Context) string {
	username := c.Param("username")
	return username
}

// GetByUsername 根据用户名查找用户信息
// @Summary 根据用户名查找用户信息
// @Description 根据用户名查找用户信息
// @Tags users
// @Accept application/json
// @Produce application/json
// @Param username path string true "用户名"
// todo 此接口需要权限认证，且需要高级权限认证
func (h *UsersHandler) GetByUsername(c *gin.Context) {
	username := getTUserUsernameFromPath(c)
	if username == "" {
		serialize.NewResponseWithErrCode(ecode.ClientError, serialize.WithErr(errors.New("username is null"))).ToJSON(c)
		return

	}
	ctx := middleware.WrapCtx(c)
	user, err := h.iDao.GetByUsername(ctx, username)
	if err != nil {
		serialize.NewResponseWithErrCode(ecode.UserNotExistError, serialize.WithErr(err)).ToJSON(c)
		return
	}

	serialize.NewResponse(200, serialize.WithData(user)).ToJSON(c)
}

// GetByUsernameDesensitization 根据用户名查找用户信息(脱敏)
// @Summary 根据用户名查找用户信息(脱敏)
// @Description 根据用户名查找用户信息(脱敏)
// @Tags users
// @Accept application/json
// @Produce application/json
// @Param username path string true "用户名"
// todo 数据脱敏
func (h *UsersHandler) GetByUsernameDesensitization(c *gin.Context) {
	username := getTUserUsernameFromPath(c)
	if username == "" {
		serialize.NewResponseWithErrCode(ecode.ClientError, serialize.WithErr(errors.New("username is null"))).ToJSON(c)
		return

	}
	ctx := middleware.WrapCtx(c)
	user, err := h.iDao.GetByUsername(ctx, username)
	if err != nil {
		serialize.NewResponseWithErrCode(ecode.UserNotExistError, serialize.WithErr(err)).ToJSON(c)
		return
	}
	//数据脱敏
	userDesensitization := types.GetByUsernameDesensitizationRespond{
		Username: user.Username,
		RealName: user.RealName,
		Phone:    user.Phone,
		Mail:     user.Mail,
	}
	serialize.NewResponse(200, serialize.WithData(userDesensitization)).ToJSON(c)
}

// HasUsername 判断用户名是否存在
// @Summary 判断用户名是否存在
// @Description 判断用户名是否存在
// @Tags users
// @Accept application/json
// @Produce application/json
// @Param username query string true "用户名"
func (h *UsersHandler) HasUsername(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		serialize.NewResponseWithErrCode(ecode.ClientError, serialize.WithErr(errors.New("username is null"))).ToJSON(c)
		return
	}

	ctx := middleware.WrapCtx(c)
	has, err := h.iDao.HasUsername(ctx, username)
	result := gin.H{
		"success": !has,
	}
	if err != nil {
		serialize.NewResponseWithErrCode(ecode.ServiceError, serialize.WithData(result), serialize.WithErr(err)).ToJSON(c)
		return
	}
	serialize.NewResponse(200, serialize.WithData(result)).ToJSON(c)
}

// Register 用户注册
// @Summary 用户注册
// @Description 用户注册
// @Tags users
// @Accept application/json
// @Produce application/json
// @Param username body string true "用户名"
// @Param password body string true "密码"
// @Param realName body string true "真实姓名"
// @Param phone body string true "手机号"
// @Param mail body string true "邮箱"
func (h *UsersHandler) Register(c *gin.Context) {
	form := new(types.RegisterRequest)
	if err := c.ShouldBindJSON(form); err != nil {
		serialize.NewResponseWithErrCode(ecode.ClientError, serialize.WithErr(err)).ToJSON(c)
		return
	}
	//1. 检测用户名是否合法
	ctx := middleware.WrapCtx(c)
	has, err := h.iDao.HasUsername(ctx, form.Username)
	if err != nil {
		serialize.NewResponseWithErrCode(ecode.ServiceError, serialize.WithErr(err)).ToJSON(c)
		return
	}
	if has {
		serialize.NewResponseWithErrCode(ecode.UserNameExistError, serialize.WithErr(errors.New("username exist"))).ToJSON(c)
		return
	}
	//todo 剩余的信息校验
	//2. 检测真实姓名是否合法
	//3. 检测密码是否合法
	//4. 检测手机号是否合法
	//5. 检测邮箱是否合法
	//todo 注册信息检测:为这几个字段添加布隆过滤器检测

	u := new(model.TUser)
	u.Username = form.Username
	//todo 密码加密
	u.Password = form.Password
	u.RealName = form.RealName
	u.Phone = form.Phone
	u.Mail = form.Mail

	//6. 注册用户
	err = h.iDao.Create(ctx, u)
	if err != nil {
		serialize.NewResponseWithErrCode(ecode.ServiceError, serialize.WithErr(err)).ToJSON(c)
		return
	}
	serialize.NewResponse(200, serialize.WithData(types.RegisterRespond{
		Username: u.Username,
		RealName: u.RealName,
		Phone:    u.Phone,
		Mail:     u.Mail,
	})).ToJSON(c)
}

func (h *UsersHandler) UpdateByUsername(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户登录
// @Tags users
// @Accept application/json
// @Produce application/json
// @Param token header string true "token"
// @Param username body string true "用户名"
// @Param password body string true "密码"
func (h *UsersHandler) Login(c *gin.Context) {
	//TODO implement me
	c.JSON(200, gin.H{})
}

func (h *UsersHandler) CheckLogin(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (h *UsersHandler) Logout(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

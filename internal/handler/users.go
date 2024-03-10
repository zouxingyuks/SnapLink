package handler

import (
	"SnapLink/internal/cache"
	"SnapLink/internal/dao"
	"SnapLink/internal/ecode"
	"SnapLink/internal/model"
	"SnapLink/internal/types"
	"SnapLink/internal/utils"
	"SnapLink/pkg/serialize"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/zhufuyi/sponge/pkg/gin/middleware"
	"github.com/zhufuyi/sponge/pkg/jwt"
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
	err := cache.BFCreate(context.Background(), "username", 0.001, 1e9)
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
		err = cache.BFMAdd(context.Background(), "username", usernames...)
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
// @Param phone body string true "手机号,e164格式"
// @Param mail body string true "邮箱"
func (h *UsersHandler) Register(c *gin.Context) {
	form := new(types.RegisterRequest)
	if err := c.ShouldBindJSON(form); err != nil {
		serialize.NewResponseWithErrCode(ecode.ClientError, serialize.WithErr(err)).ToJSON(c)
		return
	}
	//1. 检测用户名是否存在
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
	//todo 注册信息检测:为这几个字段添加布隆过滤器检测
	u := &model.TUser{
		Username: form.Username,
		Password: utils.Encrypt(form.Password),
		RealName: form.RealName,
		Phone:    form.Phone,
		Mail:     form.Mail,
	}

	//6. 注册用户
	err = h.iDao.Create(ctx, u)
	if err != nil {
		//布隆过滤器的漏网之鱼

		if dao.DuplicateEntry.Is(err) {
			serialize.NewResponseWithErrCode(ecode.UserNameExistError, serialize.WithErr(err)).ToJSON(c)
			cache.BFAdd(ctx, "username", u.Username)
			return
		}
		serialize.NewResponseWithErrCode(ecode.ServiceError, serialize.WithErr(err)).ToJSON(c)
		return
	}
	//7. 加入布隆过滤器
	err = cache.BFAdd(ctx, "username", u.Username)
	if err != nil {
		serialize.NewResponseWithErrCode(ecode.ServiceError, serialize.WithErr(err)).ToJSON(c)
		return
	}
	//返回注册信息
	serialize.NewResponse(200, serialize.WithData(types.RegisterRespond{
		Username: u.Username,
		RealName: u.RealName,
		Phone:    u.Phone,
		Mail:     u.Mail,
	})).ToJSON(c)
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户登录
// @Tags users
// @Accept application/json
// @Produce application/json
// @Param Authorization header string false "token"
// @Param username body string true "用户名"
// @Param password body string true "密码"
func (h *UsersHandler) Login(c *gin.Context) {
	form := new(types.LoginRequest)
	if err := c.ShouldBindJSON(form); err != nil {
		serialize.NewResponseWithErrCode(ecode.ClientError, serialize.WithErr(err)).ToJSON(c)
		return
	}
	//1. 检测用户名是否存在
	ctx := middleware.WrapCtx(c)
	has, err := h.iDao.HasUsername(ctx, form.Username)
	if err != nil {
		serialize.NewResponseWithErrCode(ecode.ServiceError, serialize.WithErr(err)).ToJSON(c)
		return
	}
	if !has {
		serialize.NewResponseWithErrCode(ecode.UserNotExistError, serialize.WithErr(errors.New("username not exist"))).ToJSON(c)
		return
	}
	//2. 检查用户是否登录
	token := c.Request.Header.Get("Authorization")
	if token != "" {
		c.Get("uid")
		//todo 如何更优雅的调用
	}
	//2. 检测密码是否正确
	user, err := h.iDao.GetByUsername(ctx, form.Username)
	if err != nil {
		serialize.NewResponseWithErrCode(ecode.ServiceError, serialize.WithErr(err)).ToJSON(c)
		return
	}
	err = utils.Compare(user.Password, form.Password)
	if err != nil {
		serialize.NewResponseWithErrCode(ecode.PasswordVerifyError, serialize.WithErr(err)).ToJSON(c)
		return
	}
	//3. 生成token
	token, err = jwt.GenerateToken(user.Username, "admin")
	if err != nil {
		serialize.NewResponseWithErrCode(ecode.ServiceError, serialize.WithErr(errors.Wrap(err, "generate token error"))).ToJSON(c)
		return
	}
	//4. 返回token
	serialize.NewResponse(200, serialize.WithData(gin.H{
		"token": token,
	})).ToJSON(c)
	return
}

// UpdateInfo 根据用户名更新用户信息
// @Summary 根据用户名更新用户信息
// @Description 根据用户名更新用户信息
// @Tags users
// @Accept application/json
// @Produce application/json
// @Param token header string true "token"
// @Param password body string true "密码"
// @Param realName body string true "真实姓名"
// @Param phone body string true "手机号"
// @Param mail body string true "邮箱"
func (h *UsersHandler) UpdateInfo(c *gin.Context) {
	//能到这步说明token已经验证通过
	claims, _ := jwt.ParseToken(c.GetHeader("Authorization")[7:])

	form := new(types.UpdateInfoRequest)
	if err := c.ShouldBind(form); err != nil {
		serialize.NewResponseWithErrCode(ecode.ClientError, serialize.WithErr(err)).ToJSON(c)
		return
	}
	// 此处的数据需要手工校验合法性
	if form.Password != "" {
		//todo 重新设定密码的合法性校验
		if !utils.LengthCheck(form.Password, 6, 15) || !utils.InvalidCharCheck(form.Password) {
			serialize.NewResponseWithErrCode(ecode.ClientError, serialize.WithErr(errors.New("password length error"))).ToJSON(c)
			return
		}
	}
	if form.RealName != "" {
		if !utils.LengthCheck(form.RealName, 2, 10) {
			serialize.NewResponseWithErrCode(ecode.ClientError, serialize.WithErr(errors.New("realName length error"))).ToJSON(c)
			return
		}
	}
	if form.Phone != "" {
		if !utils.IsPhone(form.Phone) {
			serialize.NewResponseWithErrCode(ecode.ClientError, serialize.WithErr(errors.New("phone format error"))).ToJSON(c)
			return
		}
	}
	if form.Mail != "" {
		if !utils.IsEmail(form.Mail) {
			serialize.NewResponseWithErrCode(ecode.ClientError, serialize.WithErr(errors.New("mail format error"))).ToJSON(c)
			return
		}
	}
	ctx := middleware.WrapCtx(c)
	user := &model.TUser{
		Username: claims.UID,
		RealName: form.RealName,
		Phone:    form.Phone,
		Mail:     form.Mail,
	}
	fmt.Println(*user)

	if form.Password != "" {
		user.Password = utils.Encrypt(form.Password)
	}
	err := h.iDao.Update(ctx, user)
	if err != nil {
		serialize.NewResponseWithErrCode(ecode.ServiceError, serialize.WithErr(err)).ToJSON(c)
		return
	}
	response := types.UpdateInfoRespond{
		Username: user.Username,
		RealName: user.RealName,
		Phone:    user.Phone,
		Mail:     user.Mail,
	}
	fmt.Println(response)
	serialize.NewResponse(200, serialize.WithData(response)).ToJSON(c)
}

// CheckLogin 检查用户是否登录
// @Summary 检查用户是否登录
// @Description 检查用户是否登录
// @Tags users
// @Accept application/json
// @Produce application/json
// @Param token header string true "token"
// @Success 200 {object} string "ok"
func (h *UsersHandler) CheckLogin(c *gin.Context) {
	serialize.NewResponse(200, serialize.WithData(true)).ToJSON(c)
}

func (h *UsersHandler) Logout(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

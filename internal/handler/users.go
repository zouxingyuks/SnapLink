package handler

import (
	"SnapLink/internal/custom_err"
	"SnapLink/internal/dao"
	"SnapLink/internal/ecode"
	"SnapLink/internal/model"
	"SnapLink/internal/types"
	"SnapLink/internal/utils"
	"SnapLink/pkg/serialize"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/zhufuyi/sponge/pkg/gin/middleware"
	"github.com/zhufuyi/sponge/pkg/jwt"
	"regexp"
)

var phoneRegexp = `^\+[1-9]\d{1,14}$`

type UsersHandler struct {
}

// NewUsersHandler creating the handler interface
func NewUsersHandler() (h *UsersHandler) {
	h = new(UsersHandler)
	return
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
	user, err := dao.TUserDao().GetByUsername(ctx, username)
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
	user, err := dao.TUserDao().GetByUsername(ctx, username)
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
	has, err := dao.TUserDao().HasUsername(ctx, username)
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
	//手工校验手机号格式合法性
	if ok, err := regexp.Match(phoneRegexp, []byte(form.Phone)); !ok || err != nil {
		serialize.NewResponseWithErrCode(ecode.ClientError, serialize.WithErr(errors.New("phone format error"))).ToJSON(c)
		return
	}
	//1. 检测用户名是否存在
	ctx := middleware.WrapCtx(c)
	has, err := dao.TUserDao().HasUsername(ctx, form.Username)
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
	err = dao.TUserDao().Create(ctx, u)
	if err != nil {
		if custom_err.ErrDuplicateEntry.Is(err) {
			serialize.NewResponseWithErrCode(ecode.UserNameExistError, serialize.WithErr(err)).ToJSON(c)
			return
		}
		serialize.NewResponseWithErrCode(ecode.ServiceError, serialize.WithErr(err)).ToJSON(c)
		return
	}
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
	has, err := dao.TUserDao().HasUsername(ctx, form.Username)
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
	user, err := dao.TUserDao().GetByUsername(ctx, form.Username)
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
	err := dao.TUserDao().Update(ctx, user)
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

// Logout
// @Summary 用户登出
// @Description 用户登出
// @Tags users
// @Accept application/json
// @Produce application/json
func (h *UsersHandler) Logout(c *gin.Context) {
	//todo 此处需要将token加入黑名单
	c.JSON(200, "ok")
}

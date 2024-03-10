package handler

import (
	"SnapLink/internal/cache"
	"SnapLink/internal/dao"
	"SnapLink/internal/ecode"
	"SnapLink/internal/model"
	"SnapLink/internal/types"
	"SnapLink/pkg/serialize"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/zhufuyi/sponge/pkg/gin/middleware"
	"github.com/zhufuyi/sponge/pkg/jwt"
)

var _ ShortLinkGroupHandler = (*shortLinkGroupsHandler)(nil)

// ShortLinkGroupHandler defining the handler interface
type ShortLinkGroupHandler interface {
	Create(c *gin.Context)
	List(c *gin.Context)
}

type shortLinkGroupsHandler struct {
	iDao dao.ShortLinkGroupDao
}

// NewShortLinkGroupHandler creating the handler interface
func NewShortLinkGroupHandler() ShortLinkGroupHandler {
	return &shortLinkGroupsHandler{
		iDao: dao.NewShortLinkGroupDao(
			model.GetDB(),
			cache.NewShortLinkGroupCache(model.GetCacheType()),
		),
	}
}

// Create  创建短链接分组
// @Summary 创建短链接分组
// @Description 创建短链接分组
// @Tags shortLinkGroup
// @Accept application/json
// @Produce application/json
// @param Authorization header string true "token"
// @Param enable body int false "是否启用"
// @Param name body string true "分组名称"
// @Param description body string false "描述"
// @Success 200 {object} types.CreateShortLinkGroupRespond{}
// @Failure 400 string "{"msg": "参数错误"}"
// @RedirectInfo /api/v1/slink/group [post]
func (h *shortLinkGroupsHandler) Create(c *gin.Context) {
	param := new(types.CreateShortLinkGroupRequest)

	if err := c.ShouldBindJSON(param); err != nil {
		serialize.NewResponseWithErrCode(ecode.RequestParamError, serialize.WithErr(err)).ToJSON(c)
		return
	}
	claims, _ := jwt.ParseToken(c.GetHeader("Authorization")[7:])
	username := claims.UID

	// 2.参数校验
	group := &model.ShortLinkGroup{
		Gid:       uuid.NewString(),
		Name:      param.Name,
		CUsername: username,
	}
	ctx := middleware.WrapCtx(c)
	err := h.iDao.Create(ctx, group)
	if err != nil {
		serialize.NewResponseWithErrCode(ecode.ServiceError, serialize.WithErr(err)).ToJSON(c)
		return
	}
	serialize.NewResponse(200).ToJSON(c)
}

// List 列出所有短链接分组
// @Summary 列出所有短链接分组，支持分页和条件查询，不传参数则返回前100条
// @Description 列出所有短链接分组，支持分页和条件查询，不传参数则返回前100条
// @Tags shortLinkGroup
// @Accept application/json
// @Produce application/json
// @param Authorization header string true "token"
// @Success 200 {object} types.ListShortLinkGroupRespond{}
// @Failure 400 string "{"msg": "参数错误"}"
// @RedirectInfo /api/v1/slink/group/list [get]
func (h *shortLinkGroupsHandler) List(c *gin.Context) {
	//1. 参数解析
	claims, _ := jwt.ParseToken(c.GetHeader("Authorization")[7:])
	username := claims.UID
	ctx := middleware.WrapCtx(c)
	groups, err := h.iDao.GetAllByCUser(ctx, username)
	if err != nil {
		serialize.NewResponseWithErrCode(ecode.ServiceError, serialize.WithErr(err)).ToJSON(c)
		return
	}
	serialize.NewResponse(200, serialize.WithData(groups)).ToJSON(c)
}

// UpdateByGID 根据gid更新短链接分组
// @Summary 根据gid更新短链接分组
// @Description 根据gid更新短链接分组，需要登录
// @Tags shortLinkGroup
// @Accept application/json
// @Produce application/json
// @Param Authorization header string true "token"
// @Param gid body string true "gid"
// @Param name body string false "name"
// @Param description body string false "description"
// @Success 200 {object} types.UpdateShortLinkGroupByIDRespond{}
// @Failure 400 string "{"msg": "参数错误"}"
// @Failure 404 string "{"msg": "未找到该记录"}"
// @Failure 500 string "{"msg": "更新失败"}"
// @RedirectInfo /api/v1/slink/group [put]
func (h *shortLinkGroupsHandler) UpdateByGID(c *gin.Context) {
	form := new(types.UpdateShortLinkGroupByGIDRequest)

	if err := c.ShouldBind(form); err != nil {
		serialize.NewResponseWithErrCode(ecode.RequestParamError, serialize.WithErr(err)).ToJSON(c)
		return
	}

	//shortLinkGroup := &model.ShortLinkGroup{
	//	Gid: form.Gid,
	//}
	//userid, _ := c.Get("c_user_id")
	//shortLinkGroup.CUserId = userid.(string)
	//if form.Name != "" {
	//	_, yes := sercurity.CleanXSS(form.Name)
	//	if yes {
	//		serialize.NewResponse(400, serialize.WithMsg("参数错误"), serialize.WithErr(errors.New("name 不合法"))).ToJSON(c)
	//		return
	//	}
	//	shortLinkGroup.Name = form.Name
	//
	//}
	//if form.Description != "" {
	//	_, yes := sercurity.CleanXSS(form.Description)
	//	if yes {
	//		serialize.NewResponse(400, serialize.WithMsg("参数错误"), serialize.WithErr(errors.New("description 不合法"))).ToJSON(c)
	//		return
	//	}
	//	shortLinkGroup.Description = form.Description
	//}
	//
	//ctx := middleware.WrapCtx(c)
	//err := h.iDao.UpdateByGidAndCUserId(ctx, shortLinkGroup)
	//if err != nil {
	//	err = errors.Wrap(err, "写入数据库时失败")
	//	logger.Error("UpdateByGID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
	//	serialize.NewResponse(500, serialize.WithMsg("更新失败"), serialize.WithErr(err)).ToJSON(c)
	//	return
	//}
	//logger.Info("更新成功", logger.Any("shortLinkGroup", shortLinkGroup), middleware.GCtxRequestIDField(c))
	//serialize.NewResponse(200, serialize.WithMsg("更新成功")).ToJSON(c)
	panic("implement me")
}

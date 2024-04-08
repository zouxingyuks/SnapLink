package handler

import (
	"SnapLink/internal/dao"
	"SnapLink/internal/ecode"
	"SnapLink/internal/model"
	"SnapLink/internal/types"
	"SnapLink/pkg/serialize"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/zhufuyi/sponge/pkg/gin/middleware"
	"github.com/zhufuyi/sponge/pkg/jwt"
)

var _ ShortLinkGroupHandler = (*shortLinkGroupsHandler)(nil)

// ShortLinkGroupHandler defining the handler interface
type ShortLinkGroupHandler interface {
	Create(c *gin.Context)
	List(c *gin.Context)
	UpdateByGID(c *gin.Context)
	DelByGID(c *gin.Context)
	UpdateSortOrder(c *gin.Context)
}

type shortLinkGroupsHandler struct {
	iDao dao.ShortLinkGroupDao
}

// NewShortLinkGroupHandler creating the handler interface
func NewShortLinkGroupHandler() ShortLinkGroupHandler {
	return &shortLinkGroupsHandler{
		iDao: dao.NewShortLinkGroupDao(
			model.GetDB(),
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
// @Redirect /api/v1/slink/group [post]
func (h *shortLinkGroupsHandler) Create(c *gin.Context) {
	param := new(types.ShortLinkGroupCreateReq)

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
// @Redirect /api/v1/slink/group/list [get]
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
	res := make([]*types.ShortLinkGroupListItem, 0, len(groups))
	for _, group := range groups {
		count, err := dao.ShortLinkDao().Count(ctx, group.Gid)
		if err != nil {
			serialize.NewResponseWithErrCode(ecode.ServiceError, serialize.WithErr(err)).ToJSON(c)
			return
		}
		res = append(res, types.NewShortLinkGroupListItem(map[string]any{
			"group": group,
			"count": count,
		}))
	}
	serialize.NewResponse(200, serialize.WithData(res)).ToJSON(c)
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
// @Redirect /api/v1/slink/group [put]
func (h *shortLinkGroupsHandler) UpdateByGID(c *gin.Context) {
	req := new(types.ShortLinkGroupUpdateByGIDReq)

	if err := c.ShouldBind(req); err != nil {
		serialize.NewResponseWithErrCode(ecode.RequestParamError, serialize.WithErr(err)).ToJSON(c)
		return
	}

	claims, _ := jwt.ParseToken(c.GetHeader("Authorization")[7:])
	username := claims.UID
	ctx := middleware.WrapCtx(c)
	group, err := h.iDao.UpdateByGidAndUsername(ctx, req.Gid, req.Name, username)
	if err != nil {
		serialize.NewResponseWithErrCode(ecode.ServiceError, serialize.WithErr(err)).ToJSON(c)
		return
	}
	serialize.NewResponse(200, serialize.WithData(group)).ToJSON(c)
}

// DelByGID 根据 gid 删除对应的短链接分组
// @Summary 根据 gid 删除对应的短链接分组
// @Description 根据 gid 删除对应的短链接分组
// @Tags shortLinkGroup
// @Accept application/json
// @Produce application/json
// @Param Authorization header string true "token"
// @Param gid body string true "gid"
func (h *shortLinkGroupsHandler) DelByGID(c *gin.Context) {
	gid := c.Query("gid")
	if gid == "" {
		serialize.NewResponseWithErrCode(ecode.RequestParamError, serialize.WithErr(errors.New("gid is empty"))).ToJSON(c)
		return
	}
	claims, _ := jwt.ParseToken(c.GetHeader("Authorization")[7:])
	username := claims.UID
	ctx := middleware.WrapCtx(c)
	err := h.iDao.DelByGidAndUsername(ctx, gid, username)
	if err != nil {
		serialize.NewResponseWithErrCode(ecode.ServiceError, serialize.WithErr(err)).ToJSON(c)
		return
	}
	serialize.NewResponse(200).ToJSON(c)
}

// UpdateSortOrder 更新排序
// @Summary 更新排序
// @Description 更新排序
// @Tags shortLinkGroup
// @Accept application/json
// @Produce application/json
// @Param Authorization header string true "token"
// @Param sort_order body int true "排序标识"
// @Success 200 {object} types.UpdateSortOrderRespond{}
func (h *shortLinkGroupsHandler) UpdateSortOrder(c *gin.Context) {
	form := make([]types.ShortLinkGroupUpdateSortOrderReq, 0)
	if err := c.ShouldBind(&form); err != nil {
		serialize.NewResponseWithErrCode(ecode.RequestParamError, serialize.WithErr(err)).ToJSON(c)
		return
	}
	claims, _ := jwt.ParseToken(c.GetHeader("Authorization")[7:])
	username := claims.UID
	n := len(form)
	gids, sortOrders := make([]string, n), make([]int, n)
	for i, v := range form {
		gids[i] = v.Gid
		sortOrders[i] = v.SortOrder
	}
	ctx := middleware.WrapCtx(c)
	err := h.iDao.UpdateSortOrderByGidAndUsername(ctx, gids, sortOrders, username)
	if err != nil {
		serialize.NewResponseWithErrCode(ecode.ServiceError, serialize.WithErr(err)).ToJSON(c)
		return
	}
	serialize.NewResponse(200).ToJSON(c)

}

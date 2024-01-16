package handler

import (
	"SnapLink/internal/cache"
	"SnapLink/internal/dao"
	"SnapLink/internal/ecode"
	"SnapLink/internal/model"
	"SnapLink/internal/types"
	"SnapLink/pkg/sercurity"
	"SnapLink/pkg/serialize"
	"github.com/pkg/errors"
	"github.com/zhufuyi/sponge/pkg/mysql/query"
	"strconv"
	"strings"

	"github.com/zhufuyi/sponge/pkg/gin/middleware"
	"github.com/zhufuyi/sponge/pkg/gin/response"
	"github.com/zhufuyi/sponge/pkg/logger"
	"github.com/zhufuyi/sponge/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
)

var _ ShortLinkGroupHandler = (*shortLinkGroupsHandler)(nil)

// ShortLinkGroupHandler defining the handler interface
type ShortLinkGroupHandler interface {
	Create(c *gin.Context)
	DeleteByID(c *gin.Context)
	DeleteByIDs(c *gin.Context)
	UpdateByGID(c *gin.Context)
	GetByID(c *gin.Context)
	GetByCondition(c *gin.Context)
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
// @Param description body string true "描述"
// @Param enable body int true "是否启用"
// @Param name body string true "分组名称"
// @Param cUserId body string true "创建人"
// @Success 200 {object} types.CreateShortLinkGroupRespond{}
// @Failure 400 string "{"msg": "参数错误"}"
// @Router /api/v1/shortLinkGroups [post]
func (h *shortLinkGroupsHandler) Create(c *gin.Context) {
	param := &types.CreateShortLinkGroupRequest{}
	err := c.ShouldBindJSON(param)
	res := serialize.NewResponse(400, serialize.WithMsg("参数错误"))
	if err != nil {
		res.ToJSON(c)
		return
	}
	shortLinkGroup := &model.ShortLinkGroup{
		Description: param.Description,
		Enable:      param.Enable,
		Name:        param.Name,
		CUserId:     param.CUserId,
	}

	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, shortLinkGroup)
	if err != nil {
		err = errors.Wrap(err, "创建短链接分组失败")
		logger.Error("创建短链接分组失败", logger.Err(err), logger.Any("shortLinkGroup", shortLinkGroup), middleware.GCtxRequestIDField(c))
		serialize.NewResponse(500, serialize.WithMsg("创建短链接分组失败"), serialize.WithErr(err)).ToJSON(c)
		return
	}
	serialize.NewResponse(200, serialize.WithMsg("创建短链接分组成功")).ToJSON(c)
}

// List 列出所有短链接分组
// @Summary 列出所有短链接分组，支持分页和条件查询，不传参数则返回前100条
// @Description 列出所有短链接分组，支持分页和条件查询，不传参数则返回前100条
// @Tags shortLinkGroup
// @Accept application/json
// @Produce application/json
// @Param page body int false "页码"
// @Param size body int false "每页数量"
// @Param sort body string false "排序字段"
// @Success 200 {object} types.ListShortLinkGroupRespond{}
// @Failure 400 string "{"msg": "参数错误"}"
// @Router /api/v1/ShortLinkGroup/list [get]
func (h *shortLinkGroupsHandler) List(c *gin.Context) {
	//1. 参数解析
	form := &types.ListShortLinkGroupRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		serialize.NewResponse(400, serialize.WithMsg("参数错误"), serialize.WithErr(err)).ToJSON(c)
		return
	}
	//2. 参数校验
	CUserId, exist := c.Get("c_user_id")
	if CUserId == "" || !exist {
		serialize.NewResponse(400, serialize.WithMsg("参数错误"), serialize.WithErr(errors.New("用户未登录"))).ToJSON(c)
	}
	// 默认值
	if (form.Page+1)*(form.Size) == 0 {
		form.Page = 0
		form.Size = 100
	}
	param := query.Params{
		Page: form.Page,
		Size: form.Size,
		Sort: form.Sort,
		Columns: []query.Column{
			{
				Name:  "c_user_id",
				Value: CUserId,
				Exp:   "like",
			},
		},
	}
	ctx := middleware.WrapCtx(c)
	shortLinkGroups, total, err := h.iDao.GetByColumns(ctx, &param)
	if err != nil {
		logger.Error("GetByColumns error", logger.Err(err), logger.Any("param", param), middleware.GCtxRequestIDField(c))
		serialize.NewResponse(500, serialize.WithMsg("查询失败"), serialize.WithErr(err)).ToJSON(c)
		return
	}
	logger.Info("查询成功", logger.Any("shortLinkGroups", shortLinkGroups), logger.Int64("total", total), middleware.GCtxRequestIDField(c))
	serialize.NewResponse(200, serialize.WithMsg("查询成功"), serialize.WithData(gin.H{
		"shortLinkGroups": shortLinkGroups,
		"total":           total,
	})).ToJSON(c)
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
// @Router /api/v1/shortLinkGroups [put]
func (h *shortLinkGroupsHandler) UpdateByGID(c *gin.Context) {
	form := &types.UpdateShortLinkGroupByIDRequest{}
	// 1.参数绑定
	if !parseParams(c, form) {
		return
	}
	// 2.参数校验
	if form.Gid < 0 {
		serialize.NewResponse(400, serialize.WithMsg("参数错误"), serialize.WithErr(errors.New("gid 不合法"))).ToJSON(c)
		return
	}
	shortLinkGroup := &model.ShortLinkGroup{
		Gid: form.Gid,
	}
	userid, _ := c.Get("c_user_id")
	shortLinkGroup.CUserId = userid.(string)
	if form.Name != "" {
		_, yes := sercurity.CleanXSS(form.Name)
		if yes {
			serialize.NewResponse(400, serialize.WithMsg("参数错误"), serialize.WithErr(errors.New("name 不合法"))).ToJSON(c)
			return
		}
		shortLinkGroup.Name = form.Name

	}
	if form.Description != "" {
		_, yes := sercurity.CleanXSS(form.Description)
		if yes {
			serialize.NewResponse(400, serialize.WithMsg("参数错误"), serialize.WithErr(errors.New("description 不合法"))).ToJSON(c)
			return
		}
		shortLinkGroup.Description = form.Description
	}

	ctx := middleware.WrapCtx(c)
	err := h.iDao.UpdateByGidAndCUserId(ctx, shortLinkGroup)
	if err != nil {
		err = errors.Wrap(err, "写入数据库时失败")
		logger.Error("UpdateByGID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		serialize.NewResponse(500, serialize.WithMsg("更新失败"), serialize.WithErr(err)).ToJSON(c)
		return
	}
	logger.Info("更新成功", logger.Any("shortLinkGroup", shortLinkGroup), middleware.GCtxRequestIDField(c))
	serialize.NewResponse(200, serialize.WithMsg("更新成功")).ToJSON(c)

}

// DeleteByID delete a record by id
// @Summary delete shortLinkGroups
// @Description delete shortLinkGroups by id
// @Tags shortLinkGroups
// @accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} types.DeleteShortLinkGroupByIDRespond{}
// @Router /api/v1/shortLinkGroups/{id} [delete]
func (h *shortLinkGroupsHandler) DeleteByID(c *gin.Context) {
	//_, id, isAbort := getShortLinkGroupIDFromPath(c)
	//if isAbort {
	//	response.Error(c, ecode.InvalidParams)
	//	return
	//}
	//
	//ctx := middleware.WrapCtx(c)
	//err := h.iDao.DeleteByID(ctx, id)
	//if err != nil {
	//	logger.Error("DeleteByID error", logger.Err(err), logger.Any("id", id), middleware.GCtxRequestIDField(c))
	//	response.Output(c, ecode.InternalServerError.ToHTTPCode())
	//	return
	//}
	//
	//response.Success(c)
	panic("implement me")
}

// DeleteByIDs delete records by batch id
// @Summary delete shortLinkGroups
// @Description delete shortLinkGroups by batch id
// @Tags shortLinkGroups
// @Param data body types.DeleteShortLinkGroupByIDsRequest true "id array"
// @Accept json
// @Produce json
// @Success 200 {object} types.DeleteShortLinkGroupByIDsRespond{}
// @Router /api/v1/shortLinkGroups/delete/ids [post]
func (h *shortLinkGroupsHandler) DeleteByIDs(c *gin.Context) {
	form := &types.DeleteShortLinkGroupByIDsRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	err = h.iDao.DeleteByIDs(ctx, form.IDs)
	if err != nil {
		logger.Error("GetByIDs error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
}

// GetByID get a record by id
// @Summary get shortLinkGroups detail
// @Description get shortLinkGroups detail by id
// @Tags shortLinkGroups
// @Param id path string true "id"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetShortLinkGroupByIDRespond{}
// @Router /api/v1/shortLinkGroups/{id} [get]
func (h *shortLinkGroupsHandler) GetByID(c *gin.Context) {
	panic("implement me")
	//idStr, id, isAbort := getShortLinkGroupIDFromPath(c)
	//if isAbort {
	//	response.Error(c, ecode.InvalidParams)
	//	return
	//}
	//
	//ctx := middleware.WrapCtx(c)
	//shortLinkGroups, err := h.iDao.GetByID(ctx, id)
	//if err != nil {
	//	if errors.Is(err, query.ErrNotFound) {
	//		logger.Warn("GetByID not found", logger.Err(err), logger.Any("id", id), middleware.GCtxRequestIDField(c))
	//		response.Error(c, ecode.NotFound)
	//	} else {
	//		logger.Error("GetByID error", logger.Err(err), logger.Any("id", id), middleware.GCtxRequestIDField(c))
	//		response.Output(c, ecode.InternalServerError.ToHTTPCode())
	//	}
	//	return
	//}
	//
	//data := &types.ShortLinkGroupObjDetail{}
	//err = copier.Copy(data, shortLinkGroups)
	//if err != nil {
	//	response.Error(c, ecode.ErrGetByIDShortLinkGroup)
	//	return
	//}
	//data.ID = idStr
	//
	//response.Success(c, gin.H{"shortLinkGroups": data})
}

// GetByCondition get a record by condition
// @Summary get shortLinkGroups by condition
// @Description get shortLinkGroups by condition
// @Tags shortLinkGroups
// @Param data body types.Conditions true "query condition"
// @Accept json
// @Produce json
// @Success 200 {object} types.GetShortLinkGroupByConditionRespond{}
// @Router /api/v1/shortLinkGroups/condition [post]
func (h *shortLinkGroupsHandler) GetByCondition(c *gin.Context) {
	panic("implement me")

	//form := &types.GetShortLinkGroupByConditionRequest{}
	//err := c.ShouldBindJSON(form)
	//if err != nil {
	//	logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
	//	response.Error(c, ecode.InvalidParams)
	//	return
	//}
	//err = form.Conditions.CheckValid()
	//if err != nil {
	//	logger.Warn("Parameters error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
	//	response.Error(c, ecode.InvalidParams)
	//	return
	//}
	//
	//ctx := middleware.WrapCtx(c)
	//shortLinkGroups, err := h.iDao.GetByCondition(ctx, &form.Conditions)
	//if err != nil {
	//	if errors.Is(err, query.ErrNotFound) {
	//		logger.Warn("GetByCondition not found", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
	//		response.Error(c, ecode.NotFound)
	//	} else {
	//		logger.Error("GetByCondition error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
	//		response.Output(c, ecode.InternalServerError.ToHTTPCode())
	//	}
	//	return
	//}
	//
	//data := &types.ShortLinkGroupObjDetail{}
	//err = copier.Copy(data, shortLinkGroups)
	//if err != nil {
	//	response.Error(c, ecode.ErrGetByIDShortLinkGroup)
	//	return
	//}
	//data.ID = utils.Uint64ToStr(shortLinkGroups.ID)
	//
	//response.Success(c, gin.H{"shortLinkGroups": data})
}

func getShortLinkGroupIDFromPath(c *gin.Context) (string, int, error) {
	gidStr := c.Param("gid")
	id, err := strconv.Atoi(gidStr)
	if err != nil || id == 0 {
		err = errors.Wrap(err, "参数错误")
		return "", 0, err
	}

	return gidStr, id, nil
}

func convertShortLinkGroup(shortLinkGroups *model.ShortLinkGroup) (*types.ShortLinkGroupObjDetail, error) {
	data := &types.ShortLinkGroupObjDetail{}
	err := copier.Copy(data, shortLinkGroups)
	if err != nil {
		return nil, err
	}
	data.ID = utils.Uint64ToStr(uint64(shortLinkGroups.ID))
	return data, nil
}

func convertShortLinkGroups(fromValues []*model.ShortLinkGroup) ([]*types.ShortLinkGroupObjDetail, error) {
	var toValues []*types.ShortLinkGroupObjDetail
	for _, v := range fromValues {
		data, err := convertShortLinkGroup(v)
		if err != nil {
			return nil, err
		}
		toValues = append(toValues, data)
	}

	return toValues, nil
}

func CheckSort(sort string) bool {
	// 定义有效的排序关键字
	validSorts := []string{"ASC", "DESC"}

	// 将输入字符串转换为大写，以便进行不区分大小写的比较
	sort = strings.ToUpper(sort)

	// 遍历有效关键字数组，检查输入字符串是否匹配
	for _, validSort := range validSorts {
		if sort == validSort {
			return true
		}
	}

	// 如果没有匹配到任何有效关键字，返回 false
	return false
}
func parseParams(c *gin.Context, form any) bool {
	err := c.ShouldBindJSON(form)
	if err != nil {
		serialize.NewResponse(400, serialize.WithMsg("参数错误"), serialize.WithErr(err)).ToJSON(c)
		return false
	}
	return true
}

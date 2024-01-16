package handler

import (
	"SnapLink/internal/cache"
	"SnapLink/internal/dao"
	"SnapLink/internal/ecode"
	"SnapLink/internal/model"
	"SnapLink/internal/types"
	"SnapLink/pkg/serialize"
	"github.com/pkg/errors"
	"github.com/zhufuyi/sponge/pkg/mysql/query"
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
	UpdateByID(c *gin.Context)
	GetByID(c *gin.Context)
	GetByCondition(c *gin.Context)
	ListByIDs(c *gin.Context)
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
// @Summary 列出所有短链接分组，支持分页和条件查询，不传参数则返回所有
// @Description 列出所有短链接分组，支持分页和条件查询，不传参数则返回所有
// @Tags shortLinkGroup
// @Accept application/json
// @Produce application/json
// @Param page body int false "页码"
// @Param size body int false "每页数量"
// @Param sort body string false "排序字段"
// @Success 200 {object} types.ListShortLinkGroupRespond{}
// @Failure 400 string "{"msg": "参数错误"}"
// @Router /api/v1/ShortLinkGroup/list [post]
func (h *shortLinkGroupsHandler) List(c *gin.Context) {
	//1. 参数解析
	form := &types.ListShortLinkGroupRequest{}
	err := c.ShouldBindJSON(form)
	if err != nil {
		serialize.NewResponse(400, serialize.WithMsg("参数错误"), serialize.WithErr(err)).ToJSON(c)
		return
	}
	param := query.Params{
		Page: form.Page,
		Size: form.Size,
		Sort: form.Sort,
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
	_, id, isAbort := getShortLinkGroupIDFromPath(c)
	if isAbort {
		response.Error(c, ecode.InvalidParams)
		return
	}

	ctx := middleware.WrapCtx(c)
	err := h.iDao.DeleteByID(ctx, id)
	if err != nil {
		logger.Error("DeleteByID error", logger.Err(err), logger.Any("id", id), middleware.GCtxRequestIDField(c))
		response.Output(c, ecode.InternalServerError.ToHTTPCode())
		return
	}

	response.Success(c)
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

// UpdateByID update information by id
// @Summary update shortLinkGroups
// @Description update shortLinkGroups information by id
// @Tags shortLinkGroups
// @accept json
// @Produce json
// @Param id path string true "id"
// @Param data body types.UpdateShortLinkGroupByIDRequest true "shortLinkGroups information"
// @Success 200 {object} types.UpdateShortLinkGroupByIDRespond{}
// @Router /api/v1/shortLinkGroups/{id} [put]
func (h *shortLinkGroupsHandler) UpdateByID(c *gin.Context) {
	//_, id, isAbort := getShortLinkGroupIDFromPath(c)
	//if isAbort {
	//	response.Error(c, ecode.InvalidParams)
	//	return
	//}
	//
	//form := &types.UpdateShortLinkGroupByIDRequest{}
	//err := c.ShouldBindJSON(form)
	//if err != nil {
	//	logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
	//	response.Error(c, ecode.InvalidParams)
	//	return
	//}
	//form.ID = id
	//
	//shortLinkGroups := &model.ShortLinkGroup{}
	//err = copier.Copy(shortLinkGroups, form)
	//if err != nil {
	//	response.Error(c, ecode.ErrUpdateByIDShortLinkGroup)
	//	return
	//}
	//
	//ctx := middleware.WrapCtx(c)
	//err = h.iDao.UpdateByID(ctx, shortLinkGroups)
	//if err != nil {
	//	logger.Error("UpdateByID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
	//	response.Output(c, ecode.InternalServerError.ToHTTPCode())
	//	return
	//}
	//
	//response.Success(c)
	panic("implement me")

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

// ListByIDs list of records by batch id
// @Summary list of shortLinkGroups by batch id
// @Description list of shortLinkGroups by batch id
// @Tags shortLinkGroups
// @Param data body types.ListShortLinkGroupByIDsRequest true "id array"
// @Accept json
// @Produce json
// @Success 200 {object} types.ListShortLinkGroupByIDsRespond{}
// @Router /api/v1/shortLinkGroups/list/ids [post]
func (h *shortLinkGroupsHandler) ListByIDs(c *gin.Context) {
	panic("implement me")

	//form := &types.ListShortLinkGroupByIDsRequest{}
	//err := c.ShouldBindJSON(form)
	//if err != nil {
	//	logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
	//	response.Error(c, ecode.InvalidParams.WithOutMsg("参数错误"), "详细错误信息")
	//	response.Output(c, ecode.Unauthorized.WithOutMsg("错误简单描述").ToHTTPCode(), "详细错误信息")
	//	return
	//}
	//
	//ctx := middleware.WrapCtx(c)
	//shortLinkGroupsMap, err := h.iDao.GetByIDs(ctx, form.IDs)
	//if err != nil {
	//	logger.Error("GetByIDs error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
	//	response.Output(c, ecode.InternalServerError.ToHTTPCode())
	//	return
	//}
	//
	//shortLinkGroups := []*types.ShortLinkGroupObjDetail{}
	//for _, id := range form.IDs {
	//	if v, ok := shortLinkGroupsMap[id]; ok {
	//		record, err := convertShortLinkGroup(v)
	//		if err != nil {
	//			response.Error(c, ecode.ErrListShortLinkGroup)
	//			return
	//		}
	//		shortLinkGroups = append(shortLinkGroups, record)
	//	}
	//}
	//
	//response.Success(c, gin.H{
	//	"shortLinkGroups": shortLinkGroups,
	//})
}

// ListByLastID get records by last id and limit
// @Summary list of shortLinkGroups by last id and limit
// @Description list of shortLinkGroups by last id and limit
// @Tags shortLinkGroups
// @accept json
// @Produce json
// @Param lastID query int true "last id, default is MaxInt64"
// @Param limit query int false "size in each page" default(10)
// @Param sort query string false "sort by column name of table, and the "-" sign before column name indicates reverse order" default(-id)
// @Success 200 {object} types.ListShortLinkGroupRespond{}
// @Router /api/v1/shortLinkGroups/list [get]
func (h *shortLinkGroupsHandler) ListByLastID(c *gin.Context) {
	panic("implement me")

	//lastID := utils.StrToUint64(c.Query("lastID"))
	//if lastID == 0 {
	//	lastID = math.MaxInt64
	//}
	//limit := utils.StrToInt(c.Query("limit"))
	//if limit == 0 {
	//	limit = 10
	//}
	//sort := c.Query("sort")
	//
	//ctx := middleware.WrapCtx(c)
	//shortLinkGroups, err := h.iDao.GetByLastID(ctx, lastID, limit, sort)
	//if err != nil {
	//	logger.Error("GetByLastID error", logger.Err(err), logger.Uint64("latsID", lastID), logger.Int("limit", limit), middleware.GCtxRequestIDField(c))
	//	response.Output(c, ecode.InternalServerError.ToHTTPCode())
	//	return
	//}
	//
	//data, err := convertShortLinkGroup(shortLinkGroups)
	//if err != nil {
	//	response.Error(c, ecode.ErrListByLastIDShortLinkGroup)
	//	return
	//}
	//
	//response.Success(c, gin.H{
	//	"shortLinkGroups": data,
	//})
}

func getShortLinkGroupIDFromPath(c *gin.Context) (string, uint64, bool) {
	idStr := c.Param("id")
	id, err := utils.StrToUint64E(idStr)
	if err != nil || id == 0 {
		logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
		return "", 0, true
	}

	return idStr, id, false
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

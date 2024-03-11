package handler

import (
	"SnapLink/internal/cache"
	"SnapLink/internal/dao"
	"SnapLink/internal/ecode"
	"SnapLink/internal/model"
	"SnapLink/internal/types"
	"SnapLink/internal/utils/GenerateShortLink"
	"SnapLink/pkg/serialize"
	"context"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zhufuyi/sponge/pkg/gin/middleware"
	"github.com/zhufuyi/sponge/pkg/logger"
)

var _ ShortLinkHandler = (*shortLinkHandler)(nil)

// ShortLinkHandler defining the handler interface
type ShortLinkHandler interface {
	Create(c *gin.Context)
	//CreateBatch(c *gin.Context)
	//DeleteByID(c *gin.Context)
	//DeleteByIDs(c *gin.Context)
	//UpdateByID(c *gin.Context)
	//GetByID(c *gin.Context)
	//GetByCondition(c *gin.Context)
	//ListByIDs(c *gin.Context)
	//ListByLastID(c *gin.Context)
	//List(c *gin.Context)
}

type shortLinkHandler struct {
	iDao dao.ShortLinkDao
}

// NewShortLinkHandler creating the handler interface
func NewShortLinkHandler() ShortLinkHandler {
	h := &shortLinkHandler{
		iDao: dao.NewShortLinkDao(
			cache.NewShortLinkCache(model.GetCacheType()),
		),
	}
	h.makeShortLinkBF()
	return h
}

func (h *shortLinkHandler) makeShortLinkBF() {
	//todo 此处改为远程配置
	err := cache.BFCreate(context.Background(), "shortLink", 0.001, 1e9)
	// 如果创建失败，说明已经存在，正常结束即可
	// 如果创建成功，说明不存在，需要添加数据
	if err == nil {
		//todo implement me
		panic("implement me")

	}
}

// Create 创建短链接
// @Summary 创建短链接
// @Description 创建短链接
// @Tags shortLink
// @Accept application/json
// @Produce application/json
// @Param Authorization header string true "Bearer token"
// @Param originUrl body string true "原始链接"
// @Param gid body string false "组id"
// @Param createdType body int false "创建类型 0:接口创建 1:控制台创建"
// @Param validDate body string true "有效时间"
// @Param validDateType body int false "有效类型"
// @Param describe body string false "描述"
// @Success 200 {object} types.CreateShortLinkRespond{}
// @RedirectInfo /api/v1/shortLink [post]
func (h *shortLinkHandler) Create(c *gin.Context) {
	form := new(types.CreateShortLinkRequest)

	if err := c.ShouldBind(form); err != nil {
		serialize.NewResponseWithErrCode(ecode.ClientError, serialize.WithErr(err)).ToJSON(c)
		return
	}
	//1. 解析url
	u, err := url.Parse(form.OriginUrl)
	if err != nil {
		err = errors.Wrap(err, "url格式错误")
		serialize.NewResponse(400, serialize.WithMsg("参数错误"), serialize.WithErr(err)).ToJSON(c)
		return
	}
	//2. 生成短链接
	sLink := model.ShortLink{
		Clicks:        0,
		Enable:        1,
		Domain:        u.Host,
		OriginUrl:     u.String(),
		Gid:           form.Gid,
		Description:   form.Description,
		CreatedType:   form.CreatedType,
		ValidDateType: form.ValidDateType,
	}
	if sLink.ValidDateType > 0 {
		sLink.ValidTime, err = time.Parse("2006-01-02 15:04:05", form.ValidDate)
	}
	if err != nil {
		serialize.NewResponse(400, serialize.WithMsg("参数错误"), serialize.WithErr(err)).ToJSON(c)
		return
	}

	//2. 生成hash
	sLink.Uri = ToHash(u)
	//3. 保存到数据库
	// 对布隆过滤器误判的情况进行判断
	ctx := middleware.WrapCtx(c)
	err = h.iDao.Create(ctx, &sLink)

	// 特别对于唯一索引的错误进行处理
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			logger.Warn("短链接已经存在", logger.Any("sLink", sLink), middleware.GCtxRequestIDField(c))
			serialize.NewResponse(500, serialize.WithMsg("短链接已经存在")).ToJSON(c)
			return
		}
		logger.Error("Create error", logger.Err(err), logger.Any("sLink", sLink), middleware.GCtxRequestIDField(c))
		serialize.NewResponse(500, serialize.WithMsg("创建失败"), serialize.WithErr(err)).ToJSON(c)
		return
	}

	fullShortURL := makeFullShortURL(sLink.Domain, sLink.Uri)
	logger.Info("创建短链接成功", logger.Any("sLink", sLink), logger.String("fullShortURL", fullShortURL), middleware.GCtxRequestIDField(c))
	serialize.NewResponse(200, serialize.WithData(fullShortURL)).ToJSON(c)
}

//	func (h *shortLinkHandler) CreateBatch(c *gin.Context) {
//		//todo implement me
//		panic("implement me")
//	}
//
// // DeleteByID delete a record by id
// // @Summary delete shortLink
// // @Description delete shortLink by id
// // @Tags shortLink
// // @accept json
// // @Produce json
// // @Param id path string true "id"
// // @Success 200 {object} types.DeleteShortLinkByIDRespond{}
// // @RedirectInfo /api/v1/shortLink/{id} [delete]
//
//	func (h *shortLinkHandler) DeleteByID(c *gin.Context) {
//		_, id, isAbort := getShortLinkIDFromPath(c)
//		if isAbort {
//			response.Error(c, ecode.InvalidParams)
//			return
//		}
//
//		ctx := middleware.WrapCtx(c)
//		err := h.iDao.DeleteByID(ctx, id)
//		if err != nil {
//			logger.Error("DeleteByGidAndCUserId error", logger.Err(err), logger.Any("id", id), middleware.GCtxRequestIDField(c))
//			response.Output(c, ecode.InternalServerError.ToHTTPCode())
//			return
//		}
//
//		response.Success(c)
//	}
//
// // DeleteByIDs delete records by batch id
// // @Summary delete shortLinks
// // @Description delete shortLinks by batch id
// // @Tags shortLink
// // @Param data body types.DeleteShortLinksByIDsRequest true "id array"
// // @Accept json
// // @Produce json
// // @Success 200 {object} types.DeleteShortLinksByIDsRespond{}
// // @RedirectInfo /api/v1/shortLink/delete/ids [post]
//
//	func (h *shortLinkHandler) DeleteByIDs(c *gin.Context) {
//		form := &types.DeleteShortLinksByIDsRequest{}
//		err := c.ShouldBindJSON(form)
//		if err != nil {
//			logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
//			response.Error(c, ecode.InvalidParams)
//			return
//		}
//
//		ctx := middleware.WrapCtx(c)
//		err = h.iDao.DeleteByIDs(ctx, form.IDs)
//		if err != nil {
//			logger.Error("GetByIDs error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
//			response.Output(c, ecode.InternalServerError.ToHTTPCode())
//			return
//		}
//
//		response.Success(c)
//	}
//
// // UpdateByID update information by id
// // @Summary update shortLink
// // @Description update shortLink information by id
// // @Tags shortLink
// // @accept json
// // @Produce json
// // @Param id path string true "id"
// // @Param data body types.UpdateShortLinkByIDRequest true "shortLink information"
// // @Success 200 {object} types.UpdateShortLinkByIDRespond{}
// // @RedirectInfo /api/v1/shortLink/{id} [put]
//
//	func (h *shortLinkHandler) UpdateByID(c *gin.Context) {
//		_, id, isAbort := getShortLinkIDFromPath(c)
//		if isAbort {
//			response.Error(c, ecode.InvalidParams)
//			return
//		}
//
//		form := &types.UpdateShortLinkByIDRequest{}
//		err := c.ShouldBindJSON(form)
//		if err != nil {
//			logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
//			response.Error(c, ecode.InvalidParams)
//			return
//		}
//		form.ID = id
//
//		shortLink := &model.ShortLink{}
//		err = copier.Copy(shortLink, form)
//		if err != nil {
//			response.Error(c, ecode.ErrUpdateByIDShortLink)
//			return
//		}
//
//		ctx := middleware.WrapCtx(c)
//		err = h.iDao.UpdateByID(ctx, shortLink)
//		if err != nil {
//			logger.Error("UpdateByGID error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
//			response.Output(c, ecode.InternalServerError.ToHTTPCode())
//			return
//		}
//
//		response.Success(c)
//	}
//
// // GetByID get a record by id
// // @Summary get shortLink detail
// // @Description get shortLink detail by id
// // @Tags shortLink
// // @Param id path string true "id"
// // @Accept json
// // @Produce json
// // @Success 200 {object} types.GetShortLinkByIDRespond{}
// // @RedirectInfo /api/v1/shortLink/{id} [get]
//
//	func (h *shortLinkHandler) GetByID(c *gin.Context) {
//		idStr, id, isAbort := getShortLinkIDFromPath(c)
//		if isAbort {
//			response.Error(c, ecode.InvalidParams)
//			return
//		}
//
//		ctx := middleware.WrapCtx(c)
//		shortLink, err := h.iDao.GetByID(ctx, id)
//		if err != nil {
//			if errors.Is(err, query.ErrNotFound) {
//				logger.Warn("GetByURI not found", logger.Err(err), logger.Any("id", id), middleware.GCtxRequestIDField(c))
//				response.Error(c, ecode.NotFound)
//			} else {
//				logger.Error("GetByURI error", logger.Err(err), logger.Any("id", id), middleware.GCtxRequestIDField(c))
//				response.Output(c, ecode.InternalServerError.ToHTTPCode())
//			}
//			return
//		}
//
//		data := &types.ShortLinkObjDetail{}
//		err = copier.Copy(data, shortLink)
//		if err != nil {
//			response.Error(c, ecode.ErrGetByIDShortLink)
//			return
//		}
//		data.ID = idStr
//
//		response.Success(c, gin.H{"shortLink": data})
//	}
//
// // GetByCondition get a record by condition
// // @Summary get shortLink by condition
// // @Description get shortLink by condition
// // @Tags shortLink
// // @Param data body types.Conditions true "query condition"
// // @Accept json
// // @Produce json
// // @Success 200 {object} types.GetShortLinkByConditionRespond{}
// // @RedirectInfo /api/v1/shortLink/condition [post]
//
//	func (h *shortLinkHandler) GetByCondition(c *gin.Context) {
//		form := &types.GetShortLinkByConditionRequest{}
//		err := c.ShouldBindJSON(form)
//		if err != nil {
//			logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
//			response.Error(c, ecode.InvalidParams)
//			return
//		}
//		err = form.Conditions.CheckValid()
//		if err != nil {
//			logger.Warn("Parameters error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
//			response.Error(c, ecode.InvalidParams)
//			return
//		}
//
//		ctx := middleware.WrapCtx(c)
//		shortLink, err := h.iDao.GetByCondition(ctx, &form.Conditions)
//		if err != nil {
//			if errors.Is(err, query.ErrNotFound) {
//				logger.Warn("GetByCondition not found", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
//				response.Error(c, ecode.NotFound)
//			} else {
//				logger.Error("GetByCondition error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
//				response.Output(c, ecode.InternalServerError.ToHTTPCode())
//			}
//			return
//		}
//
//		data := &types.ShortLinkObjDetail{}
//		err = copier.Copy(data, shortLink)
//		if err != nil {
//			response.Error(c, ecode.ErrGetByIDShortLink)
//			return
//		}
//		data.ID = utils.Uint64ToStr(uint64(shortLink.ID))
//
//		response.Success(c, gin.H{"shortLink": data})
//	}
//
// // ListByIDs list of records by batch id
// // @Summary list of shortLinks by batch id
// // @Description list of shortLinks by batch id
// // @Tags shortLink
// // @Param data body types.ListShortLinksByIDsRequest true "id array"
// // @Accept json
// // @Produce json
// // @Success 200 {object} types.ListShortLinksByIDsRespond{}
// // @RedirectInfo /api/v1/shortLink/list/ids [post]
//
//	func (h *shortLinkHandler) ListByIDs(c *gin.Context) {
//		form := &types.ListShortLinksByIDsRequest{}
//		err := c.ShouldBindJSON(form)
//		if err != nil {
//			logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
//			response.Error(c, ecode.InvalidParams.WithOutMsg("参数错误"), "详细错误信息")
//			response.Output(c, ecode.Unauthorized.WithOutMsg("错误简单描述").ToHTTPCode(), "详细错误信息")
//			return
//		}
//
//		ctx := middleware.WrapCtx(c)
//		shortLinkMap, err := h.iDao.GetByIDs(ctx, form.IDs)
//		if err != nil {
//			logger.Error("GetByIDs error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
//			response.Output(c, ecode.InternalServerError.ToHTTPCode())
//			return
//		}
//
//		var shortLinks []*types.ShortLinkObjDetail
//		for _, id := range form.IDs {
//			if v, ok := shortLinkMap[id]; ok {
//				record, err := convertShortLink(v)
//				if err != nil {
//					response.Error(c, ecode.ErrListShortLink)
//					return
//				}
//				shortLinks = append(shortLinks, record)
//			}
//		}
//
//		response.Success(c, gin.H{
//			"shortLinks": shortLinks,
//		})
//	}
//
// // ListByLastID get records by last id and limit
// // @Summary list of shortLinks by last id and limit
// // @Description list of shortLinks by last id and limit
// // @Tags shortLink
// // @accept json
// // @Produce json
// // @Param lastID query int true "last id, default is MaxInt64"
// // @Param limit query int false "size in each page" default(10)
// // @Param sort query string false "sort by column name of table, and the "-" sign before column name indicates reverse order" default(-id)
// // @Success 200 {object} types.ListShortLinksRespond{}
// // @RedirectInfo /api/v1/shortLink/list [get]
//
//	func (h *shortLinkHandler) ListByLastID(c *gin.Context) {
//		lastID := utils.StrToUint64(c.Query("lastID"))
//		if lastID == 0 {
//			lastID = math.MaxInt64
//		}
//		limit := utils.StrToInt(c.Query("limit"))
//		if limit == 0 {
//			limit = 10
//		}
//		sort := c.Query("sort")
//
//		ctx := middleware.WrapCtx(c)
//		shortLinks, err := h.iDao.GetByLastID(ctx, lastID, limit, sort)
//		if err != nil {
//			logger.Error("GetByLastID error", logger.Err(err), logger.Uint64("latsID", lastID), logger.Int("limit", limit), middleware.GCtxRequestIDField(c))
//			response.Output(c, ecode.InternalServerError.ToHTTPCode())
//			return
//		}
//
//		data, err := convertShortLinks(shortLinks)
//		if err != nil {
//			response.Error(c, ecode.ErrListByLastIDShortLink)
//			return
//		}
//
//		response.Success(c, gin.H{
//			"shortLinks": data,
//		})
//	}
//
// // List of records by query parameters
// // @Summary list of shortLinks by query parameters
// // @Description list of shortLinks by paging and conditions
// // @Tags shortLink
// // @accept json
// // @Produce json
// // @Param data body types.Params true "query parameters"
// // @Success 200 {object} types.ListShortLinksRespond{}
// // @RedirectInfo /api/v1/shortLink/list [post]
//
//	func (h *shortLinkHandler) List(c *gin.Context) {
//		form := &types.ListShortLinksRequest{}
//		err := c.ShouldBindJSON(form)
//		if err != nil {
//			logger.Warn("ShouldBindJSON error: ", logger.Err(err), middleware.GCtxRequestIDField(c))
//			response.Error(c, ecode.InvalidParams)
//			return
//		}
//
//		ctx := middleware.WrapCtx(c)
//		shortLinks, total, err := h.iDao.GetByColumns(ctx, &form.Params)
//		if err != nil {
//			logger.Error("GetByColumns error", logger.Err(err), logger.Any("form", form), middleware.GCtxRequestIDField(c))
//			response.Output(c, ecode.InternalServerError.ToHTTPCode())
//			return
//		}
//
//		data, err := convertShortLinks(shortLinks)
//		if err != nil {
//			response.Error(c, ecode.ErrListShortLink)
//			return
//		}
//
//		response.Success(c, gin.H{
//			"shortLinks": data,
//			"total":      total,
//		})
//	}
//
//	func getShortLinkIDFromPath(c *gin.Context) (string, uint64, bool) {
//		idStr := c.Param("id")
//		id, err := utils.StrToUint64E(idStr)
//		if err != nil || id == 0 {
//			logger.Warn("StrToUint64E error: ", logger.String("idStr", idStr), middleware.GCtxRequestIDField(c))
//			return "", 0, true
//		}
//
//		return idStr, id, false
//	}
//
//	func convertShortLink(shortLink *model.ShortLink) (*types.ShortLinkObjDetail, error) {
//		data := &types.ShortLinkObjDetail{}
//		err := copier.Copy(data, shortLink)
//		if err != nil {
//			return nil, err
//		}
//		data.ID = utils.Uint64ToStr(uint64(shortLink.ID))
//		return data, nil
//	}
//
//	func convertShortLinks(fromValues []*model.ShortLink) ([]*types.ShortLinkObjDetail, error) {
//		var toValues []*types.ShortLinkObjDetail
//		for _, v := range fromValues {
//			data, err := convertShortLink(v)
//			if err != nil {
//				return nil, err
//			}
//			toValues = append(toValues, data)
//		}
//
//		return toValues, nil
//	}
//
// makeFullShortURL 生成完整的短链接
func makeFullShortURL(domain, uri string) string {
	//此处配置从配置文件中获取
	//todo 从配置文件中获取
	u := url.URL{
		Scheme: "http",
		Host:   domain,
		Path:   uri,
	}
	return u.String()
}

// ToHash  短链接转hash
func ToHash(u *url.URL) string {
	// 生成 hash
	// 1. 生成 hash
	// 尝试生成 10 次，直到生成不重复的hash
	uri := GenerateShortLink.GenerateHash(u.Path)
	for i := 1; i <= 10; i++ {
		// 同一域名下的短链接不能重复
		data := makeFullShortURL(u.Host, uri)
		//为了在布隆过滤器挂掉后仍然可以使用,忽略布隆过滤器的错误
		exist, _ := cache.BFExists(context.Background(), "shortLink", data)
		//如果此数据已经存在，再次生成
		if exist {
			uri = GenerateShortLink.GenerateHash(u.Path)
			continue
		}
		// 误判的情况有
		// 1. 误判为存在，但是实际不存在。这种情况可以无视
		// 2. 误判为不存在，但是实际存在，这种情况可以基于数据库的唯一索引来解决。这种情况主要是由于部分短链接未被加载入布隆过滤器中。
		_ = cache.BFAdd(context.Background(), "shortLink", data)
		break
	}
	return uri
}

package handler

import (
	"SnapLink/internal/bloomFilter"
	"SnapLink/internal/config"
	"SnapLink/internal/dao"
	"SnapLink/internal/ecode"
	"SnapLink/internal/elasticsearch"
	"SnapLink/internal/model"
	"SnapLink/internal/types"
	"SnapLink/internal/utils/GenerateShortLink"
	"SnapLink/pkg/serialize"
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zhufuyi/sponge/pkg/gin/middleware"
	"github.com/zhufuyi/sponge/pkg/logger"
)

var _ ShortLinkHandler = (*shortLinkHandler)(nil)

var Domain = "localhost"

// ShortLinkHandler defining the handler interface
type ShortLinkHandler interface {
	Create(c *gin.Context)
	CreateBatch(c *gin.Context)
	Update(c *gin.Context)
	List(c *gin.Context)
	Delete(c *gin.Context)
}

type shortLinkHandler struct {
	iDao dao.ShortLinkDao
}

// NewShortLinkHandler creating the handler interface
func NewShortLinkHandler() ShortLinkHandler {
	h := &shortLinkHandler{
		iDao: dao.NewShortLinkDao(),
		//iDaoStat: dao.NewLinkAccessStatisticDao(
		//	cache.NewLinkAccessStatisticCache(model.GetCacheType()),
		//),
	}
	Domain = config.Get().App.Domain
	return h
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
// @Redirect /api/v1/shortLink [post]
// 创建逻辑：https://drive.google.com/file/d/1GvDCdeJaA90WbBmUbVBH-1jsgT0XCiUZ/view?usp=sharing
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
		serialize.NewResponse(500, serialize.WithMsg("创建失败"), serialize.WithErr(err)).ToJSON(c)
		return
	}

	fullShortURL := makeFullShortURL(Domain, sLink.Uri)
	serialize.NewResponse(200, serialize.WithData(fullShortURL)).ToJSON(c)
}

// CreateBatch
// @Summary 批量创建短链接
// @Description 批量创建短链接
// @Tags shortLink
// @Accept application/json
// @Produce application/json
// @Param Authorization header string true "Bearer token"
// @Param
func (h *shortLinkHandler) CreateBatch(c *gin.Context) {
	forms := make([]*types.CreateShortLinkRequest, 0)
	if err := c.ShouldBind(&forms); err != nil {
		serialize.NewResponseWithErrCode(ecode.ClientError, serialize.WithErr(err)).ToJSON(c)
		return
	}
	l := len(forms)
	shortLinks := make([]*model.ShortLink, 0, l)
	for i := 0; i < l; i++ {
		u, err := url.Parse(forms[i].OriginUrl)
		if err != nil {
			err = errors.Wrap(err, "url格式错误")
			serialize.NewResponse(400, serialize.WithMsg("参数错误"), serialize.WithErr(err)).ToJSON(c)
			return
		}
		//2. 生成短链接
		sLink := &model.ShortLink{
			Enable:        1,
			Domain:        u.Host,
			OriginUrl:     u.String(),
			Gid:           forms[i].Gid,
			Description:   forms[i].Description,
			CreatedType:   forms[i].CreatedType,
			ValidDateType: forms[i].ValidDateType,
		}
		if sLink.ValidDateType > 0 {
			sLink.ValidTime, err = time.Parse("2006-01-02 15:04:05", forms[i].ValidDate)
		}
		if err != nil {
			serialize.NewResponse(400, serialize.WithMsg("参数错误"), serialize.WithErr(err)).ToJSON(c)
			return
		}
		//3. 生成hash
		sLink.Uri = ToHash(u)

		shortLinks = append(shortLinks, sLink)
	}
	ctx := middleware.WrapCtx(c)

	// 特别对于唯一索引的错误进行处理
	if sLink, err := h.iDao.CreateBatch(ctx, shortLinks); err != nil || sLink != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			serialize.NewResponseWithErrCode(ecode.ServiceError, serialize.WithMsg("短链接已经存在")).ToJSON(c)
			return
		}
		serialize.NewResponseWithErrCode(ecode.ServiceError, serialize.WithErr(err)).ToJSON(c)
		return
	}

	fullShortURLs := make([]string, 0, l)
	for i := 0; i < l; i++ {
		fullShortURLs = append(fullShortURLs, makeFullShortURL(Domain, shortLinks[i].Uri))
	}
	serialize.NewResponse(200, serialize.WithData(fullShortURLs)).ToJSON(c)
}

// List 分页查询短链接
// @Summary 分页查询短链接
// @Description 分页查询短链接
// @Tags shortLink
// @Accept application/json
// @Produce application/json
// @Param Authorization header string true "token"
// @Param gid query string false "组id"
// @Param current query int false "当前页"
// @Param size query int false "每页大小"
// @Param orderTag query string false "排序"
func (h *shortLinkHandler) List(c *gin.Context) {

	gid := c.Query("gid")
	currentStr := c.DefaultQuery("current", "1")
	sizeStr := c.DefaultQuery("size", "10")
	orderTag := c.Query("orderTag")
	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		serialize.NewResponseWithErrCode(ecode.ClientError, serialize.WithErr(err)).ToJSON(c)
		return
	}
	current, err := strconv.Atoi(currentStr)
	if err != nil {
		serialize.NewResponseWithErrCode(ecode.ClientError, serialize.WithErr(err)).ToJSON(c)
		return
	}
	ctx := middleware.WrapCtx(c)
	var list []*model.ShortLink
	var uris []string
	total, err := h.iDao.Count(ctx, gid)
	if err != nil {
		serialize.NewResponseWithErrCode(ecode.ServiceError, serialize.WithErr(err)).ToJSON(c)
		return
	}

	//转换
	res := types.ListShortLinkResponse{
		Total:   total,
		Size:    size,
		Current: current,
	}
	if orderTag == "" {
		//查询
		list, err = h.iDao.List(ctx, gid, current, size)
		if err != nil {
			serialize.NewResponseWithErrCode(ecode.ServiceError, serialize.WithErr(err)).ToJSON(c)
			return
		}
		//查询基本统计数据
		l := len(list)
		uris = make([]string, 0, l)
		for i := 0; i < l; i++ {
			uris = append(uris, list[i].Uri)
		}
		// 根据实际查询到的条数,进行数据钻取
		res.Records = make([]*types.ShortLinkRecord, 0, l)
		statics, err := searchTodayAccessStatic(ctx, uris)
		if err != nil {
			// 如果不是今日无数据，则返回错误
			if !errors.Is(err, ErrBucketsIsEmpty) {
				serialize.NewResponseWithErrCode(ecode.ServiceError, serialize.WithErr(err)).ToJSON(c)
				return
			}
		}
		// 制造响应数据
		for i := 0; i < l; i++ {
			record := &types.ShortLinkRecord{
				CreatedAt:     list[i].CreatedAt.Format("2006-01-02 15:04:05"),
				OriginUrl:     list[i].OriginUrl,
				ShortUrl:      makeFullShortURL(Domain, list[i].Uri),
				ValidDateType: list[i].ValidDateType,
				ValidDate:     list[i].ValidTime.Format("2006-01-02 15:04:05"),
				Describe:      list[i].Description,
			}
			// 如果查询不到数据，则返回 0
			if err == nil {
				static, ok := statics[list[i].Uri].(map[string]any)
				if ok {
					record.TodayPV = int(static["today_pv"].(map[string]any)["value"].(float64))
					record.TodayUV = int(static["today_uv"].(map[string]any)["value"].(float64))
					record.TodayUIP = int(static["today_uip"].(map[string]any)["value"].(float64))
				}

			}
			res.Records = append(res.Records, record)
		}
	} else {

	}

	serialize.NewResponse(200, serialize.WithData(res)).ToJSON(c)
}

var (
	ErrBucketsIsEmpty = errors.New("buckets is empty")
)

// 去 ES 中查询访问情况
func searchTodayAccessStatic(ctx context.Context, uris []string) (map[string]any, error) {
	// 获取 ES 实例
	body := map[string]any{
		"size":    0,
		"_source": false,
		"query": map[string]any{
			"terms": map[string]any{
				"info.uri.keyword": uris,
			},
		},
		"aggs": map[string]any{
			"statics": map[string]any{
				"terms": map[string]any{
					"field": "info.uri.keyword",
				},
				"aggs": map[string]any{
					"today_pv": map[string]any{
						"value_count": map[string]any{
							"field": "info.uri.keyword",
						},
					},
					"today_uip": map[string]any{
						"cardinality": map[string]any{
							"field": "ip.keyword",
						},
					},
					"today_uv": map[string]any{
						"cardinality": map[string]any{
							"field": "uid.keyword",
						},
					},
				},
			},
		},
	}
	index := fmt.Sprintf("logstash-accesslog-%s.*", time.Now().Format("2006.01.02"))
	return elasticsearch.Search(ctx, index, body, accessStaticResponseParser)
}

// accessStaticResponseParser 访问日志查询响应
func accessStaticResponseParser(body io.ReadCloser) (map[string]any, error) {
	var r map[string]interface{}
	if err := json.NewDecoder(body).Decode(&r); err != nil {
		return nil, errors.Wrap(err, "解析响应失败")
	}

	aggregations, ok := r["aggregations"].(map[string]any)
	if !ok {
		return nil, ErrBucketsIsEmpty
	}

	buckets := aggregations["statics"].(map[string]any)["buckets"].([]any)
	if len(buckets) == 0 {
		return nil, ErrBucketsIsEmpty
	}

	statics := make(map[string]any, len(buckets))
	for _, bucket := range buckets {
		key := bucket.(map[string]any)["key"].(string)
		statics[key] = bucket
	}

	return statics, nil
}

// Delete 删除短链接
// @Summary 删除短链接
// @Description 删除短链接
// @Tags shortLink
// @Accept application/json
// @Produce application/json
// @Param Authorization header string true "token"
// @Param uri path string true "短链接"
func (h *shortLinkHandler) Delete(c *gin.Context) {
	uri := c.Param("uri")
	if uri == "" {
		serialize.NewResponseWithErrCode(ecode.ClientError, serialize.WithMsg("uri不能为空")).ToJSON(c)
		return
	}
	ctx := middleware.WrapCtx(c)
	err := h.iDao.Delete(ctx, uri)
	if err != nil {
		serialize.NewResponseWithErrCode(ecode.ServiceError, serialize.WithErr(err)).ToJSON(c)
		return
	}
	serialize.NewResponse(200).ToJSON(c)
}

// Update 更新短链接
// API 链接: https://app.apifox.com/link/project/4131066/apis/api-153792353
// TAPD 需求链接: 【短链接修改】https://www.tapd.cn/52430403/prong/stories/view/1152430403001000021
func (h *shortLinkHandler) Update(c *gin.Context) {
	// 参数解析
	form := new(types.UpdateShortLinkRequest)
	if err := c.ShouldBind(form); err != nil {
		serialize.NewResponseWithErrCode(ecode.ClientError, serialize.WithErr(err)).ToJSON(c)
		return
	}
	validTime, err := time.Parse("2006-01-02 15:04:05", form.ValidDate)
	if err != nil {
		serialize.NewResponseWithErrCode(ecode.ClientError, serialize.WithErr(err)).ToJSON(c)
		return
	}
	ctx := middleware.WrapCtx(c)
	// 0. 获取对应短链接的基本信息
	info, err := h.iDao.GeRedirectByURI(ctx, form.Uri)
	if err != nil {
		// 尝试从此处进行攻击，对此进行防御
		if errors.Is(err, model.ErrRecordNotFound) {
			serialize.NewResponseWithErrCode(ecode.ClientError, serialize.WithErr(err)).ToJSON(c)
			return
		}
		serialize.NewResponseWithErrCode(ecode.ServiceError, serialize.WithErr(err)).ToJSON(c)
		return
	}
	// 构建更新后的短链接
	sl := &model.ShortLink{
		OriginUrl:     form.OriginUrl,
		Gid:           info.Gid,
		Uri:           form.Uri,
		Description:   form.Description,
		ValidDateType: form.ValidDateType,
		ValidTime:     validTime,
	}
	// 1. 校验短链接 gid 是否变更
	// 短链接未发生改变
	if strings.Compare(info.Gid, form.Gid) == 0 {
		// 更新短链接
		err = h.iDao.Update(ctx, sl)
		if err != nil {
			serialize.NewResponseWithErrCode(ecode.ServiceError, serialize.WithErr(err)).ToJSON(c)
			return
		}
		serialize.NewResponse(200, serialize.WithData(sl)).ToJSON(c)
		return
	}
	// 短链接发生改变
	err = h.iDao.UpdateWithMove(ctx, sl, form.Gid)
	if err != nil {
		serialize.NewResponseWithErrCode(ecode.ServiceError, serialize.WithErr(err)).ToJSON(c)
		return
	}
	serialize.NewResponse(200, serialize.WithData(sl)).ToJSON(c)
	return
}

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
		//为了在布隆过滤器挂掉后仍然可以使用,忽略布隆过滤器的错误
		exist, _ := bloomFilter.BFExists(context.Background(), "uri", uri)
		//如果此数据已经存在，再次生成
		if exist {
			uri = GenerateShortLink.GenerateHash(uri + u.Path)
			continue
		}
		// 误判的情况有
		// 1. 误判为存在，但是实际不存在。这种情况可以无视
		// 2. 误判为不存在，但是实际存在，这种情况可以基于数据库的唯一索引来解决。这种情况主要是由于部分短链接未被加载入布隆过滤器中。
		_ = bloomFilter.BFAdd(context.Background(), "uri", uri)
		break
	}
	return uri
}

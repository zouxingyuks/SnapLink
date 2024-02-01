package handler

import (
	"SnapLink/assets"
	"SnapLink/internal/cache"
	"SnapLink/internal/dao"
	"SnapLink/internal/model"
	"SnapLink/pkg/serialize"
	"github.com/gin-gonic/gin"
)

var _ RedirectHandler = (*redirectHandler)(nil)

type RedirectHandler interface {
	Redirect(c *gin.Context)
}
type redirectHandler struct {
	iDao dao.RedirectsDao
}

func NewRedirectHandler() RedirectHandler {
	return &redirectHandler{
		iDao: dao.NewRedirectsDao(
			cache.NewRedirectsCache(model.GetCacheType()),
		),
	}
}

// Redirect 访问短链接重定向到原始链接
// @Summary 访问短链接重定向到原始链接
// @Description 访问短链接重定向到原始链接
// @Tags 短链接
// @Accept json
// @Produce json
// @Param short_uri path string true "短链接"
// @Success 302 {string} string "重定向到原始链接"
// @Failure 400 {string} string "请求失败"
// @Router /{short_uri} [get]
func (h *redirectHandler) Redirect(c *gin.Context) {
	shortUri := c.Param("short_uri")
	//获取 short_uri 对应的原始链接
	ctx := c.Request.Context()
	info, err := h.iDao.GetByURI(ctx, shortUri)
	if err != nil {
		serialize.NewResponse(
			400,
			serialize.WithMsg("请求失败"),
			serialize.WithErr(err),
		).ToJSON(c)
		return
	}
	if info.OriginalURL == "" {
		c.HTML(200, assets.Path("html/page_not_found.html"), gin.H{})
		return
	}
	// 进行重定向
	c.Redirect(302, info.OriginalURL)
}

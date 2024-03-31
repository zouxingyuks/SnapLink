package handler

import (
	"SnapLink/internal/cache"
	"SnapLink/internal/dao"
	"SnapLink/internal/model"
	"SnapLink/pkg/serialize"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type RedirectHandler struct {
	iDao dao.RedirectsDao
}

func NewRedirectHandler() *RedirectHandler {
	h := &RedirectHandler{
		iDao: dao.NewRedirectsDao(
			cache.NewRedirectsCache(model.GetCacheType()),
		),
	}
	return h
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
// @Router /{uri} [get]
// 流程图: https://drive.google.com/file/d/1hAHa5ZzhMjueqcIlkjkpvrejxsdo0Qk_/view?usp=sharing
func (h *RedirectHandler) Redirect(c *gin.Context) {
	shortUri := c.Param("uri")
	//获取 short_uri 对应的原始链接
	ctx := c.Request.Context()
	info, err := h.iDao.GetByURI(ctx, shortUri)
	if err != nil {
		if errors.Is(err, model.ErrRecordNotFound) {
			serialize.NewResponse(404, serialize.WithMsg("短链接不存在")).ToJSON(c)
			return
		}
		serialize.NewResponse(
			400,
			serialize.WithMsg("请求失败"),
			serialize.WithErr(err),
		).ToJSON(c)
		return
	}
	c.Set("info", info)
	// 进行重定向
	c.Redirect(302, info.OriginalURL)
}

package handler

import (
	"SnapLink/internal/dao"
	"SnapLink/internal/ecode"
	"SnapLink/pkg/serialize"
	"github.com/gin-gonic/gin"
)

type FixHandler struct {
	iDao *dao.FixDao
}

func NewFixHandler() *FixHandler {
	return &FixHandler{
		iDao: dao.NewFixDao(),
	}
}

// RebulidBF 重建布隆过滤器
// 流程图 : https://drive.google.com/file/d/1RvYY7vC3be0z9u2ELOQukrQymTOQX-FZ/view?usp=sharing
func (h *FixHandler) RebulidBF(c *gin.Context) {
	errs := h.iDao.RebulidBF()
	if len(errs) > 0 {
		serialize.NewResponseWithErrCode(ecode.ServiceError, serialize.WithMsg("重建布隆过滤器失败"), serialize.WithData(errs)).ToJSON(c)
		return
	}
	serialize.NewResponse(200, serialize.WithMsg("重建布隆过滤器成功")).ToJSON(c)
}

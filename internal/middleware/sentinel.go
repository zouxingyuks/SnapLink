package middleware

import (
	"SnapLink/internal/ecode"
	"SnapLink/pkg/serialize"
	"github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/gin-gonic/gin"
)

func Sentinel(resource string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 通过api.Entry创建一个资源的访问入口，如果资源的访问量超过了限流规则中的阈值，就会返回限流错误
		e, b := api.Entry(resource, api.WithTrafficType(base.Inbound))
		defer e.Exit()
		if b != nil {
			serialize.NewResponseWithErrCode(ecode.FlowLimitError).ToJSON(c)
			c.Abort()
			return
		}
		defer e.Exit()

		c.Next()
	}
}

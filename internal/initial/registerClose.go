package initial

import (
	"context"
	"time"

	"SnapLink/internal/config"
	"SnapLink/internal/model"

	"github.com/zhufuyi/sponge/pkg/app"
	"github.com/zhufuyi/sponge/pkg/tracer"
)

// RegisterClose 注册关闭函数
func RegisterClose(servers []app.IServer) []app.Close {
	var closes []app.Close

	// close server
	for _, s := range servers {
		closes = append(closes, s.Stop)
	}

	// close mysql
	closes = append(closes, func() error {
		return model.CloseMysql()
	})

	// close redis
	if config.Get().App.CacheType == "redis" {
		closes = append(closes, func() error {
			return model.CloseRedis()
		})
	}

	// close tracing
	if config.Get().App.EnableTrace {
		closes = append(closes, func() error {
			ctx, _ := context.WithTimeout(context.Background(), 2*time.Second) //nolint
			return tracer.Close(ctx)
		})
	}

	return closes
}

package logger

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/pkg/errors"
	"github.com/zhufuyi/sponge/pkg/logger"
	"go.uber.org/zap"
)

// WatermillAdapter 日志适配器
type WatermillAdapter struct {
	fields []zap.Field // 用于存储累积的字段
}

// 转换 Watermill LogFields 到 zap.Fields
func (l *WatermillAdapter) toZapFields(fields watermill.LogFields) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields)+len(l.fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	return append(zapFields, l.fields...)
}

func (l *WatermillAdapter) Error(msg string, err error, fields watermill.LogFields) {
	logger.Error(errors.Wrap(err, msg).Error(), l.toZapFields(fields)...)
}

func (l *WatermillAdapter) Info(msg string, fields watermill.LogFields) {
	logger.Info(msg, l.toZapFields(fields)...)
}

func (l *WatermillAdapter) Debug(msg string, fields watermill.LogFields) {
	logger.Debug(msg, l.toZapFields(fields)...)
}

func (l *WatermillAdapter) Trace(msg string, fields watermill.LogFields) {
	// 注意：zap 默认不提供 Trace 级别，这里我们使用 Debug 级别作为替代
	logger.Debug(msg, l.toZapFields(fields)...)
}

// With 方法用于累积字段
func (l *WatermillAdapter) With(fields watermill.LogFields) watermill.LoggerAdapter {
	newFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		newFields = append(newFields, zap.Any(k, v))
	}

	// 返回一个新的 WatermillAdapter 实例，包含了新累积的字段
	return &WatermillAdapter{
		fields: append(l.fields, newFields...),
	}
}

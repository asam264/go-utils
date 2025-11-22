package batchwriter

import "github.com/asam264/go-utils/common"

// log 内部日志方法
func (bw *BatchWriter) log(level, msg string, fields ...common.Field) {
	allFields := append([]common.Field{{"name", bw.config.Name}}, fields...)
	switch level {
	case "info":
		bw.config.Logger.Info(msg, allFields...)
	case "warn":
		bw.config.Logger.Warn(msg, allFields...)
	case "error":
		bw.config.Logger.Error(msg, allFields...)
	}
}


package log

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// AnonymousLogFormatter Default Gin Log Formatter without IPs
func AnonymousLogFormatter(param gin.LogFormatterParams) string {
	var statusColor, methodColor, resetColor string
	if param.IsOutputColor() {
		statusColor = param.StatusCodeColor()
		methodColor = param.MethodColor()
		resetColor = param.ResetColor()
	}

	if param.Latency > time.Minute {
		// Truncate in a golang < 1.8 safe way
		param.Latency = param.Latency - param.Latency%time.Second
	}
	return fmt.Sprintf("[GIN] %v |%s %3d %s| %13v |%s %-7s %s %#v\n%s",
		param.TimeStamp.Format("2006/01/02 - 15:04:05"),
		statusColor, param.StatusCode, resetColor,
		param.Latency,
		methodColor, param.Method, resetColor,
		param.Path,
		param.ErrorMessage,
	)
}

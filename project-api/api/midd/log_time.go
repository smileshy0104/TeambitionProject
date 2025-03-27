package midd

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"time"
)

// RequestLog 返回一个中间件函数，用于记录HTTP请求的耗时信息。
// 该中间件主要关注请求的URI和请求处理的耗时，并通过日志记录下来。
// 使用场景：在处理HTTP请求时，需要对请求的处理时间进行监控和分析。
func RequestLog() func(*gin.Context) {
	// 返回一个闭包函数，用于实际的请求处理和日志记录。
	return func(c *gin.Context) {
		// 记录请求开始的时间。
		start := time.Now()
		// 调用c.Next()以继续执行后续的请求处理函数。
		c.Next()
		// 计算请求处理的耗时，单位为毫秒。
		diff := time.Now().UnixMilli() - start.UnixMilli()
		// 使用zap日志库记录请求URI和耗时信息。
		// 选择zap是因为其高性能和结构化日志的特点，适合在生产环境中使用。
		zap.L().Info(fmt.Sprintf("%s 用时 %d ms", c.Request.RequestURI, diff))
	}
}

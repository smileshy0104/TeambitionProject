package logs

import (
	"github.com/gin-gonic/gin"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"
	"time"
)

// LG 是全局的zap.Logger实例
var LG *zap.Logger

// LogConfig 是日志配置的结构体
type LogConfig struct {
	DebugFileName string `json:"debugFileName"`
	InfoFileName  string `json:"infoFileName"`
	WarnFileName  string `json:"warnFileName"`
	MaxSize       int    `json:"maxsize"`
	MaxAge        int    `json:"max_age"`
	MaxBackups    int    `json:"max_backups"`
}

// InitLogger 初始化Logger
func InitLogger(cfg *LogConfig) (err error) {
	// 分别获取debug、info和warn日志的写入器
	writeSyncerDebug := getLogWriter(cfg.DebugFileName, cfg.MaxSize, cfg.MaxBackups, cfg.MaxAge)
	writeSyncerInfo := getLogWriter(cfg.InfoFileName, cfg.MaxSize, cfg.MaxBackups, cfg.MaxAge)
	writeSyncerWarn := getLogWriter(cfg.WarnFileName, cfg.MaxSize, cfg.MaxBackups, cfg.MaxAge)
	// 获取日志的编码器
	encoder := getEncoder()
	//文件输出
	// 分别创建debug、info和warn日志的核心配置
	debugCore := zapcore.NewCore(encoder, writeSyncerDebug, zapcore.DebugLevel)
	infoCore := zapcore.NewCore(encoder, writeSyncerInfo, zapcore.InfoLevel)
	warnCore := zapcore.NewCore(encoder, writeSyncerWarn, zapcore.WarnLevel)
	//标准输出
	// 创建控制台日志的编码器
	consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	// 创建控制台日志的核心配置
	std := zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stdout), zapcore.DebugLevel)
	// 合并所有日志核心配置
	core := zapcore.NewTee(debugCore, infoCore, warnCore, std)
	// 创建并初始化全局的logger实例
	LG = zap.New(core, zap.AddCaller())
	// 替换zap包中全局的logger实例，后续在其他包中只需使用zap.L()调用即可
	zap.ReplaceGlobals(LG)
	return
}

// getEncoder 创建并返回日志的编码器
func getEncoder() zapcore.Encoder {
	// 设置日志编码器的配置
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.TimeKey = "time"
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	// 创建并返回JSON格式的日志编码器
	return zapcore.NewJSONEncoder(encoderConfig)
}

// getLogWriter 创建并返回日志的写入器
// 参数:
//
//	filename - 日志文件名
//	maxSize - 日志文件最大大小（MB）
//	maxBackup - 最多备份的日志文件数
//	maxAge - 最长保留的日志文件时间（天）
//
// 返回值:
//
//	日志的写入器
func getLogWriter(filename string, maxSize, maxBackup, maxAge int) zapcore.WriteSyncer {
	// 创建lumberjack.Logger实例
	lumberJackLogger := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    maxSize,
		MaxBackups: maxBackup,
		MaxAge:     maxAge,
	}
	// 返回添加了同步写入功能的日志写入器
	return zapcore.AddSync(lumberJackLogger)
}

// GinLogger 接收gin框架默认的日志
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录请求开始时间
		start := time.Now()
		// 获取请求路径和查询参数
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		// 继续执行后续的中间件和处理器
		c.Next()

		// 计算请求耗时
		cost := time.Since(start)
		// 使用全局logger记录请求信息
		LG.Info(path,
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()),
			zap.Duration("cost", cost),
		)
	}
}

// GinRecovery recover掉项目可能出现的panic，并使用zap记录相关日志
func GinRecovery(stack bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 延迟执行recover
		defer func() {
			if err := recover(); err != nil {
				// 检查是否为断开连接错误
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				// 获取HTTP请求信息
				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				// 如果是断开连接错误
				if brokenPipe {
					LG.Error(c.Request.URL.Path,
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
					// 如果连接断开，无法写入状态
					c.Error(err.(error)) // nolint: errcheck
					c.Abort()
					return
				}

				// 如果需要记录堆栈信息
				if stack {
					LG.Error("[Recovery from panic]",
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
						zap.String("stack", string(debug.Stack())),
					)
				} else {
					LG.Error("[Recovery from panic]",
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
				}
				// 中止请求并返回500状态码
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		// 继续执行后续的中间件和处理器
		c.Next()
	}
}

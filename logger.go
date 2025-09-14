package logger

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/shuakami/logmiao/config"
	"github.com/shuakami/logmiao/formatter"
	"github.com/shuakami/logmiao/handler"
	"github.com/shuakami/logmiao/middleware"
)

var (
	// GlobalLogger 全局日志实例
	GlobalLogger *slog.Logger
	// GlobalConfig 全局配置
	GlobalConfig *config.Config
)

// Init 使用默认配置文件初始化日志系统
func Init(configPath ...string) error {
	path := "configs/logger.yaml"
	if len(configPath) > 0 && configPath[0] != "" {
		path = configPath[0]
	}
	return InitWithConfig(path)
}

// InitWithConfig 使用指定配置文件初始化日志系统
func InitWithConfig(configPath string) error {
	// 加载配置
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		// 使用默认配置
		cfg = config.LoadConfigWithDefaults(configPath)
	}
	GlobalConfig = cfg

	// 初始化日志系统
	logger, err := createLogger(cfg)
	if err != nil {
		return err
	}

	// 设置为全局默认日志器
	slog.SetDefault(logger)
	GlobalLogger = logger

	// 重定向Gin日志
	if cfg.Logger.Features.SmartFilter {
		gin.DefaultWriter = handler.NewGinLogWriter(true)
		gin.DefaultErrorWriter = handler.NewGinLogWriter(true)
	}

	return nil
}

// InitWithDefaults 使用默认配置初始化日志系统
func InitWithDefaults() error {
	cfg := config.LoadConfigWithDefaults("")
	GlobalConfig = cfg

	logger, err := createLogger(cfg)
	if err != nil {
		return err
	}

	slog.SetDefault(logger)
	GlobalLogger = logger
	return nil
}

// createLogger 根据配置创建日志器
func createLogger(cfg *config.Config) (*slog.Logger, error) {
	var handlers []slog.Handler

	// 解析日志级别
	level := parseLogLevel(cfg.Logger.Level)
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
	}

	// 1. 创建控制台处理器
	if cfg.Logger.Output.Console.Enabled {
		var consoleHandler slog.Handler
		switch cfg.Logger.Output.Console.Format {
		case "color":
			consoleHandler = handler.NewColorHandlerWithOptions(
				os.Stderr,
				opts,
				cfg.Logger.Features.KeywordHighlight,
				false, // 不使用紧凑模式
			)
		case "json":
			consoleHandler = slog.NewJSONHandler(os.Stderr, opts)
		default: // text
			consoleHandler = slog.NewTextHandler(os.Stderr, opts)
		}

		// 如果启用了智能过滤，包装处理器
		if cfg.Logger.Features.SmartFilter {
			filterConfig := handler.FilterConfig{
				IgnoreGinDebug:    true,
				IgnoreHealthCheck: true,
				MinLevel:          level,
			}
			consoleHandler = handler.NewSmartFilterHandler(consoleHandler, filterConfig)
		}

		handlers = append(handlers, consoleHandler)
	}

	// 2. 创建文件处理器
	if cfg.Logger.Output.File.Enabled {
		// 确保日志目录存在
		logDir := filepath.Dir(cfg.Logger.Output.File.Path)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return nil, err
		}

		// 创建文件写入器（带轮转）
		fileWriter := &lumberjack.Logger{
			Filename:   cfg.Logger.Output.File.Path,
			MaxSize:    cfg.Logger.Output.File.Rotation.MaxSize, // MB
			MaxBackups: cfg.Logger.Output.File.Rotation.MaxBackups,
			MaxAge:     cfg.Logger.Output.File.Rotation.MaxAge, // days
			Compress:   cfg.Logger.Output.File.Rotation.Compress,
		}

		var fileHandler slog.Handler
		switch cfg.Logger.Output.File.Format {
		case "json":
			fileHandler = slog.NewJSONHandler(fileWriter, opts)
		default: // text
			fileHandler = slog.NewTextHandler(fileWriter, opts)
		}

		// 文件日志通常不需要智能过滤，保留所有信息用于调试
		handlers = append(handlers, fileHandler)
	}

	// 3. 创建多路分发处理器
	if len(handlers) == 0 {
		// 如果没有配置任何处理器，使用默认控制台处理器
		handlers = append(handlers, handler.NewColorHandler(os.Stderr, opts))
	}

	var finalHandler slog.Handler
	if len(handlers) == 1 {
		finalHandler = handlers[0]
	} else {
		finalHandler = NewMultiHandler(handlers...)
	}

	return slog.New(finalHandler), nil
}

// parseLogLevel 解析日志级别字符串
func parseLogLevel(levelStr string) slog.Level {
	switch levelStr {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// PrintBanner 打印应用启动横幅
func PrintBanner(appName, version string) {
	if GlobalConfig != nil {
		formatter.PrintBanner(appName, version, GlobalConfig)
	}
}

// PrintStartupSuccess 打印启动成功消息
func PrintStartupSuccess(port string) {
	formatter.PrintStartupSuccess(port)
}

// GinMiddleware 返回Gin框架的日志中间件
func GinMiddleware() gin.HandlerFunc {
	return middleware.GinMiddleware()
}

// GinMiddlewareWithConfig 返回带配置的Gin框架日志中间件
func GinMiddlewareWithConfig(cfg middleware.GinMiddlewareConfig) gin.HandlerFunc {
	return middleware.GinMiddlewareWithConfig(cfg)
}

// RequestID 返回请求ID中间件
func RequestID() gin.HandlerFunc {
	return middleware.RequestID()
}

// Recovery 返回带日志记录的恢复中间件
func Recovery() gin.HandlerFunc {
	return middleware.Recovery()
}

// Error 创建错误属性，用于结构化错误记录
func Error(err error) slog.Attr {
	if err == nil {
		return slog.String("error", "")
	}
	return slog.String("error", err.Error())
}

// ErrorWithStack 创建带堆栈的错误属性
func ErrorWithStack(err error, stack string) slog.Attr {
	if err == nil {
		return slog.String("error", "")
	}
	return slog.Group("error",
		slog.String("message", err.Error()),
		slog.String("stack", stack),
	)
}

// MultiHandler 多路分发处理器
type MultiHandler struct {
	handlers []slog.Handler
}

// NewMultiHandler 创建多路分发处理器
func NewMultiHandler(handlers ...slog.Handler) *MultiHandler {
	return &MultiHandler{handlers: handlers}
}

func (h *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	// 只要有一个处理器启用，就启用
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h *MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	// 将记录分发给所有处理器
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, r.Level) {
			// 克隆记录以避免并发问题
			recordClone := r.Clone()
			if err := handler.Handle(ctx, recordClone); err != nil {
				// 记录处理错误，但继续处理其他处理器
				slog.Default().Error("Handler error", "error", err)
			}
		}
	}
	return nil
}

func (h *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithAttrs(attrs)
	}
	return &MultiHandler{handlers: newHandlers}
}

func (h *MultiHandler) WithGroup(name string) slog.Handler {
	newHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithGroup(name)
	}
	return &MultiHandler{handlers: newHandlers}
}

// GetLogger 获取当前的日志器实例
func GetLogger() *slog.Logger {
	if GlobalLogger != nil {
		return GlobalLogger
	}
	return slog.Default()
}

// SetLevel 动态设置日志级别
func SetLevel(level slog.Level) {
	// 注意：slog的处理器级别在创建时设定，无法动态修改
	// 这里提供接口，实际实现需要根据具体需求调整
	slog.Info("Log level change requested", slog.String("new_level", level.String()))
}

// Flush 刷新所有处理器的缓冲区
func Flush() {
	// 对于lumberjack等文件写入器，可能需要手动刷新
	// 这里提供接口，实际实现依赖于具体的处理器类型
	slog.Debug("Logger flush requested")
}

// Close 关闭日志系统，释放资源
func Close() error {
	slog.Info("Logger is shutting down")
	// 这里可以添加清理逻辑，比如关闭文件句柄等
	return nil
}

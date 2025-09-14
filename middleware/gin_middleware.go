package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shuakami/logmiao/config"
	"github.com/shuakami/logmiao/utils"
)

// GinMiddlewareConfig Gin中间件配置
type GinMiddlewareConfig struct {
	LogBody     bool     // 是否记录请求体（仅在错误时）
	LogHeaders  bool     // 是否记录请求头
	MaxBodySize int      // 最大请求体记录大小
	SkipPaths   []string // 跳过记录的路径（如健康检查）
}

// DefaultGinMiddlewareConfig 默认配置
func DefaultGinMiddlewareConfig() GinMiddlewareConfig {
	return GinMiddlewareConfig{
		LogBody:     true,
		LogHeaders:  false,
		MaxBodySize: 2048,
		SkipPaths:   []string{"/health", "/ping", "/metrics"},
	}
}

// GinMiddleware 返回Gin框架的日志中间件
func GinMiddleware() gin.HandlerFunc {
	cfg := DefaultGinMiddlewareConfig()
	if config.GlobalConfig != nil {
		cfg.LogBody = config.GlobalConfig.Logger.Middleware.LogBody
		cfg.LogHeaders = config.GlobalConfig.Logger.Middleware.LogHeaders
		cfg.MaxBodySize = config.GlobalConfig.Logger.Middleware.MaxBodySize
	}
	return GinMiddlewareWithConfig(cfg)
}

// GinMiddlewareWithConfig 返回带配置的Gin框架日志中间件
func GinMiddlewareWithConfig(cfg GinMiddlewareConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		rawQuery := c.Request.URL.RawQuery

		// 检查是否需要跳过记录
		for _, skipPath := range cfg.SkipPaths {
			if strings.HasPrefix(path, skipPath) {
				c.Next()
				return
			}
		}

		var bodyBytes []byte
		var requestSize int64

		// 读取请求体（仅对特定方法和非文件上传）
		if cfg.LogBody && shouldLogRequestBody(c.Request.Method, c.Request.Header.Get("Content-Type")) {
			if c.Request.Body != nil {
				bodyBytes, _ = io.ReadAll(c.Request.Body)
				requestSize = int64(len(bodyBytes))
				// 重新填充请求体
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

		// 处理请求
		c.Next()

		// 计算响应时间
		latency := time.Since(start)
		status := c.Writer.Status()
		responseSize := int64(c.Writer.Size())

		// 准备日志属性
		attrs := []slog.Attr{
			slog.String("type", "http_request"),
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.Int("status", status),
			slog.Duration("latency", latency),
			slog.String("client_ip", utils.GetClientIP(c)),
			slog.String("user_agent", c.Request.UserAgent()),
			slog.Int64("request_size", requestSize),
			slog.Int64("response_size", responseSize),
		}

		if rawQuery != "" {
			attrs = append(attrs, slog.String("query", rawQuery))
		}

		// 添加缓存状态（如果有）
		if cacheStatus, exists := c.Get("cache_status"); exists {
			if status, ok := cacheStatus.(string); ok {
				attrs = append(attrs, slog.String("cache", status))
			}
		}

		// 记录请求头（如果配置了）
		if cfg.LogHeaders && (status >= 400 || slog.Default().Enabled(nil, slog.LevelDebug)) {
			headers := make(map[string]string)
			for name, values := range c.Request.Header {
				// 过滤敏感头信息
				if isSensitiveHeader(name) {
					headers[name] = "[FILTERED]"
				} else {
					headers[name] = strings.Join(values, ", ")
				}
			}
			if len(headers) > 0 {
				attrs = append(attrs, slog.Any("headers", headers))
			}
		}

		// 记录请求体（仅在错误时或调试模式）
		if cfg.LogBody && len(bodyBytes) > 0 && (status >= 400 || slog.Default().Enabled(nil, slog.LevelDebug)) {
			bodyToLog := prepareBodyForLogging(bodyBytes, cfg.MaxBodySize)
			attrs = append(attrs, slog.String("request_body", bodyToLog))
		}

		// 记录错误信息
		if len(c.Errors) > 0 {
			errorMessages := make([]string, len(c.Errors))
			for i, err := range c.Errors {
				errorMessages[i] = err.Error()
			}
			attrs = append(attrs, slog.String("errors", strings.Join(errorMessages, "; ")))
		}

		// 根据状态码选择日志级别
		level := getLogLevelForStatus(status)
		message := "HTTP Request"

		// 添加请求标识符（如果有）
		if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
			attrs = append(attrs, slog.String("request_id", requestID))
		}

		slog.LogAttrs(c.Request.Context(), level, message, attrs...)
	}
}

// shouldLogRequestBody 检查是否应该记录请求体
func shouldLogRequestBody(method, contentType string) bool {
	// 只对可能有请求体的方法记录
	if method != "POST" && method != "PUT" && method != "PATCH" {
		return false
	}

	// 不记录文件上传
	if strings.HasPrefix(contentType, "multipart/form-data") {
		return false
	}

	return true
}

// isSensitiveHeader 检查是否是敏感的HTTP头
func isSensitiveHeader(name string) bool {
	sensitiveHeaders := []string{
		"authorization",
		"cookie",
		"x-api-key",
		"x-auth-token",
		"x-access-token",
		"x-csrf-token",
	}

	nameLower := strings.ToLower(name)
	for _, sensitive := range sensitiveHeaders {
		if nameLower == sensitive {
			return true
		}
	}
	return false
}

// prepareBodyForLogging 准备用于日志记录的请求体
func prepareBodyForLogging(bodyBytes []byte, maxSize int) string {
	if len(bodyBytes) == 0 {
		return ""
	}

	bodyStr := string(bodyBytes)

	// 限制大小
	if len(bodyStr) > maxSize {
		bodyStr = bodyStr[:maxSize] + "...(truncated)"
	}

	// 美化JSON格式（如果是JSON的话）
	trimmed := strings.TrimSpace(bodyStr)
	if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		return formatJSONBody(trimmed)
	}

	return bodyStr
}

// formatJSONBody 格式化JSON请求体用于日志记录
func formatJSONBody(jsonStr string) string {
	var obj interface{}
	if err := json.Unmarshal([]byte(jsonStr), &obj); err != nil {
		// 不是有效的JSON，直接返回原文
		return jsonStr
	}

	// 紧凑格式化，避免日志过长
	compactJSON, err := json.Marshal(obj)
	if err != nil {
		return jsonStr
	}

	return string(compactJSON)
}

// getLogLevelForStatus 根据HTTP状态码获取日志级别
func getLogLevelForStatus(status int) slog.Level {
	switch {
	case status >= 500:
		return slog.LevelError
	case status >= 400:
		return slog.LevelWarn
	case status >= 300:
		return slog.LevelInfo
	default:
		return slog.LevelInfo
	}
}

// RequestID 中间件，为每个请求添加唯一标识符
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = utils.GenerateRequestID()
			c.Header("X-Request-ID", requestID)
		}
		c.Set("request_id", requestID)
		c.Next()
	}
}

// Recovery 带日志记录的恢复中间件
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		slog.Error("Panic recovered",
			slog.String("type", "panic"),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.String("client_ip", utils.GetClientIP(c)),
			slog.Any("error", recovered),
			slog.String("user_agent", c.Request.UserAgent()),
		)
		c.AbortWithStatus(500)
	})
}

package handler

import (
	"context"
	"log/slog"
	"regexp"
	"strings"
	"sync"
	"time"
)

// SmartFilterHandler 智能过滤处理器，过滤不需要的日志消息
type SmartFilterHandler struct {
	handler           slog.Handler
	ignoreGinDebug    bool
	ignoreHealthCheck bool
	minLevel          slog.Level

	// 预编译的正则表达式，提高性能
	ginDebugRegex         *regexp.Regexp
	cookiePartitionRegex  *regexp.Regexp
	healthCheckRegex      *regexp.Regexp
	chromedpInternalRegex *regexp.Regexp

	// 重复错误检测
	errorTracker map[string]time.Time
	errorMutex   sync.RWMutex
	errorWindow  time.Duration // 错误去重时间窗口
}

// FilterConfig 过滤器配置
type FilterConfig struct {
	IgnoreGinDebug    bool       // 过滤Gin调试信息
	IgnoreHealthCheck bool       // 过滤健康检查请求
	MinLevel          slog.Level // 最低日志级别
}

// NewSmartFilterHandler 创建智能过滤处理器
func NewSmartFilterHandler(handler slog.Handler, config FilterConfig) *SmartFilterHandler {
	return &SmartFilterHandler{
		handler:           handler,
		ignoreGinDebug:    config.IgnoreGinDebug,
		ignoreHealthCheck: config.IgnoreHealthCheck,
		minLevel:          config.MinLevel,

		// 预编译正则表达式
		ginDebugRegex:         regexp.MustCompile(`^\[GIN-debug\]|\[GIN\]`),
		cookiePartitionRegex:  regexp.MustCompile(`could not unmarshal event.*CookiePartitionKey`),
		healthCheckRegex:      regexp.MustCompile(`/health|/ping|/status|/metrics`),
		chromedpInternalRegex: regexp.MustCompile(`chromedp: could not retrieve|context deadline exceeded.*chromedp`),

		// 重复错误检测配置
		errorTracker: make(map[string]time.Time),
		errorWindow:  5 * time.Minute, // 5分钟内的相同错误只记录一次
	}
}

func (h *SmartFilterHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.minLevel && h.handler.Enabled(ctx, level)
}

func (h *SmartFilterHandler) Handle(ctx context.Context, r slog.Record) error {
	// 1. 级别过滤
	if r.Level < h.minLevel {
		return nil
	}

	msg := r.Message

	// 2. 过滤Gin调试信息
	if h.ignoreGinDebug && h.ginDebugRegex.MatchString(msg) {
		return nil
	}

	// 3. 过滤CookiePartitionKey错误（chromedp内部错误）
	if h.cookiePartitionRegex.MatchString(msg) {
		return nil
	}

	// 4. 过滤chromedp内部错误
	if h.chromedpInternalRegex.MatchString(msg) {
		return nil
	}

	// 5. 过滤健康检查请求
	if h.ignoreHealthCheck && h.shouldIgnoreHealthCheck(r) {
		return nil
	}

	// 6. 过滤重复的上下文取消错误
	if h.isDuplicateContextError(msg) {
		return nil
	}

	// 7. 过滤空消息或只包含空白字符的消息
	if strings.TrimSpace(msg) == "" {
		return nil
	}

	return h.handler.Handle(ctx, r)
}

// shouldIgnoreHealthCheck 检查是否应该忽略健康检查相关的日志
func (h *SmartFilterHandler) shouldIgnoreHealthCheck(r slog.Record) bool {
	// 检查消息内容
	if h.healthCheckRegex.MatchString(r.Message) {
		return true
	}

	// 检查属性中的路径
	shouldIgnore := false
	r.Attrs(func(a slog.Attr) bool {
		if a.Key == "path" || a.Key == "url" {
			if h.healthCheckRegex.MatchString(a.Value.String()) {
				shouldIgnore = true
				return false // 停止迭代
			}
		}
		return true
	})

	return shouldIgnore
}

// isDuplicateContextError 检查是否是重复的上下文错误
func (h *SmartFilterHandler) isDuplicateContextError(msg string) bool {
	contextErrors := []string{
		"context canceled",
		"context deadline exceeded",
		"connection reset by peer",
		"broken pipe",
	}

	msgLower := strings.ToLower(msg)
	for _, errMsg := range contextErrors {
		if strings.Contains(msgLower, errMsg) {
			// 基于时间窗口的重复检测
			return h.shouldFilterDuplicateError(msg)
		}
	}

	return false
}

// shouldFilterDuplicateError 判断是否应该过滤重复错误
func (h *SmartFilterHandler) shouldFilterDuplicateError(msg string) bool {
	now := time.Now()

	h.errorMutex.Lock()
	defer h.errorMutex.Unlock()

	// 清理过期的错误记录
	for key, timestamp := range h.errorTracker {
		if now.Sub(timestamp) > h.errorWindow {
			delete(h.errorTracker, key)
		}
	}

	// 检查当前错误是否在时间窗口内已记录过
	if lastTime, exists := h.errorTracker[msg]; exists {
		if now.Sub(lastTime) < h.errorWindow {
			return true // 过滤重复错误
		}
	}

	// 记录新的错误
	h.errorTracker[msg] = now
	return false
}

func (h *SmartFilterHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &SmartFilterHandler{
		handler:               h.handler.WithAttrs(attrs),
		ignoreGinDebug:        h.ignoreGinDebug,
		ignoreHealthCheck:     h.ignoreHealthCheck,
		minLevel:              h.minLevel,
		ginDebugRegex:         h.ginDebugRegex,
		cookiePartitionRegex:  h.cookiePartitionRegex,
		healthCheckRegex:      h.healthCheckRegex,
		chromedpInternalRegex: h.chromedpInternalRegex,
		errorTracker:          h.errorTracker, // 共享错误追踪器
		errorMutex:            h.errorMutex,   // 共享互斥锁
		errorWindow:           h.errorWindow,  // 共享时间窗口
	}
}

func (h *SmartFilterHandler) WithGroup(name string) slog.Handler {
	return &SmartFilterHandler{
		handler:               h.handler.WithGroup(name),
		ignoreGinDebug:        h.ignoreGinDebug,
		ignoreHealthCheck:     h.ignoreHealthCheck,
		minLevel:              h.minLevel,
		ginDebugRegex:         h.ginDebugRegex,
		cookiePartitionRegex:  h.cookiePartitionRegex,
		healthCheckRegex:      h.healthCheckRegex,
		chromedpInternalRegex: h.chromedpInternalRegex,
		errorTracker:          h.errorTracker, // 共享错误追踪器
		errorMutex:            h.errorMutex,   // 共享互斥锁
		errorWindow:           h.errorWindow,  // 共享时间窗口
	}
}

// GinLogWriter 实现 io.Writer 接口，用于重定向 Gin 的日志输出
type GinLogWriter struct {
	ignoreDebug bool
}

// NewGinLogWriter 创建 Gin 日志写入器
func NewGinLogWriter(ignoreDebug bool) *GinLogWriter {
	return &GinLogWriter{
		ignoreDebug: ignoreDebug,
	}
}

func (w *GinLogWriter) Write(p []byte) (n int, err error) {
	msg := strings.TrimRight(string(p), "\n")
	level := slog.LevelInfo

	// 根据关键字判断日志级别
	if strings.Contains(msg, "[WARNING]") {
		level = slog.LevelWarn
		msg = strings.ReplaceAll(msg, "[WARNING]", "")
	} else if strings.Contains(msg, "[ERROR]") {
		level = slog.LevelError
		msg = strings.ReplaceAll(msg, "[ERROR]", "")
	}

	// 清理 Gin 标记和时间戳
	msg = regexp.MustCompile(`^\[GIN-debug\]|\[GIN\]`).ReplaceAllString(msg, "")
	msg = regexp.MustCompile(`\d{4}/\d{2}/\d{2} - \d{2}:\d{2}:\d{2}`).ReplaceAllString(msg, "")
	msg = strings.TrimSpace(msg)

	// 过滤调试信息
	if w.ignoreDebug && strings.Contains(msg, "[GIN-debug]") {
		return len(p), nil
	}

	// 只有在消息不为空时才记录
	if msg != "" {
		slog.Log(context.Background(), level, msg, slog.String("source", "gin"))
	}

	return len(p), nil
}

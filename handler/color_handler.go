package handler

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

var (
	// 定义关键字和对应的颜色函数
	keywordColors = map[string]func(string, ...interface{}) string{
		// 状态 - 成功 (绿色)
		"success":      color.HiGreenString,
		"successfully": color.HiGreenString,
		"loaded":       color.HiGreenString,
		"completed":    color.HiGreenString,
		"established":  color.HiGreenString,
		"initialized":  color.HiGreenString,
		"enabled":      color.HiGreenString,
		"ok":           color.HiGreenString,
		"finished":     color.HiGreenString,
		"resolved":     color.HiGreenString,
		"found":        color.HiGreenString,
		"true":         color.HiGreenString,
		"connected":    color.HiGreenString,

		// 状态 - 失败 (红色)
		"error":        color.HiRedString,
		"failed":       color.HiRedString,
		"fail":         color.HiRedString,
		"invalid":      color.HiRedString,
		"unable":       color.HiRedString,
		"cannot":       color.HiRedString,
		"denied":       color.HiRedString,
		"rejected":     color.HiRedString,
		"timeout":      color.HiRedString,
		"panic":        color.HiRedString,
		"false":        color.HiRedString,
		"disconnected": color.HiRedString,
		"crashed":      color.HiRedString,

		// 状态 - 警告 (黄色)
		"warn":       color.HiYellowString,
		"warning":    color.HiYellowString,
		"deprecated": color.HiYellowString,
		"fallback":   color.HiYellowString,
		"skipping":   color.HiYellowString,
		"missing":    color.HiYellowString,
		"empty":      color.HiYellowString,
		"ignored":    color.HiYellowString,
		"retrying":   color.HiYellowString,

		// 操作 (蓝色/青色)
		"info":       color.HiBlueString,
		"request":    color.HiBlueString,
		"response":   color.HiBlueString,
		"starting":   color.HiBlueString,
		"stopping":   color.HiBlueString,
		"connecting": color.HiBlueString,
		"sending":    color.HiBlueString,
		"receiving":  color.HiBlueString,
		"processing": color.HiBlueString,
		"get":        color.HiCyanString,
		"post":       color.HiCyanString,
		"put":        color.HiCyanString,
		"delete":     color.HiCyanString,
		"patch":      color.HiCyanString,

		// 对象 (品红色)
		"debug":      color.HiMagentaString,
		"parsing":    color.HiMagentaString,
		"generating": color.HiMagentaString,
		"writing":    color.HiMagentaString,
		"reading":    color.HiMagentaString,
		"database":   color.HiMagentaString,
		"cache":      color.HiMagentaString,
		"router":     color.HiMagentaString,
		"server":     color.HiMagentaString,
		"client":     color.HiMagentaString,
		"user":       color.HiMagentaString,
		"session":    color.HiMagentaString,
	}

	// 用于高亮数字的正则表达式
	numericRegex = regexp.MustCompile(`\b\d+(\.\d+)?(ms|s|MB|KB|GB)?\b`)
	// 用于高亮URL的正则表达式
	urlRegex = regexp.MustCompile(`https?://[^\s]+`)
	// 用于高亮IP地址的正则表达式
	ipRegex = regexp.MustCompile(`\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b`)
	// 用于高亮关键字的预编译正则表达式
	keywordRegexes map[string]*regexp.Regexp
)

func init() {
	keywordRegexes = make(map[string]*regexp.Regexp, len(keywordColors))
	for keyword := range keywordColors {
		keywordRegexes[keyword] = regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(keyword) + `\b`)
	}
}

// colorize 通过高亮关键字和数字来美化消息
func colorize(msg string, enableHighlight bool) string {
	if !enableHighlight {
		return msg
	}

	// 1. 高亮关键字
	for keyword, colorFunc := range keywordColors {
		if re, ok := keywordRegexes[keyword]; ok {
			msg = re.ReplaceAllStringFunc(msg, func(match string) string {
				return colorFunc(match)
			})
		}
	}

	// 2. 高亮数字和单位
	msg = numericRegex.ReplaceAllStringFunc(msg, func(match string) string {
		return color.New(color.FgHiWhite, color.Bold).Sprint(match)
	})

	// 3. 高亮URL
	msg = urlRegex.ReplaceAllStringFunc(msg, func(match string) string {
		return color.New(color.FgCyan, color.Underline).Sprint(match)
	})

	// 4. 高亮IP地址
	msg = ipRegex.ReplaceAllStringFunc(msg, func(match string) string {
		return color.New(color.FgYellow).Sprint(match)
	})

	return msg
}

// ColorHandler 彩色日志处理器，提供美观的控制台输出
type ColorHandler struct {
	w               io.Writer
	opts            *slog.HandlerOptions
	levelColors     map[slog.Level]*color.Color
	mu              sync.Mutex
	lastLogTime     time.Time
	enableHighlight bool
	compactMode     bool
}

// NewColorHandler 创建新的彩色处理器
func NewColorHandler(w io.Writer, opts *slog.HandlerOptions) *ColorHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}

	return &ColorHandler{
		w:               w,
		opts:            opts,
		enableHighlight: true,
		compactMode:     false,
		levelColors: map[slog.Level]*color.Color{
			slog.LevelDebug: color.New(color.FgHiWhite),
			slog.LevelInfo:  color.New(color.FgGreen),
			slog.LevelWarn:  color.New(color.FgYellow),
			slog.LevelError: color.New(color.FgRed),
		},
	}
}

// NewColorHandlerWithOptions 创建带选项的彩色处理器
func NewColorHandlerWithOptions(w io.Writer, opts *slog.HandlerOptions, enableHighlight, compactMode bool) *ColorHandler {
	handler := NewColorHandler(w, opts)
	handler.enableHighlight = enableHighlight
	handler.compactMode = compactMode
	return handler
}

func (h *ColorHandler) Enabled(ctx context.Context, level slog.Level) bool {
	minLevel := slog.LevelInfo
	if h.opts != nil && h.opts.Level != nil {
		minLevel = h.opts.Level.Level()
	}
	return level >= minLevel
}

func (h *ColorHandler) Handle(ctx context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	now := time.Now()
	// 如果距离上一条日志超过200毫秒，就加一个空行作为视觉分割
	if !h.compactMode && !h.lastLogTime.IsZero() && now.Sub(h.lastLogTime) > 200*time.Millisecond {
		fmt.Fprintln(h.w)
	}
	h.lastLogTime = now

	// 获取级别颜色
	levelColor := h.levelColors[r.Level]
	if levelColor == nil {
		levelColor = color.New(color.FgWhite)
	}

	// 输出日志级别和时间
	if h.compactMode {
		levelColor.Fprintf(h.w, "[%s]", r.Level)
		fmt.Fprintf(h.w, " %s", r.Time.Format("15:04:05.000"))
	} else {
		levelColor.Fprintf(h.w, "[%s]", r.Level)
		fmt.Fprintf(h.w, " %s", r.Time.Format("2006-01-02 15:04:05.000"))
	}

	// 对消息进行关键字高亮
	colorizedMessage := colorize(r.Message, h.enableHighlight)
	fmt.Fprintf(h.w, " %s", colorizedMessage)

	// 处理结构化属性
	attrs := make([]slog.Attr, 0)
	r.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})

	if len(attrs) > 0 {
		fmt.Fprintln(h.w) // 换行
		for _, attr := range attrs {
			h.handleAttr(attr, 1)
		}
	} else {
		fmt.Fprintln(h.w) // 结束当前日志行
	}

	return nil
}

// handleAttr 处理结构化属性
func (h *ColorHandler) handleAttr(a slog.Attr, indent int) {
	keyColor := color.New(color.FgCyan)
	defaultValColor := color.New(color.FgWhite)

	indentStr := strings.Repeat("    ", indent) // 4个空格缩进

	// 1. 处理特殊的错误和堆栈信息
	if a.Key == "error" || a.Key == "stack" || a.Key == "trace" {
		errorColor := color.New(color.FgHiRed)
		errorColor.Fprintf(h.w, "%s%s:\n", indentStr, a.Key)
		valStr := a.Value.String()
		for _, line := range splitLines(valStr) {
			if line != "" {
				errorColor.Fprintf(h.w, "%s    %s\n", indentStr, line)
			}
		}
		return
	}

	// 2. 处理特殊字段的彩色输出
	keyColor.Fprintf(h.w, "%s%s: ", indentStr, a.Key)

	valStr := a.Value.String()
	handled := true

	switch a.Key {
	case "method":
		color.New(color.FgHiBlue, color.Bold).Fprintln(h.w, valStr)
	case "status", "status_code":
		if status, err := strconv.Atoi(valStr); err == nil {
			switch {
			case status >= 500:
				color.New(color.FgRed, color.Bold).Fprintln(h.w, valStr)
			case status >= 400:
				color.New(color.FgYellow, color.Bold).Fprintln(h.w, valStr)
			case status >= 200:
				color.New(color.FgGreen, color.Bold).Fprintln(h.w, valStr)
			default:
				defaultValColor.Fprintln(h.w, valStr)
			}
		} else {
			defaultValColor.Fprintln(h.w, valStr)
		}
	case "duration", "latency":
		color.New(color.FgMagenta).Fprintln(h.w, valStr)
	case "url", "path":
		color.New(color.FgCyan, color.Underline).Fprintln(h.w, valStr)
	case "ip", "client_ip":
		color.New(color.FgYellow).Fprintln(h.w, valStr)
	case "cache", "cache_status":
		if valStr == "HIT" {
			color.New(color.FgGreen).Fprintln(h.w, valStr)
		} else if valStr == "MISS" {
			color.New(color.FgYellow).Fprintln(h.w, valStr)
		} else {
			color.New(color.FgMagenta).Fprintln(h.w, valStr)
		}
	case "user_id", "session_id":
		color.New(color.FgCyan, color.Bold).Fprintln(h.w, valStr)
	default:
		handled = false
	}

	// 3. 处理普通字段和分组
	if !handled {
		if a.Value.Kind() == slog.KindGroup {
			fmt.Fprintln(h.w) // 换行
			attrs := a.Value.Group()
			for _, ga := range attrs {
				h.handleAttr(ga, indent+1)
			}
		} else {
			// 应用关键字高亮到值
			colorizedValue := colorize(valStr, h.enableHighlight)
			fmt.Fprintln(h.w, colorizedValue)
		}
	}
}

func (h *ColorHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *ColorHandler) WithGroup(name string) slog.Handler {
	return h
}

// splitLines 分割多行字符串
func splitLines(s string) []string {
	// 同时处理字面换行符 (\n) 和转义的换行符 (\\n)
	// 首先将 "\\n" 替换为 "\n"，然后按 "\n" 分割
	return strings.Split(strings.ReplaceAll(s, `\n`, "\n"), "\n")
}

// SetCompactMode 设置紧凑模式
func (h *ColorHandler) SetCompactMode(compact bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.compactMode = compact
}

// SetHighlightEnabled 设置是否启用关键字高亮
func (h *ColorHandler) SetHighlightEnabled(enabled bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.enableHighlight = enabled
}

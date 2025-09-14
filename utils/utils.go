package utils

import (
	"crypto/rand"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shuakami/logmiao/config"
)

// GetClientIP 获取客户端真实IP地址
func GetClientIP(c *gin.Context) string {
	// 检查 X-Forwarded-For 头
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For 可能包含多个IP，取第一个
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if isValidIP(ip) {
				return ip
			}
		}
	}

	// 检查 X-Real-IP 头
	if xri := c.GetHeader("X-Real-IP"); xri != "" && isValidIP(xri) {
		return xri
	}

	// 检查 CF-Connecting-IP 头（Cloudflare）
	if cfip := c.GetHeader("CF-Connecting-IP"); cfip != "" && isValidIP(cfip) {
		return cfip
	}

	// 使用 RemoteAddr
	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return c.Request.RemoteAddr
	}
	return ip
}

// isValidIP 检查IP地址是否有效
func isValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

// GenerateRequestID 生成唯一的请求ID
func GenerateRequestID() string {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		// 如果随机数生成失败，使用时间戳
		return fmt.Sprintf("req_%d", getCurrentTimestamp())
	}
	return fmt.Sprintf("req_%x", bytes)
}

// getCurrentTimestamp 获取当前时间戳（纳秒）
func getCurrentTimestamp() int64 {
	return time.Now().UnixNano()
}

// MaskEmail 脱敏邮箱地址 - 根据配置决定是否脱敏
func MaskEmail(email string) string {
	if email == "" {
		return ""
	}

	// 检查是否启用邮箱脱敏
	if config.GlobalConfig != nil && !config.GlobalConfig.Logger.Features.Privacy.EnableEmailMask {
		return email // 不脱敏，直接返回原值
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return maskString(email)
	}

	username := parts[0]
	domain := parts[1]

	// 用户名部分脱敏：保留前2位和后0位，中间用*替换
	if len(username) <= 2 {
		return username + "@" + domain // 太短不脱敏
	} else if len(username) <= 4 {
		return username[:2] + "**@" + domain
	} else {
		starCount := len(username) - 2
		return username[:2] + strings.Repeat("*", starCount) + "@" + domain
	}
}

// MaskPhone 脱敏手机号 - 根据配置决定是否脱敏
func MaskPhone(phone string) string {
	if phone == "" {
		return ""
	}

	// 检查是否启用手机号脱敏
	if config.GlobalConfig != nil && !config.GlobalConfig.Logger.Features.Privacy.EnablePhoneMask {
		return phone // 不脱敏，直接返回原值
	}

	if len(phone) < 7 {
		if len(phone) <= 2 {
			return strings.Repeat("*", len(phone))
		} else if len(phone) <= 4 {
			return phone[:2] + strings.Repeat("*", len(phone)-2)
		}
		return phone[:2] + "***"
	}

	// 11位手机号：保留前3位和后4位
	if len(phone) == 11 && !strings.HasPrefix(phone, "+") {
		return phone[:3] + "****" + phone[7:]
	}

	// 带国际区号的号码：如+8613812345678
	if strings.HasPrefix(phone, "+") && len(phone) > 10 {
		// 保留国际区号(前3位)+****+后4位
		if len(phone) >= 10 {
			return phone[:3] + "****" + phone[len(phone)-4:]
		}
	}

	// 通用脱敏：保留前2位和后2位
	if len(phone) <= 4 {
		return maskString(phone)
	}
	return phone[:2] + strings.Repeat("*", len(phone)-4) + phone[len(phone)-2:]
}

// maskString 通用字符串脱敏
func maskString(s string) string {
	if len(s) <= 2 {
		return strings.Repeat("*", len(s))
	}
	if len(s) <= 4 {
		return s[:1] + strings.Repeat("*", len(s)-1)
	}
	return s[:2] + strings.Repeat("*", len(s)-4) + s[len(s)-2:]
}

// SanitizeUserInput 清理用户输入，防止日志注入 - 根据配置决定是否清理
func SanitizeUserInput(input string) string {
	// 检查是否启用输入清理
	if config.GlobalConfig != nil && !config.GlobalConfig.Logger.Features.Privacy.EnableInputSanitize {
		return input // 不清理，直接返回原值
	}

	// 替换换行符和回车符
	input = strings.ReplaceAll(input, "\n", "\\n")
	input = strings.ReplaceAll(input, "\r", "\\r")
	input = strings.ReplaceAll(input, "\t", "\\t")

	// 限制长度
	if len(input) > 1000 {
		input = input[:1000] + "...(truncated)"
	}

	return input
}

// FormatBytes 格式化字节大小
func FormatBytes(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}
	if bytes < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	}
	if bytes < 1024*1024*1024 {
		return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
	}
	return fmt.Sprintf("%.1f GB", float64(bytes)/(1024*1024*1024))
}

// TruncateString 截断字符串到指定长度
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return strings.Repeat(".", maxLen)
	}
	return s[:maxLen-3] + "..."
}

// IsHealthCheckPath 检查是否是健康检查路径
func IsHealthCheckPath(path string) bool {
	healthPaths := []string{
		"/health",
		"/ping",
		"/status",
		"/metrics",
		"/ready",
		"/live",
		"/heartbeat",
	}

	pathLower := strings.ToLower(path)
	for _, healthPath := range healthPaths {
		if strings.HasPrefix(pathLower, healthPath) {
			return true
		}
	}
	return false
}

// ExtractDomainFromURL 从URL中提取域名
func ExtractDomainFromURL(url string) string {
	// 移除协议前缀
	if strings.HasPrefix(url, "https://") {
		url = url[8:]
	} else if strings.HasPrefix(url, "http://") {
		url = url[7:]
	}

	// 提取域名部分
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		domain := parts[0]
		// 移除端口号
		if colonIndex := strings.LastIndex(domain, ":"); colonIndex != -1 {
			domain = domain[:colonIndex]
		}
		return domain
	}

	return url
}

// MapToSlogAttrs 将map[string]interface{}转换为slog属性
func MapToSlogAttrs(m map[string]interface{}) []interface{} {
	attrs := make([]interface{}, 0, len(m)*2)
	for k, v := range m {
		attrs = append(attrs, k, v)
	}
	return attrs
}

// SliceContains 检查切片是否包含指定元素
func SliceContains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// SafeString 安全地将interface{}转换为字符串
func SafeString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

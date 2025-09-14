package utils

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestGetClientIP 测试客户端IP获取功能
func TestGetClientIP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
		expectedIP string
	}{
		{
			name: "X-Forwarded-For header",
			headers: map[string]string{
				"X-Forwarded-For": "192.168.1.100, 10.0.0.1",
			},
			remoteAddr: "127.0.0.1:8080",
			expectedIP: "192.168.1.100",
		},
		{
			name: "X-Real-IP header",
			headers: map[string]string{
				"X-Real-IP": "192.168.1.200",
			},
			remoteAddr: "127.0.0.1:8080",
			expectedIP: "192.168.1.200",
		},
		{
			name: "CF-Connecting-IP header",
			headers: map[string]string{
				"CF-Connecting-IP": "203.0.113.1",
			},
			remoteAddr: "127.0.0.1:8080",
			expectedIP: "203.0.113.1",
		},
		{
			name:       "RemoteAddr fallback",
			headers:    map[string]string{},
			remoteAddr: "192.168.1.50:8080",
			expectedIP: "192.168.1.50",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// 创建测试请求
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = test.remoteAddr

			// 设置头部
			for key, value := range test.headers {
				req.Header.Set(key, value)
			}

			c.Request = req

			// 测试IP获取
			ip := GetClientIP(c)
			if ip != test.expectedIP {
				t.Errorf("Expected IP %s, got %s", test.expectedIP, ip)
			}
		})
	}
}

// TestGenerateRequestID 测试请求ID生成
func TestGenerateRequestID(t *testing.T) {
	// 生成多个请求ID，检查唯一性
	ids := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id := GenerateRequestID()

		// 检查格式
		if len(id) == 0 {
			t.Fatal("Generated request ID should not be empty")
		}

		if id[:4] != "req_" {
			t.Errorf("Request ID should start with 'req_', got: %s", id)
		}

		// 检查唯一性
		if ids[id] {
			t.Errorf("Duplicate request ID generated: %s", id)
		}
		ids[id] = true
	}
}

// TestMaskEmail 测试邮箱脱敏
func TestMaskEmail(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"test@example.com", "te**@example.com"},
		{"a@b.com", "a@b.com"}, // 太短不脱敏用户名部分
		{"very.long.email@example.com", "ve*************@example.com"},
		{"invalid-email", "in*********il"},
	}

	for _, test := range tests {
		result := MaskEmail(test.input)
		if result != test.expected {
			t.Errorf("MaskEmail(%s) = %s, expected %s", test.input, result, test.expected)
		}
	}
}

// TestMaskPhone 测试手机号脱敏
func TestMaskPhone(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"13812345678", "138****5678"},
		{"12345", "12***"}, // 短号码
		{"+8613812345678", "+86****5678"},
	}

	for _, test := range tests {
		result := MaskPhone(test.input)
		if result != test.expected {
			t.Errorf("MaskPhone(%s) = %s, expected %s", test.input, result, test.expected)
		}
	}
}

// TestSanitizeUserInput 测试用户输入清理
func TestSanitizeUserInput(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"normal text", "normal text"},
		{"text\nwith\nnewlines", "text\\nwith\\nnewlines"},
		{"text\rwith\rcarriage", "text\\rwith\\rcarriage"},
		{"text\twith\ttabs", "text\\twith\\ttabs"},
	}

	for _, test := range tests {
		result := SanitizeUserInput(test.input)
		if result != test.expected {
			t.Errorf("SanitizeUserInput(%s) = %s, expected %s", test.input, result, test.expected)
		}
	}

	// 测试长文本截断
	longInput := strings.Repeat("a", 1500) // 超过1000字符的文本
	result := SanitizeUserInput(longInput)
	if len(result) > 1014 { // 1000 + "...(truncated)" 的长度
		t.Errorf("Long input should be truncated, got length: %d", len(result))
	}
}

// TestFormatBytes 测试字节格式化
func TestFormatBytes(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, test := range tests {
		result := FormatBytes(test.input)
		if result != test.expected {
			t.Errorf("FormatBytes(%d) = %s, expected %s", test.input, result, test.expected)
		}
	}
}

// TestIsHealthCheckPath 测试健康检查路径识别
func TestIsHealthCheckPath(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"/health", true},
		{"/health/check", true},
		{"/ping", true},
		{"/status", true},
		{"/metrics", true},
		{"/api/users", false},
		{"/HOME", false},  // 大小写测试
		{"/HEALTH", true}, // 大小写不敏感
	}

	for _, test := range tests {
		result := IsHealthCheckPath(test.path)
		if result != test.expected {
			t.Errorf("IsHealthCheckPath(%s) = %t, expected %t", test.path, result, test.expected)
		}
	}
}

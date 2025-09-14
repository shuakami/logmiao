package logger

import (
	"log/slog"
	"testing"

	"github.com/shuakami/logmiao/config"
)

// TestInitWithDefaults 测试默认初始化
func TestInitWithDefaults(t *testing.T) {
	err := InitWithDefaults()
	if err != nil {
		t.Fatalf("InitWithDefaults failed: %v", err)
	}

	// 测试全局日志器是否正确设置
	if GlobalLogger == nil {
		t.Fatal("GlobalLogger should not be nil after initialization")
	}

	// 测试日志记录功能
	slog.Info("Test log message", slog.String("test", "value"))
}

// TestParseLogLevel 测试日志级别解析
func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"warning", slog.LevelWarn},
		{"error", slog.LevelError},
		{"invalid", slog.LevelInfo}, // 默认值
	}

	for _, test := range tests {
		result := parseLogLevel(test.input)
		if result != test.expected {
			t.Errorf("parseLogLevel(%s) = %v, expected %v", test.input, result, test.expected)
		}
	}
}

// TestErrorAttr 测试错误属性创建
func TestErrorAttr(t *testing.T) {
	// 测试 nil 错误
	attr := Error(nil)
	if attr.Key != "error" || attr.Value.String() != "" {
		t.Errorf("Error(nil) should return empty error attribute")
	}

	// 测试实际错误
	testErr := &TestErrorType{message: "test error"}
	attr = Error(testErr)
	if attr.Key != "error" || attr.Value.String() != "test error" {
		t.Errorf("Error should correctly format error message")
	}
}

// TestErrorType 自定义错误类型用于测试
type TestErrorType struct {
	message string
}

func (e *TestErrorType) Error() string {
	return e.message
}

// TestGetVersionInfo 测试版本信息获取
func TestGetVersionInfo(t *testing.T) {
	info := GetVersionInfo()

	expectedKeys := []string{"name", "version", "author", "license", "github"}
	for _, key := range expectedKeys {
		if _, exists := info[key]; !exists {
			t.Errorf("Version info missing key: %s", key)
		}
	}

	if info["name"] != Name {
		t.Errorf("Version info name mismatch: got %s, expected %s", info["name"], Name)
	}

	if info["version"] != Version {
		t.Errorf("Version info version mismatch: got %s, expected %s", info["version"], Version)
	}
}

// BenchmarkColorHandler 彩色处理器性能测试
func BenchmarkColorHandler(b *testing.B) {
	// 初始化日志系统
	err := InitWithDefaults()
	if err != nil {
		b.Fatalf("Failed to initialize logger: %v", err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			slog.Info("Benchmark test message",
				slog.String("key1", "value1"),
				slog.Int("key2", 42),
				slog.Duration("key3", 100),
			)
		}
	})
}

// BenchmarkJSONHandler JSON处理器性能测试
func BenchmarkJSONHandler(b *testing.B) {
	// 使用JSON格式初始化
	cfg := &config.Config{
		Logger: config.LoggerConfig{
			Level:  "info",
			Format: "json",
			Output: config.OutputConfig{
				Console: config.ConsoleConfig{
					Enabled: true,
					Format:  "json",
				},
				File: config.FileConfig{
					Enabled: false,
				},
			},
		},
	}

	logger, err := createLogger(cfg)
	if err != nil {
		b.Fatalf("Failed to create logger: %v", err)
	}

	slog.SetDefault(logger)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			slog.Info("Benchmark test message",
				slog.String("key1", "value1"),
				slog.Int("key2", 42),
				slog.Duration("key3", 100),
			)
		}
	})
}

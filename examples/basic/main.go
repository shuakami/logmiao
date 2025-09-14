package main

import (
	"log/slog"
	"time"

	logger "github.com/shuakami/logmiao"
)

func main() {
	// 1. 初始化日志系统
	if err := logger.Init("../../configs/logger.yaml"); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}

	// 2. 打印启动横幅
	logger.PrintBanner("Basic Example", "1.0.0")

	// 3. 开始记录日志
	slog.Info("应用程序启动",
		slog.String("app", "basic-example"),
		slog.String("version", "1.0.0"),
	)

	// 4. 演示不同级别的日志
	slog.Debug("这是调试信息",
		slog.String("module", "main"),
		slog.String("function", "main"),
	)

	slog.Info("用户登录成功",
		slog.String("user_id", "12345"),
		slog.String("ip", "192.168.1.100"),
		slog.Duration("login_time", 150*time.Millisecond),
	)

	slog.Warn("缓存未命中",
		slog.String("cache_key", "user:12345"),
		slog.String("fallback", "database"),
	)

	// 5. 演示错误记录
	err := processUser("invalid-user-id")
	if err != nil {
		slog.Error("用户处理失败",
			logger.Error(err),
			slog.String("user_id", "invalid-user-id"),
			slog.String("operation", "process"),
		)
	}

	// 6. 演示结构化日志
	slog.Info("API调用统计",
		slog.Group("stats",
			slog.Int("total_requests", 1234),
			slog.Int("successful", 1200),
			slog.Int("failed", 34),
			slog.Float64("success_rate", 97.24),
		),
		slog.Group("performance",
			slog.Duration("avg_response_time", 45*time.Millisecond),
			slog.Duration("max_response_time", 500*time.Millisecond),
		),
	)

	slog.Info("应用程序正常关闭")
}

// processUser 模拟用户处理函数
func processUser(userID string) error {
	if userID == "invalid-user-id" {
		return &UserProcessError{
			UserID: userID,
			Reason: "用户ID格式无效",
		}
	}
	return nil
}

// UserProcessError 自定义错误类型
type UserProcessError struct {
	UserID string
	Reason string
}

func (e *UserProcessError) Error() string {
	return "用户处理错误: " + e.Reason + " (UserID: " + e.UserID + ")"
}

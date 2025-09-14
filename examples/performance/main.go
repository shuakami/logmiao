package main

import (
	"fmt"
	"log/slog"
	"math/rand"
	"sync"
	"time"

	logger "github.com/shuakami/logmiao"
)

func main() {
	// 初始化日志系统
	if err := logger.Init("../../configs/logger.yaml"); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}

	logger.PrintBanner("Performance Test", "1.0.0")

	slog.Info("开始性能测试",
		slog.String("test_type", "concurrent_logging"),
	)

	// 测试1: 并发日志写入性能
	testConcurrentLogging()

	// 测试2: 结构化日志性能
	testStructuredLogging()

	// 测试3: 大量数据日志性能
	testBulkLogging()

	slog.Info("性能测试完成")
}

// 测试并发日志写入
func testConcurrentLogging() {
	slog.Info("开始并发日志测试")

	const (
		numGoroutines    = 100
		logsPerGoroutine = 100
	)

	start := time.Now()
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < logsPerGoroutine; j++ {
				level := rand.Intn(4)
				switch level {
				case 0:
					slog.Debug("调试消息",
						slog.Int("worker", workerID),
						slog.Int("message", j),
						slog.String("type", "debug"),
					)
				case 1:
					slog.Info("信息消息",
						slog.Int("worker", workerID),
						slog.Int("message", j),
						slog.String("type", "info"),
						slog.Duration("random_duration", time.Duration(rand.Intn(1000))*time.Millisecond),
					)
				case 2:
					slog.Warn("警告消息",
						slog.Int("worker", workerID),
						slog.Int("message", j),
						slog.String("type", "warning"),
						slog.String("reason", "模拟警告"),
					)
				case 3:
					slog.Error("错误消息",
						slog.Int("worker", workerID),
						slog.Int("message", j),
						slog.String("type", "error"),
						slog.String("error", "模拟错误"),
					)
				}
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	totalLogs := numGoroutines * logsPerGoroutine
	logsPerSecond := float64(totalLogs) / duration.Seconds()

	slog.Info("并发日志测试完成",
		slog.Int("total_logs", totalLogs),
		slog.Duration("duration", duration),
		slog.Float64("logs_per_second", logsPerSecond),
		slog.Int("goroutines", numGoroutines),
	)
}

// 测试结构化日志性能
func testStructuredLogging() {
	slog.Info("开始结构化日志测试")

	const numLogs = 1000
	start := time.Now()

	for i := 0; i < numLogs; i++ {
		slog.Info("用户操作记录",
			slog.String("user_id", generateUserID()),
			slog.String("action", generateAction()),
			slog.String("resource", generateResource()),
			slog.Int("attempt", i),
			slog.Duration("processing_time", time.Duration(rand.Intn(500))*time.Millisecond),
			slog.Group("metadata",
				slog.String("ip", generateIP()),
				slog.String("user_agent", generateUserAgent()),
				slog.String("session_id", generateSessionID()),
			),
			slog.Group("performance",
				slog.Int("cpu_usage", rand.Intn(100)),
				slog.Int("memory_usage", rand.Intn(1024)),
				slog.Int("io_operations", rand.Intn(50)),
			),
		)
	}

	duration := time.Since(start)
	logsPerSecond := float64(numLogs) / duration.Seconds()

	slog.Info("结构化日志测试完成",
		slog.Int("total_logs", numLogs),
		slog.Duration("duration", duration),
		slog.Float64("logs_per_second", logsPerSecond),
	)
}

// 测试大量数据日志
func testBulkLogging() {
	slog.Info("开始大量数据日志测试")

	const numBatches = 10
	const logsPerBatch = 500

	start := time.Now()

	for batch := 0; batch < numBatches; batch++ {
		batchStart := time.Now()

		for i := 0; i < logsPerBatch; i++ {
			// 生成大量结构化数据
			userData := generateLargeUserData()

			slog.Info("批量用户数据处理",
				slog.Int("batch", batch),
				slog.Int("record", i),
				slog.Any("user_data", userData),
				slog.String("processing_status", "completed"),
			)
		}

		batchDuration := time.Since(batchStart)
		slog.Info("批次处理完成",
			slog.Int("batch", batch),
			slog.Int("logs_in_batch", logsPerBatch),
			slog.Duration("batch_duration", batchDuration),
		)
	}

	duration := time.Since(start)
	totalLogs := numBatches * logsPerBatch
	logsPerSecond := float64(totalLogs) / duration.Seconds()

	slog.Info("大量数据日志测试完成",
		slog.Int("total_logs", totalLogs),
		slog.Int("batches", numBatches),
		slog.Duration("total_duration", duration),
		slog.Float64("logs_per_second", logsPerSecond),
	)
}

// 辅助函数：生成测试数据

func generateUserID() string {
	return "user_" + randomString(8)
}

func generateAction() string {
	actions := []string{"login", "logout", "view", "edit", "delete", "create", "update"}
	return actions[rand.Intn(len(actions))]
}

func generateResource() string {
	resources := []string{"user", "post", "comment", "file", "setting", "profile"}
	return resources[rand.Intn(len(resources))]
}

func generateIP() string {
	return fmt.Sprintf("%d.%d.%d.%d",
		rand.Intn(256), rand.Intn(256), rand.Intn(256), rand.Intn(256))
}

func generateUserAgent() string {
	agents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/91.0.4472.124",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) Firefox/89.0",
	}
	return agents[rand.Intn(len(agents))]
}

func generateSessionID() string {
	return "sess_" + randomString(16)
}

func generateLargeUserData() map[string]interface{} {
	return map[string]interface{}{
		"id":       generateUserID(),
		"username": randomString(10),
		"email":    randomString(8) + "@example.com",
		"profile": map[string]interface{}{
			"first_name": randomString(6),
			"last_name":  randomString(8),
			"age":        rand.Intn(80) + 18,
			"country":    randomString(5),
			"preferences": []string{
				randomString(4), randomString(6), randomString(5),
			},
		},
		"statistics": map[string]interface{}{
			"login_count":    rand.Intn(1000),
			"posts_created":  rand.Intn(50),
			"last_login":     time.Now().Add(-time.Duration(rand.Intn(24*7)) * time.Hour),
			"account_status": "active",
		},
		"metadata": map[string]interface{}{
			"created_at": time.Now().Add(-time.Duration(rand.Intn(365*2)) * 24 * time.Hour),
			"updated_at": time.Now().Add(-time.Duration(rand.Intn(24)) * time.Hour),
			"version":    rand.Intn(10) + 1,
		},
	}
}

func randomString(length int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = letters[rand.Intn(len(letters))]
	}
	return string(result)
}

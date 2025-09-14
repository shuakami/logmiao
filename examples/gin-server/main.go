package main

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	logger "github.com/shuakami/logmiao"
)

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func main() {
	// 1. 初始化日志系统
	if err := logger.Init("../../configs/logger.yaml"); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}

	// 2. 打印启动横幅
	logger.PrintBanner("Gin API Server", "1.0.0")

	// 3. 创建Gin实例
	r := gin.New()

	// 4. 添加中间件
	r.Use(logger.RequestID())     // 请求ID中间件
	r.Use(logger.GinMiddleware()) // 日志中间件
	r.Use(logger.Recovery())      // 恢复中间件
	r.Use(gin.Recovery())         // Gin原生恢复中间件

	// 5. 定义路由
	setupRoutes(r)

	// 6. 启动服务器
	port := "8080"
	logger.PrintStartupSuccess(port)

	slog.Info("服务器启动",
		slog.String("port", port),
		slog.String("mode", gin.Mode()),
	)

	if err := r.Run(":" + port); err != nil {
		slog.Error("服务器启动失败", logger.Error(err))
	}
}

func setupRoutes(r *gin.Engine) {
	// API v1 路由组
	v1 := r.Group("/api/v1")
	{
		// 用户相关
		users := v1.Group("/users")
		{
			users.GET("", getUsers)
			users.GET("/:id", getUserByID)
			users.POST("", createUser)
			users.PUT("/:id", updateUser)
			users.DELETE("/:id", deleteUser)
		}

		// 其他端点
		v1.GET("/health", healthCheck)
		v1.GET("/status", getStatus)
		v1.POST("/error", simulateError) // 测试错误处理
	}

	// 静态文件服务
	r.Static("/static", "./static")

	// 默认路由
	r.NoRoute(func(c *gin.Context) {
		slog.Warn("未找到路由",
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
		)
		c.JSON(404, Response{
			Code:    404,
			Message: "路由不存在",
		})
	})
}

// 健康检查
func healthCheck(c *gin.Context) {
	c.JSON(200, Response{
		Code:    200,
		Message: "服务健康",
		Data: map[string]interface{}{
			"timestamp": time.Now().Unix(),
			"status":    "healthy",
		},
	})
}

// 获取服务状态
func getStatus(c *gin.Context) {
	slog.Info("获取服务状态",
		slog.String("client_ip", c.ClientIP()),
	)

	c.JSON(200, Response{
		Code:    200,
		Message: "获取状态成功",
		Data: map[string]interface{}{
			"uptime":     time.Since(startTime).String(),
			"version":    "1.0.0",
			"go_version": "1.21+",
		},
	})
}

var startTime = time.Now()

// 获取用户列表
func getUsers(c *gin.Context) {
	// 模拟数据库查询延迟
	time.Sleep(50 * time.Millisecond)

	users := []User{
		{ID: "1", Name: "Alice", Email: "alice@example.com"},
		{ID: "2", Name: "Bob", Email: "bob@example.com"},
		{ID: "3", Name: "Charlie", Email: "charlie@example.com"},
	}

	slog.Info("用户列表查询成功",
		slog.Int("user_count", len(users)),
		slog.String("requester_ip", c.ClientIP()),
	)

	c.JSON(200, Response{
		Code:    200,
		Message: "获取用户列表成功",
		Data:    users,
	})
}

// 根据ID获取用户
func getUserByID(c *gin.Context) {
	userID := c.Param("id")

	// 模拟数据库查询
	time.Sleep(30 * time.Millisecond)

	if userID == "999" {
		slog.Warn("用户不存在",
			slog.String("user_id", userID),
			slog.String("requester_ip", c.ClientIP()),
		)
		c.JSON(404, Response{
			Code:    404,
			Message: "用户不存在",
		})
		return
	}

	user := User{
		ID:    userID,
		Name:  "User " + userID,
		Email: "user" + userID + "@example.com",
	}

	slog.Info("用户查询成功",
		slog.String("user_id", userID),
		slog.String("user_name", user.Name),
	)

	c.JSON(200, Response{
		Code:    200,
		Message: "获取用户成功",
		Data:    user,
	})
}

// 创建用户
func createUser(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		slog.Warn("用户创建失败：请求数据无效",
			logger.Error(err),
			slog.String("content_type", c.ContentType()),
		)
		c.JSON(400, Response{
			Code:    400,
			Message: "请求数据无效: " + err.Error(),
		})
		return
	}

	// 模拟数据库操作
	time.Sleep(100 * time.Millisecond)

	// 分配新ID
	user.ID = "new-" + time.Now().Format("20060102150405")

	slog.Info("用户创建成功",
		slog.String("user_id", user.ID),
		slog.String("user_name", user.Name),
		slog.String("user_email", user.Email),
	)

	c.JSON(201, Response{
		Code:    201,
		Message: "用户创建成功",
		Data:    user,
	})
}

// 更新用户
func updateUser(c *gin.Context) {
	userID := c.Param("id")

	var updateData User
	if err := c.ShouldBindJSON(&updateData); err != nil {
		slog.Warn("用户更新失败：请求数据无效",
			logger.Error(err),
			slog.String("user_id", userID),
		)
		c.JSON(400, Response{
			Code:    400,
			Message: "请求数据无效: " + err.Error(),
		})
		return
	}

	// 模拟数据库操作
	time.Sleep(80 * time.Millisecond)

	updateData.ID = userID

	slog.Info("用户更新成功",
		slog.String("user_id", userID),
		slog.String("new_name", updateData.Name),
		slog.String("new_email", updateData.Email),
	)

	c.JSON(200, Response{
		Code:    200,
		Message: "用户更新成功",
		Data:    updateData,
	})
}

// 删除用户
func deleteUser(c *gin.Context) {
	userID := c.Param("id")

	// 模拟权限检查
	if userID == "1" {
		slog.Warn("删除用户失败：权限不足",
			slog.String("user_id", userID),
			slog.String("reason", "管理员账户不能删除"),
		)
		c.JSON(403, Response{
			Code:    403,
			Message: "权限不足：管理员账户不能删除",
		})
		return
	}

	// 模拟数据库操作
	time.Sleep(60 * time.Millisecond)

	slog.Info("用户删除成功",
		slog.String("user_id", userID),
	)

	c.JSON(200, Response{
		Code:    200,
		Message: "用户删除成功",
	})
}

// 模拟错误（测试用）
func simulateError(c *gin.Context) {
	errorType := c.DefaultQuery("type", "general")

	switch errorType {
	case "panic":
		slog.Error("即将触发panic", slog.String("type", errorType))
		panic("这是一个测试panic")
	case "timeout":
		slog.Warn("模拟超时错误", slog.String("type", errorType))
		time.Sleep(5 * time.Second)
		c.JSON(500, Response{Code: 500, Message: "操作超时"})
	default:
		slog.Error("模拟一般错误",
			slog.String("type", errorType),
			slog.String("details", "这是一个测试错误"),
		)
		c.JSON(500, Response{
			Code:    500,
			Message: "内部服务器错误",
		})
	}
}

# LogMiao

[![GPL3 License](https://img.shields.io/badge/License-GPL3-4a5568?style=flat-square)](https://www.gnu.org/licenses/gpl-3.0)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go)](https://golang.org/)
[![Release](https://img.shields.io/github/v/release/shuakami/logmiao?style=flat-square&color=667eea)](https://github.com/shuakami/logmiao/releases) 
![Platform](https://img.shields.io/badge/平台-Win/Mac/Linux-48bb78?style=flat-square)

## 这是什么

LogMiao 是一个专为 Go 应用程序设计的**高性能结构化日志库**，基于 Go 1.21+ 的官方 `log/slog` 包构建，提供**生产级**的日志解决方案。

核心理念是**简单、高效、美观** - 让开发者能够快速集成并获得专业级的日志输出体验，无需复杂配置即可拥有彩色高亮、智能过滤、自动轮转等企业级特性。

## 快速开始

### 安装

```bash
go get github.com/shuakami/logmiao
```

### 基础使用（零配置启动）

```go
package main

import (
    "log/slog"
    logger "github.com/shuakami/logmiao"
)

func main() {
    // 一行代码初始化（使用内置默认配置）
    logger.InitWithDefaults()
    
    // 开始记录日志
    slog.Info("应用程序启动成功", 
        slog.String("version", "1.0.0"),
        slog.Int("port", 8080),
    )
    
    slog.Error("数据库连接失败", 
        logger.Error(fmt.Errorf("connection refused")),
        slog.String("host", "localhost:5432"),
    )
}
```

### Gin 框架集成

```go
package main

import (
    "github.com/gin-gonic/gin"
    logger "github.com/shuakami/logmiao"
)

func main() {
    // 初始化日志系统
    logger.Init("configs/logger.yaml")
    
    // 打印启动横幅
    logger.PrintBanner("My API", "1.0.0")
    
    r := gin.New()
    
    // 添加日志中间件（包含请求ID、错误恢复等）
    r.Use(logger.RequestID())
    r.Use(logger.GinMiddleware()) 
    r.Use(logger.Recovery())
    
    r.GET("/users", func(c *gin.Context) {
        slog.Info("查询用户列表", 
            slog.String("user_id", c.Query("id")),
        )
        c.JSON(200, gin.H{"users": []string{"Alice", "Bob"}})
    })
    
    logger.PrintStartupSuccess("8080")
    r.Run(":8080")
}
```

## 配置文件

### 完整配置示例 (configs/logger.yaml)

```yaml
logger:
  level: "info"                    # 日志级别: debug, info, warn, error
  format: "color"                  # 控制台格式: color, json, text
  
  output:
    console:
      enabled: true                # 启用控制台输出
      format: "color"              # 控制台专用彩色格式
    file:
      enabled: true                # 启用文件输出
      path: "logs/app.log"         # 日志文件路径
      format: "json"               # 文件格式（建议 JSON）
      rotation:
        max_size: 50               # 单文件最大大小(MB)
        max_backups: 10            # 保留备份数量
        max_age: 30                # 保留天数
        compress: true             # 压缩历史文件

  features:
    smart_filter: true             # 智能过滤（推荐开启）
    keyword_highlight: true        # 关键词高亮

  middleware:
    log_body: false                # 是否记录请求体（生产环境建议关闭）
    log_headers: false             # 是否记录请求头
    max_body_size: 1024            # 最大请求体记录大小
```

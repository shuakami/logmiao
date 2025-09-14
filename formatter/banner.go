package formatter

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/shuakami/logmiao/config"
)

// PrintBanner 打印美观的应用启动横幅
func PrintBanner(appName, version string, cfg *config.Config) {
	// 定义调色板
	titleColor := color.New(color.FgHiCyan, color.Bold)
	labelColor := color.New(color.FgWhite)
	valueColor := color.New(color.FgGreen)
	treeColor := color.New(color.FgHiBlack, color.Bold) // 暗灰色

	// ASCII艺术标题（可选）
	if appName != "" {
		titleColor.Printf("\n > %s v%s is starting...\n\n", appName, version)
	} else {
		titleColor.Printf("\n > Application v%s is starting...\n\n", version)
	}

	// 系统信息组
	treeColor.Println("  ● System")
	labelColor.Printf("    ├─ Go Version:  ")
	valueColor.Println(runtime.Version())
	labelColor.Printf("    ├─ Platform:    ")
	valueColor.Printf("%s/%s\n", runtime.GOOS, runtime.GOARCH)
	labelColor.Printf("    └─ CPUs:        ")
	valueColor.Printf("%d\n\n", runtime.NumCPU())

	// 日志配置信息组
	if cfg != nil {
		treeColor.Println("  ● Logger Config")
		labelColor.Printf("    ├─ Level:       ")
		printLogLevel(cfg.Logger.Level)
		fmt.Println()
		labelColor.Printf("    ├─ Format:      ")
		valueColor.Println(cfg.Logger.Format)
		labelColor.Printf("    ├─ Console:     ")
		if cfg.Logger.Output.Console.Enabled {
			valueColor.Printf("✓ Enabled (%s)", cfg.Logger.Output.Console.Format)
		} else {
			color.Red("✗ Disabled")
		}
		fmt.Println()
		labelColor.Printf("    └─ File:        ")
		if cfg.Logger.Output.File.Enabled {
			valueColor.Printf("✓ Enabled → %s", cfg.Logger.Output.File.Path)
		} else {
			color.Red("✗ Disabled")
		}
		fmt.Printf("\n\n")

		// 功能配置
		treeColor.Println("  ● Features")
		features := []struct {
			name    string
			enabled bool
		}{
			{"Smart Filter", cfg.Logger.Features.SmartFilter},
			{"Keyword Highlight", cfg.Logger.Features.KeywordHighlight},
			{"Performance Tracking", cfg.Logger.Features.PerformanceTracking},
			{"Auto Sampling", cfg.Logger.Features.AutoSampling},
		}

		for i, feature := range features {
			var prefix string
			if i == len(features)-1 {
				prefix = "    └─ "
			} else {
				prefix = "    ├─ "
			}

			labelColor.Printf("%s%-18s", prefix, feature.name+":")
			if feature.enabled {
				valueColor.Print("✓ ON")
			} else {
				color.New(color.FgHiBlack).Print("✗ OFF")
			}
			fmt.Println()
		}
		fmt.Println()

		// Web查看器状态
		if cfg.Logger.Viewer.Enabled {
			treeColor.Println("  ● Web Viewer")
			labelColor.Printf("    └─ URL:         ")
			color.New(color.FgCyan, color.Underline).Printf("http://localhost:%d\n\n", cfg.Logger.Viewer.Port)
		}
	}

	// 分隔线
	color.New(color.FgHiBlack).Println("  " + strings.Repeat("─", 50))
	fmt.Println()
}

// printLogLevel 根据不同的日志级别打印带颜色的标签
func printLogLevel(level string) {
	var levelColor *color.Color
	upperLevel := strings.ToUpper(level)

	switch upperLevel {
	case "DEBUG":
		levelColor = color.New(color.FgWhite, color.BgMagenta, color.Bold)
	case "INFO":
		levelColor = color.New(color.FgWhite, color.BgGreen, color.Bold)
	case "WARN", "WARNING":
		levelColor = color.New(color.FgBlack, color.BgYellow, color.Bold)
	case "ERROR":
		levelColor = color.New(color.FgWhite, color.BgRed, color.Bold)
	default:
		levelColor = color.New(color.FgWhite, color.BgBlack, color.Bold)
	}
	levelColor.Printf(" %s ", upperLevel)
}

// PrintStartupSuccess 打印启动成功消息
func PrintStartupSuccess(port string) {
	successColor := color.New(color.FgGreen, color.Bold)
	addressColor := color.New(color.FgCyan, color.Underline)

	successColor.Print("  🚀 Server is running on ")
	addressColor.Printf("http://localhost:%s\n", port)
	fmt.Println()
}

// PrintShutdownMessage 打印关闭消息
func PrintShutdownMessage() {
	shutdownColor := color.New(color.FgYellow, color.Bold)
	shutdownColor.Println("  👋 Server is shutting down gracefully...")
	fmt.Println()
}

// PrintCustomBanner 打印自定义横幅
func PrintCustomBanner(lines []string, titleColor *color.Color) {
	if titleColor == nil {
		titleColor = color.New(color.FgHiCyan, color.Bold)
	}

	fmt.Println()
	for _, line := range lines {
		titleColor.Printf("  %s\n", line)
	}
	fmt.Println()
}

// GetASCIIArt 获取预定义的ASCII艺术
func GetASCIIArt(name string) []string {
	arts := map[string][]string{
		"logger": {
			"╦  ┌─┐┌─┐┌─┐┌─┐┬─┐",
			"║  │ ││ ┬│ ┬├┤ ├┬┘",
			"╩═╝└─┘└─┘└─┘└─┘┴└─",
		},
		"go": {
			"   _____ ____",
			"  / ___// __ \\",
			" / (_ // /_/ /",
			" \\___/ \\____/",
		},
		"simple": {
			"▄▄▄▄▄▄▄▄▄▄▄▄▄",
			"█ GO LOGGER █",
			"▀▀▀▀▀▀▀▀▀▀▀▀▀",
		},
	}

	if art, exists := arts[name]; exists {
		return art
	}
	return []string{"Advanced Logger"}
}

// PrintHealthCheck 打印健康检查信息
func PrintHealthCheck(status string, details map[string]interface{}) {
	statusColor := color.New(color.FgGreen)
	if status != "healthy" {
		statusColor = color.New(color.FgRed)
	}

	fmt.Printf("  ● Health Status: ")
	statusColor.Printf("%s\n", strings.ToUpper(status))

	if len(details) > 0 {
		for key, value := range details {
			fmt.Printf("    ├─ %-15s %v\n", key+":", value)
		}
	}
	fmt.Println()
}

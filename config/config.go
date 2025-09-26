package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config 是日志系统的整体配置
type Config struct {
	Logger LoggerConfig `mapstructure:"logger"`
}

// LoggerConfig 日志配置
type LoggerConfig struct {
	Level      string           `mapstructure:"level"`      // 日志级别: debug, info, warn, error
	Format     string           `mapstructure:"format"`     // 输出格式: color, json, text
	Output     OutputConfig     `mapstructure:"output"`     // 输出配置
	Features   FeaturesConfig   `mapstructure:"features"`   // 功能配置
	Middleware MiddlewareConfig `mapstructure:"middleware"` // 中间件配置
	Viewer     ViewerConfig     `mapstructure:"viewer"`     // Web查看器配置
}

// OutputConfig 输出配置
type OutputConfig struct {
	Console ConsoleConfig `mapstructure:"console"`
	File    FileConfig    `mapstructure:"file"`
}

// ConsoleConfig 控制台输出配置
type ConsoleConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Format  string `mapstructure:"format"` // color, json, text
}

// FileConfig 文件输出配置
type FileConfig struct {
	Enabled  bool           `mapstructure:"enabled"`
	Path     string         `mapstructure:"path"`
	Format   string         `mapstructure:"format"` // json, text
	Rotation RotationConfig `mapstructure:"rotation"`
}

// RotationConfig 日志轮转配置
type RotationConfig struct {
	MaxSize    int  `mapstructure:"max_size"`    // MB
	MaxBackups int  `mapstructure:"max_backups"` // 备份文件数
	MaxAge     int  `mapstructure:"max_age"`     // 保存天数
	Compress   bool `mapstructure:"compress"`    // 压缩旧文件
}

// FeaturesConfig 功能配置
type FeaturesConfig struct {
	SmartFilter         bool          `mapstructure:"smart_filter"`         // 智能过滤
	KeywordHighlight    bool          `mapstructure:"keyword_highlight"`    // 关键词高亮
	AutoSampling        bool          `mapstructure:"auto_sampling"`        // 自动采样
	PerformanceTracking bool          `mapstructure:"performance_tracking"` // 性能追踪
	Privacy             PrivacyConfig `mapstructure:"privacy"`              // 隐私脱敏配置
}

// PrivacyConfig 隐私脱敏配置
type PrivacyConfig struct {
	EnableEmailMask     bool `mapstructure:"enable_email_mask"`     // 启用邮箱脱敏
	EnablePhoneMask     bool `mapstructure:"enable_phone_mask"`     // 启用手机号脱敏
	EnableInputSanitize bool `mapstructure:"enable_input_sanitize"` // 启用输入清理
}

// MiddlewareConfig 中间件配置
type MiddlewareConfig struct {
	LogBody     bool `mapstructure:"log_body"`      // 记录请求体
	LogHeaders  bool `mapstructure:"log_headers"`   // 记录请求头
	MaxBodySize int  `mapstructure:"max_body_size"` // 最大请求体大小
}

// ViewerConfig Web日志查看器配置
type ViewerConfig struct {
	Enabled bool       `mapstructure:"enabled"`
	Port    int        `mapstructure:"port"`
	Auth    AuthConfig `mapstructure:"auth"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

// GlobalConfig 全局配置实例
var GlobalConfig *Config

// LoadConfig 从指定的路径加载配置
func LoadConfig(path string) (*Config, error) {
	viper.SetConfigName("logger")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("../configs")

	if path != "" {
		viper.SetConfigFile(path)
	}

	// 设置默认值
	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// 配置文件未找到，使用默认配置（静默处理）
		} else {
			return nil, fmt.Errorf("读取配置文件失败: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	GlobalConfig = &config
	return &config, nil
}

// setDefaults 设置默认配置值
func setDefaults() {
	// 日志级别和格式
	viper.SetDefault("logger.level", "info")
	viper.SetDefault("logger.format", "color")

	// 控制台输出
	viper.SetDefault("logger.output.console.enabled", true)
	viper.SetDefault("logger.output.console.format", "color")

	// 文件输出
	viper.SetDefault("logger.output.file.enabled", true)
	viper.SetDefault("logger.output.file.path", "logs/app.log")
	viper.SetDefault("logger.output.file.format", "json")
	viper.SetDefault("logger.output.file.rotation.max_size", 10)
	viper.SetDefault("logger.output.file.rotation.max_backups", 5)
	viper.SetDefault("logger.output.file.rotation.max_age", 30)
	viper.SetDefault("logger.output.file.rotation.compress", true)

	// 功能配置
	viper.SetDefault("logger.features.smart_filter", true)
	viper.SetDefault("logger.features.keyword_highlight", true)
	viper.SetDefault("logger.features.auto_sampling", false)
	viper.SetDefault("logger.features.performance_tracking", true)

	// 隐私脱敏配置 - 默认全部关闭
	viper.SetDefault("logger.features.privacy.enable_email_mask", false)
	viper.SetDefault("logger.features.privacy.enable_phone_mask", false)
	viper.SetDefault("logger.features.privacy.enable_input_sanitize", false)

	// 中间件配置
	viper.SetDefault("logger.middleware.log_body", true)
	viper.SetDefault("logger.middleware.log_headers", false)
	viper.SetDefault("logger.middleware.max_body_size", 2048)

	// Web查看器配置
	viper.SetDefault("logger.viewer.enabled", false)
	viper.SetDefault("logger.viewer.port", 8081)
	viper.SetDefault("logger.viewer.auth.username", "admin")
	viper.SetDefault("logger.viewer.auth.password", "secret")
}

// LoadConfigWithDefaults 加载配置，如果文件不存在则使用默认配置
func LoadConfigWithDefaults(path string) *Config {
	config, err := LoadConfig(path)
	if err != nil {
		fmt.Printf("使用默认配置: %v\n", err)
		// 返回默认配置
		setDefaults()
		config = &Config{
			Logger: LoggerConfig{
				Level:  viper.GetString("logger.level"),
				Format: viper.GetString("logger.format"),
				Output: OutputConfig{
					Console: ConsoleConfig{
						Enabled: viper.GetBool("logger.output.console.enabled"),
						Format:  viper.GetString("logger.output.console.format"),
					},
					File: FileConfig{
						Enabled: viper.GetBool("logger.output.file.enabled"),
						Path:    viper.GetString("logger.output.file.path"),
						Format:  viper.GetString("logger.output.file.format"),
						Rotation: RotationConfig{
							MaxSize:    viper.GetInt("logger.output.file.rotation.max_size"),
							MaxBackups: viper.GetInt("logger.output.file.rotation.max_backups"),
							MaxAge:     viper.GetInt("logger.output.file.rotation.max_age"),
							Compress:   viper.GetBool("logger.output.file.rotation.compress"),
						},
					},
				},
				Features: FeaturesConfig{
					SmartFilter:         viper.GetBool("logger.features.smart_filter"),
					KeywordHighlight:    viper.GetBool("logger.features.keyword_highlight"),
					AutoSampling:        viper.GetBool("logger.features.auto_sampling"),
					PerformanceTracking: viper.GetBool("logger.features.performance_tracking"),
					Privacy: PrivacyConfig{
						EnableEmailMask:     viper.GetBool("logger.features.privacy.enable_email_mask"),
						EnablePhoneMask:     viper.GetBool("logger.features.privacy.enable_phone_mask"),
						EnableInputSanitize: viper.GetBool("logger.features.privacy.enable_input_sanitize"),
					},
				},
				Middleware: MiddlewareConfig{
					LogBody:     viper.GetBool("logger.middleware.log_body"),
					LogHeaders:  viper.GetBool("logger.middleware.log_headers"),
					MaxBodySize: viper.GetInt("logger.middleware.max_body_size"),
				},
				Viewer: ViewerConfig{
					Enabled: viper.GetBool("logger.viewer.enabled"),
					Port:    viper.GetInt("logger.viewer.port"),
					Auth: AuthConfig{
						Username: viper.GetString("logger.viewer.auth.username"),
						Password: viper.GetString("logger.viewer.auth.password"),
					},
				},
			},
		}
		GlobalConfig = config
	}
	return config
}

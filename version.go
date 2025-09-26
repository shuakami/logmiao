package logger

// Version 包版本信息
const (
	Version = "1.0.1"
	Name    = "LogMiao"
	Author  = "shuakami"
	License = "GPL-3.0"
	GitHub  = "https://github.com/shuakami/logmiao"
)

// GetVersionInfo 返回版本信息
func GetVersionInfo() map[string]string {
	return map[string]string{
		"name":    Name,
		"version": Version,
		"author":  Author,
		"license": License,
		"github":  GitHub,
	}
}

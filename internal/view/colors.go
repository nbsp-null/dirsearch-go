package view

import (
	"fmt"
	"strconv"

	"github.com/fatih/color"
)

// StatusColors 状态码颜色配置
type StatusColors struct {
	Success     *color.Color // 2xx
	Redirect    *color.Color // 3xx
	ClientError *color.Color // 4xx
	ServerError *color.Color // 5xx
	Info        *color.Color // 1xx
	Default     *color.Color // 其他
}

// ColorManager 颜色管理器
type ColorManager struct {
	enabled bool
	colors  *StatusColors
}

// NewColorManager 创建新的颜色管理器
func NewColorManager(enabled bool) *ColorManager {
	// 在Windows上强制启用颜色
	if enabled {
		color.NoColor = false
	} else {
		color.NoColor = true
	}

	return &ColorManager{
		enabled: enabled,
		colors: &StatusColors{
			Success:     color.New(color.FgGreen, color.Bold),
			Redirect:    color.New(color.FgYellow, color.Bold),
			ClientError: color.New(color.FgRed, color.Bold),
			ServerError: color.New(color.FgMagenta, color.Bold),
			Info:        color.New(color.FgCyan, color.Bold),
			Default:     color.New(color.FgWhite),
		},
	}
}

// ColorizeStatus 为状态码添加颜色
func (cm *ColorManager) ColorizeStatus(statusCode int) string {
	if !cm.enabled {
		return strconv.Itoa(statusCode)
	}

	// 使用简单的文本颜色标识
	switch {
	case statusCode >= 200 && statusCode < 300:
		return fmt.Sprintf("✓%d✓", statusCode) // 成功 - 绿色标识
	case statusCode >= 300 && statusCode < 400:
		return fmt.Sprintf("→%d→", statusCode) // 重定向 - 黄色标识
	case statusCode >= 400 && statusCode < 500:
		return fmt.Sprintf("✗%d✗", statusCode) // 客户端错误 - 红色标识
	case statusCode >= 500 && statusCode < 600:
		return fmt.Sprintf("⚠%d⚠", statusCode) // 服务器错误 - 紫色标识
	case statusCode >= 100 && statusCode < 200:
		return fmt.Sprintf("ℹ%dℹ", statusCode) // 信息 - 蓝色标识
	default:
		return strconv.Itoa(statusCode)
	}
}

// ColorizeURL 为URL添加颜色
func (cm *ColorManager) ColorizeURL(url string) string {
	if !cm.enabled {
		return url
	}
	return url // 保持原样
}

// ColorizeSize 为响应大小添加颜色
func (cm *ColorManager) ColorizeSize(size int64) string {
	if !cm.enabled {
		return fmt.Sprintf("%d", size)
	}
	return fmt.Sprintf("%d", size) // 保持原样
}

// ColorizeTitle 为标题添加颜色
func (cm *ColorManager) ColorizeTitle(title string) string {
	if !cm.enabled {
		return title
	}
	return title // 保持原样
}

// ColorizeRedirect 为重定向添加颜色
func (cm *ColorManager) ColorizeRedirect(redirect string) string {
	if !cm.enabled {
		return redirect
	}
	return fmt.Sprintf("→%s→", redirect) // 重定向标识
}

// ColorizeError 为错误添加颜色
func (cm *ColorManager) ColorizeError(err string) string {
	if !cm.enabled {
		return err
	}
	return fmt.Sprintf("✗%s✗", err) // 错误标识
}

// ColorizeInfo 为信息添加颜色
func (cm *ColorManager) ColorizeInfo(info string) string {
	if !cm.enabled {
		return info
	}
	return fmt.Sprintf("ℹ%sℹ", info) // 信息标识
}

// ColorizeSuccess 为成功信息添加颜色
func (cm *ColorManager) ColorizeSuccess(success string) string {
	if !cm.enabled {
		return success
	}
	return fmt.Sprintf("✓%s✓", success) // 成功标识
}

// ColorizeWarning 为警告信息添加颜色
func (cm *ColorManager) ColorizeWarning(warning string) string {
	if !cm.enabled {
		return warning
	}
	return fmt.Sprintf("⚠%s⚠", warning) // 警告标识
}

// GetStatusColor 获取状态码对应的颜色
func (cm *ColorManager) GetStatusColor(statusCode int) *color.Color {
	if !cm.enabled {
		return color.New()
	}

	switch {
	case statusCode >= 200 && statusCode < 300:
		return cm.colors.Success
	case statusCode >= 300 && statusCode < 400:
		return cm.colors.Redirect
	case statusCode >= 400 && statusCode < 500:
		return cm.colors.ClientError
	case statusCode >= 500 && statusCode < 600:
		return cm.colors.ServerError
	case statusCode >= 100 && statusCode < 200:
		return cm.colors.Info
	default:
		return cm.colors.Default
	}
}

// Disable 禁用颜色
func (cm *ColorManager) Disable() {
	cm.enabled = false
}

// Enable 启用颜色
func (cm *ColorManager) Enable() {
	cm.enabled = true
}

// IsEnabled 检查颜色是否启用
func (cm *ColorManager) IsEnabled() bool {
	return cm.enabled
}

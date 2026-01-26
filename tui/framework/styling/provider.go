// Package styling 提供样式相关的抽象接口
// 作为组件和主题系统之间的隔离层，实现依赖反转
package styling

import (
	"sync"
)

// =============================================================================
// 依赖反转：组件样式提供者接口
// =============================================================================

// StyleProvider 样式提供者接口
// 组件依赖此抽象接口，而不是具体的 Theme 或 ThemeHolder
// 这实现了依赖反转原则（DIP）
type StyleProvider interface {
	// GetStyle 获取组件在指定状态下的样式配置
	GetStyle(componentID, state string) StyleConfig
}

// =============================================================================
// 样式配置（与 theme.StyleConfig 独立但兼容）
// =============================================================================

// Color 颜色表示（简化版，避免循环依赖）
type Color struct {
	Type  ColorType
	Value [3]int // RGB值
}

// ColorType 颜色类型
type ColorType int

const (
	ColorNone ColorType = iota
	ColorNamed
	ColorRGB
	ColorHex
	Color256
)

// StyleConfig 样式配置
// 这是一个轻量级的样式配置结构，避免与 theme 包循环依赖
type StyleConfig struct {
	Foreground *Color
	Background *Color
	Bold       bool
	Italic     bool
	Underline  bool
	Strikethrough bool
	Reverse    bool
	Blink      bool
}

// IsEmpty 检查样式配置是否为空
func (s StyleConfig) IsEmpty() bool {
	return s.Foreground == nil &&
		s.Background == nil &&
		!s.Bold &&
		!s.Italic &&
		!s.Underline &&
		!s.Strikethrough &&
		!s.Reverse &&
		!s.Blink
}

// =============================================================================
// 全局样式提供者注册表
// =============================================================================

var (
	globalRegistry struct {
		mu       sync.RWMutex
		provider StyleProvider
	}
)

// RegisterGlobalProvider 注册全局样式提供者
// 通常在应用初始化时调用，设置主题管理器
func RegisterGlobalProvider(provider StyleProvider) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()
	globalRegistry.provider = provider
}

// GetGlobalProvider 获取全局样式提供者
// 如果没有注册，返回默认提供者
func GetGlobalProvider() StyleProvider {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	if globalRegistry.provider != nil {
		return globalRegistry.provider
	}
	return DefaultProvider{}
}

// =============================================================================
// 默认样式提供者
// =============================================================================

// DefaultProvider 默认样式提供者
// 当没有设置主题时使用，提供基本的默认样式
type DefaultProvider struct{}

// GetStyle 实现 StyleProvider 接口
func (p DefaultProvider) GetStyle(componentID, state string) StyleConfig {
	// 为不同组件和状态提供合理的默认样式
	switch componentID {
	case "input":
		return p.getInputStyle(state)
	case "button":
		return p.getButtonStyle(state)
	default:
		return StyleConfig{}
	}
}

// getInputStyle 获取 input 组件的默认样式
func (p DefaultProvider) getInputStyle(state string) StyleConfig {
	switch state {
	case "focus":
		return StyleConfig{
			Foreground: &Color{Type: ColorNamed, Value: [3]int{0, 255, 255}}, // Cyan
		}
	case "placeholder":
		return StyleConfig{
			Foreground: &Color{Type: ColorNamed, Value: [3]int{128, 128, 128}}, // Gray
		}
	default:
		return StyleConfig{}
	}
}

// getButtonStyle 获取 button 组件的默认样式
func (p DefaultProvider) getButtonStyle(state string) StyleConfig {
	switch state {
	case "focus":
		return StyleConfig{
		Foreground: &Color{Type: ColorNamed, Value: [3]int{0, 255, 255}}, // Cyan
		Bold:       true,
	}
	case "disabled":
		return StyleConfig{
		Foreground: &Color{Type: ColorNamed, Value: [3]int{128, 128, 128}}, // Gray
	}
	default:
		return StyleConfig{}
	}
}

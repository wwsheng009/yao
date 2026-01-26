package styling

import (
	"sync/atomic"

	"github.com/yaoapp/yao/tui/framework/style"
)

// =============================================================================
// 全局样式获取函数 - 让组件完全不依赖 styling 包
// =============================================================================

// StyleGetter 样式获取函数类型
// 组件可以持有此函数类型，而不是 StyleProvider 接口
// 这样组件只需要导入 style 包，而不需要导入 styling 包
type StyleGetter func(componentID, state string) style.Style

// globalGetter 全局样式获取函数（原子存储）
var globalGetter atomic.Value // 存储 StyleGetter

// init 初始化默认获取函数
func init() {
	globalGetter.Store(defaultStyleGetter)
}

// GetStyleGetter 获取全局样式获取函数
// 这是组件获取样式的入口点
func GetStyleGetter() StyleGetter {
	getter := globalGetter.Load()
	if getter == nil {
		return defaultStyleGetter
	}
	return getter.(StyleGetter)
}

// SetStyleGetter 设置全局样式获取函数
// 通常由主题系统调用
func SetStyleGetter(getter StyleGetter) {
	globalGetter.Store(getter)
}

// defaultStyleGetter 默认样式获取函数
func defaultStyleGetter(componentID, state string) style.Style {
	switch componentID {
	case "input":
		return defaultInputStyle(state)
	case "button":
		return defaultButtonStyle(state)
	default:
		return style.Style{}
	}
}

// defaultInputStyle 默认 input 样式
func defaultInputStyle(state string) style.Style {
	switch state {
	case "focus":
		return style.Style{}.Foreground(style.Cyan)
	case "placeholder":
		return style.Style{}.Foreground(style.BrightBlack)
	default:
		return style.Style{}
	}
}

// defaultButtonStyle 默认 button 样式
func defaultButtonStyle(state string) style.Style {
	switch state {
	case "focus":
		return style.Style{}.Foreground(style.Cyan).Bold(true)
	case "disabled":
		return style.Style{}.Foreground(style.BrightBlack)
	default:
		return style.Style{}
	}
}

// =============================================================================
// StyleProvider 接口（向后兼容）
// =============================================================================

// StyleProvider 样式提供者接口
type StyleProvider interface {
	GetStyle(componentID, state string) style.Style
}

// DefaultProvider 默认样式提供者
type DefaultProvider struct{}

func (p DefaultProvider) GetStyle(componentID, state string) style.Style {
	return defaultStyleGetter(componentID, state)
}

// NewDefaultProvider 创建默认样式提供者
func NewDefaultProvider() *DefaultProvider {
	return &DefaultProvider{}
}

// =============================================================================
// 转换函数
// =============================================================================

// ProviderToGetter 将 StyleProvider 转换为 StyleGetter
func ProviderToGetter(provider StyleProvider) StyleGetter {
	if provider == nil {
		return defaultStyleGetter
	}
	return func(componentID, state string) style.Style {
		return provider.GetStyle(componentID, state)
	}
}

// GetterToProvider 将 StyleGetter 转换为 StyleProvider
func GetterToProvider(getter StyleGetter) StyleProvider {
	return styleGetterProvider{getter: getter}
}

// styleGetterProvider StyleGetter 的适配器，实现 StyleProvider 接口
type styleGetterProvider struct {
	getter StyleGetter
}

func (p styleGetterProvider) GetStyle(componentID, state string) style.Style {
	return p.getter(componentID, state)
}

// =============================================================================
// 全局提供者管理（向后兼容）
// =============================================================================

// GetProvider 获取当前样式提供者
func GetProvider() StyleProvider {
	getter := globalGetter.Load()
	if getter == nil {
		return DefaultProvider{}
	}
	return GetterToProvider(getter.(StyleGetter))
}

// SetProvider 设置样式提供者
func SetProvider(provider StyleProvider) {
	SetStyleGetter(ProviderToGetter(provider))
}

// =============================================================================
// 主题提供者接口
// =============================================================================

// ThemeProvider 主题样式提供者接口
// 主题系统需要实现此接口来支持全局切换
type ThemeProvider interface {
	StyleProvider

	// Name 返回主题名称
	Name() string

	// SetTheme 切换主题
	SetTheme(name string) error

	// CurrentTheme 返回当前主题
	CurrentTheme() string
}

// =============================================================================
// 便捷主题切换函数
// =============================================================================

// SetTheme 设置主题（如果全局提供者是 ThemeProvider）
// 这是应用切换主题的主要入口
func SetTheme(name string) error {
	if provider, ok := GetProvider().(ThemeProvider); ok {
		return provider.SetTheme(name)
	}
	return ErrNotAThemeProvider
}

// CurrentTheme 返回当前主题名称
func CurrentTheme() string {
	if provider, ok := GetProvider().(ThemeProvider); ok {
		return provider.CurrentTheme()
	}
	return "default"
}

// =============================================================================
// 错误
// =============================================================================

var (
	ErrNotAThemeProvider = &StyleError{Msg: "global provider is not a ThemeProvider"}
	ErrThemeNotFound     = &StyleError{Msg: "theme not found"}
)

// StyleError 样式错误
type StyleError struct {
	Msg string
}

func (e *StyleError) Error() string {
	return e.Msg
}

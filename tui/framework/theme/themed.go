package theme

import (
	"strings"
)

// Themed 有主题的组件接口
// 组件可以实现此接口以获得主题支持
type Themed interface {
	// GetThemeStyle 返回指定样式键的样式配置
	GetThemeStyle(styleKey string) StyleConfig

	// SetThemeManager 设置主题管理器
	SetThemeManager(manager *Manager)

	// GetThemeManager 获取主题管理器
	GetThemeManager() *Manager
}

// ThemeHolder 主题持有者（嵌入组件）
// 可以嵌入到任何组件中以提供主题支持
type ThemeHolder struct {
	themeMgr      *Manager
	styleOverrides map[string]StyleConfig
	componentID   string
}

// NewThemeHolder 创建主题持有者
func NewThemeHolder(manager *Manager) *ThemeHolder {
	return &ThemeHolder{
		themeMgr:        manager,
		styleOverrides:  make(map[string]StyleConfig),
	}
}

// NewThemeHolderWithID 创建带有组件 ID 的主题持有者
func NewThemeHolderWithID(manager *Manager, componentID string) *ThemeHolder {
	return &ThemeHolder{
		themeMgr:        manager,
		styleOverrides:  make(map[string]StyleConfig),
		componentID:     componentID,
	}
}

// SetThemeManager 设置主题管理器
func (h *ThemeHolder) SetThemeManager(manager *Manager) {
	h.themeMgr = manager
}

// GetThemeManager 获取主题管理器
func (h *ThemeHolder) GetThemeManager() *Manager {
	return h.themeMgr
}

// SetComponentID 设置组件 ID
func (h *ThemeHolder) SetComponentID(id string) {
	h.componentID = id
}

// GetComponentID 获取组件 ID
func (h *ThemeHolder) GetComponentID() string {
	return h.componentID
}

// GetStyle 获取样式
// 优先级：组件级覆盖 > 组件特定样式 > 全局样式 > 默认样式
func (h *ThemeHolder) GetStyle(styleKey string) StyleConfig {
	// 1. 检查组件级覆盖
	if override, ok := h.styleOverrides[styleKey]; ok {
		return override
	}

	// 2. 从主题管理器获取
	if h.themeMgr != nil {
		return h.themeMgr.GetStyle(h.componentID, styleKey)
	}

	return StyleConfig{}
}

// GetStyleWithState 获取带状态的样式
func (h *ThemeHolder) GetStyleWithState(styleKey, state string) StyleConfig {
	// 1. 检查组件级覆盖（状态样式）
	key := styleKey + "." + state
	if override, ok := h.styleOverrides[key]; ok {
		return override
	}

	// 2. 从主题管理器获取
	if h.themeMgr != nil {
		// 先尝试组件特定状态样式
		compStyle := h.themeMgr.GetComponentStyle(h.componentID, state)
		if compStyle.Foreground != nil || compStyle.Background != nil {
			return compStyle
		}
	}

	// 3. 回退到基础样式
	return h.GetStyle(styleKey)
}

// OverrideStyle 覆盖样式
func (h *ThemeHolder) OverrideStyle(styleKey string, config StyleConfig) {
	h.styleOverrides[styleKey] = config
}

// OverrideStyleWithState 覆盖带状态的样式
func (h *ThemeHolder) OverrideStyleWithState(styleKey, state string, config StyleConfig) {
	key := styleKey + "." + state
	h.styleOverrides[key] = config
}

// ClearOverrides 清除所有覆盖
func (h *ThemeHolder) ClearOverrides() {
	h.styleOverrides = make(map[string]StyleConfig)
}

// ClearStyleOverride 清除指定样式的覆盖
func (h *ThemeHolder) ClearStyleOverride(styleKey string) {
	delete(h.styleOverrides, styleKey)
}

// GetColor 获取颜色
func (h *ThemeHolder) GetColor(colorKey string) Color {
	if h.themeMgr != nil {
		return h.themeMgr.GetColor(colorKey)
	}
	return NoColor
}

// GetPrimary 获取主色
func (h *ThemeHolder) GetPrimary() Color {
	return h.GetColor("primary")
}

// GetSecondary 获取次要色
func (h *ThemeHolder) GetSecondary() Color {
	return h.GetColor("secondary")
}

// GetSuccess 获取成功色
func (h *ThemeHolder) GetSuccess() Color {
	return h.GetColor("success")
}

// GetWarning 获取警告色
func (h *ThemeHolder) GetWarning() Color {
	return h.GetColor("warning")
}

// GetError 获取错误色
func (h *ThemeHolder) GetError() Color {
	return h.GetColor("error")
}

// GetInfo 获取信息色
func (h *ThemeHolder) GetInfo() Color {
	return h.GetColor("info")
}

// GetBackground 获取背景色
func (h *ThemeHolder) GetBackground() Color {
	return h.GetColor("background")
}

// GetForeground 获取前景色
func (h *ThemeHolder) GetForeground() Color {
	return h.GetColor("foreground")
}

// GetBorder 获取边框色
func (h *ThemeHolder) GetBorder() Color {
	return h.GetColor("border")
}

// GetFocus 获取焦点色
func (h *ThemeHolder) GetFocus() Color {
	return h.GetColor("focus")
}

// GetDisabled 获取禁用色
func (h *ThemeHolder) GetDisabled() Color {
	return h.GetColor("disabled")
}

// GetHover 获取悬停色
func (h *ThemeHolder) GetHover() Color {
	return h.GetColor("hover")
}

// GetActive 获取激活色
func (h *ThemeHolder) GetActive() Color {
	return h.GetColor("active")
}

// GetMuted 获取静默色
func (h *ThemeHolder) GetMuted() Color {
	return h.GetColor("muted")
}

// GetSpacing 获取间距值
func (h *ThemeHolder) GetSpacing(size string) int {
	if h.themeMgr != nil && h.themeMgr.current != nil {
		return h.themeMgr.current.GetSpacing(size)
	}
	return 0
}

// GetTheme 获取当前主题
func (h *ThemeHolder) GetTheme() *Theme {
	if h.themeMgr != nil {
		return h.themeMgr.Current()
	}
	return nil
}

// Subscribe 订阅主题变化
func (h *ThemeHolder) Subscribe(listener ThemeChangeListener) func() {
	if h.themeMgr != nil {
		return h.themeMgr.Subscribe(listener)
	}
	return func() {}
}

// =============================================================================
// StyleConfig 辅助函数
// =============================================================================

// NewStyleConfig 创建样式配置
func NewStyleConfig() StyleConfig {
	return StyleConfig{}
}

// WithForeground 设置前景色
func (s StyleConfig) WithForeground(color Color) StyleConfig {
	s.Foreground = &color
	return s
}

// WithBackground 设置背景色
func (s StyleConfig) WithBackground(color Color) StyleConfig {
	s.Background = &color
	return s
}

// WithBold 设置粗体
func (s StyleConfig) WithBold() StyleConfig {
	s.Bold = true
	return s
}

// WithItalic 设置斜体
func (s StyleConfig) WithItalic() StyleConfig {
	s.Italic = true
	return s
}

// WithUnderline 设置下划线
func (s StyleConfig) WithUnderline() StyleConfig {
	s.Underline = true
	return s
}

// WithStrikethrough 设置删除线
func (s StyleConfig) WithStrikethrough() StyleConfig {
	s.Strikethrough = true
	return s
}

// WithReverse 设置反白
func (s StyleConfig) WithReverse() StyleConfig {
	s.Reverse = true
	return s
}

// WithBlink 设置闪烁
func (s StyleConfig) WithBlink() StyleConfig {
	s.Blink = true
	return s
}

// WithPadding 设置内边距
func (s StyleConfig) WithPadding(top, right, bottom, left int) StyleConfig {
	s.Padding = &[4]int{top, right, bottom, left}
	return s
}

// WithPaddingUniform 设置统一内边距
func (s StyleConfig) WithPaddingUniform(padding int) StyleConfig {
	s.Padding = &[4]int{padding, padding, padding, padding}
	return s
}

// WithMargin 设置外边距
func (s StyleConfig) WithMargin(top, right, bottom, left int) StyleConfig {
	s.Margin = &[4]int{top, right, bottom, left}
	return s
}

// WithMarginUniform 设置统一外边距
func (s StyleConfig) WithMarginUniform(margin int) StyleConfig {
	s.Margin = &[4]int{margin, margin, margin, margin}
	return s
}

// WithWidth 设置宽度
func (s StyleConfig) WithWidth(width int) StyleConfig {
	s.Width = &width
	return s
}

// WithHeight 设置高度
func (s StyleConfig) WithHeight(height int) StyleConfig {
	s.Height = &height
	return s
}

// WithBorder 设置边框
func (s StyleConfig) WithBorder(border BorderStyle) StyleConfig {
	s.Border = &border
	return s
}

// Merge 合并另一个样式配置
func (s StyleConfig) Merge(other StyleConfig) StyleConfig {
	if other.Foreground != nil {
		s.Foreground = other.Foreground
	}
	if other.Background != nil {
		s.Background = other.Background
	}
	if other.Bold {
		s.Bold = true
	}
	if other.Italic {
		s.Italic = true
	}
	if other.Underline {
		s.Underline = true
	}
	if other.Strikethrough {
		s.Strikethrough = true
	}
	if other.Reverse {
		s.Reverse = true
	}
	if other.Blink {
		s.Blink = true
	}
	if other.Padding != nil {
		s.Padding = other.Padding
	}
	if other.Margin != nil {
		s.Margin = other.Margin
	}
	if other.Width != nil {
		s.Width = other.Width
	}
	if other.Height != nil {
		s.Height = other.Height
	}
	if other.Border != nil {
		s.Border = other.Border
	}
	return s
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
		!s.Blink &&
		s.Padding == nil &&
		s.Margin == nil &&
		s.Width == nil &&
		s.Height == nil &&
		s.Border == nil
}

// ParseStylePath 解析样式路径
// 支持格式: "component.state", "component", "state"
func ParseStylePath(path string) (component, state string) {
	parts := strings.Split(path, ".")
	switch len(parts) {
	case 1:
		return "", parts[0]
	case 2:
		return parts[0], parts[1]
	default:
		return strings.Join(parts[:len(parts)-1], "."), parts[len(parts)-1]
	}
}

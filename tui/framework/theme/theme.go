package theme

import (
	"strings"
	"sync"
)

// Theme 主题
type Theme struct {
	// 主题名称
	Name string

	// 主题版本
	Version string

	// 父主题（用于继承）
	Parent *Theme

	// 颜色调色板
	Colors ColorPalette

	// 间距配置
	Spacing SpacingSet

	// 样式映射（全局样式）
	Styles map[string]StyleConfig

	// 组件特定配置
	Components map[string]ComponentStyle

	// 元数据
	Metadata map[string]interface{}

	mu sync.RWMutex
}

// SpacingSet 间距集
type SpacingSet struct {
	XS int
	SM int
	MD int
	LG int
	XL int
}

// StyleConfig 样式配置
type StyleConfig struct {
	Foreground *Color
	Background *Color
	Bold       bool
	Italic     bool
	Underline  bool
	Strikethrough bool
	Reverse    bool
	Blink      bool
	Padding    *[4]int // top, right, bottom, left
	Margin     *[4]int
	Width      *int
	Height     *int
	Border     *BorderStyle
}

// ComponentStyle 组件样式
type ComponentStyle struct {
	Base   StyleConfig
	States map[string]StyleConfig // hover, focus, disabled, etc.
}

// NewTheme 创建新主题
func NewTheme(name string) *Theme {
	return &Theme{
		Name:       name,
		Version:    "1.0.0",
		Colors:     NewColorPalette(),
		Spacing:    DefaultSpacingSet(),
		Styles:     make(map[string]StyleConfig),
		Components: make(map[string]ComponentStyle),
		Metadata:   make(map[string]interface{}),
	}
}

// NewThemeWithPalette 创建带有指定调色板的主题
func NewThemeWithPalette(name string, colors ColorPalette) *Theme {
	return &Theme{
		Name:       name,
		Version:    "1.0.0",
		Colors:     colors,
		Spacing:    DefaultSpacingSet(),
		Styles:     make(map[string]StyleConfig),
		Components: make(map[string]ComponentStyle),
		Metadata:   make(map[string]interface{}),
	}
}

// DefaultSpacingSet 返回默认间距集
func DefaultSpacingSet() SpacingSet {
	return SpacingSet{
		XS: 1,
		SM: 2,
		MD: 4,
		LG: 6,
		XL: 8,
	}
}

// GetColor 获取颜色（支持继承）
func (t *Theme) GetColor(key string) Color {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.resolveColor(key)
}

// resolveColor 递归解析颜色
func (t *Theme) resolveColor(key string) Color {
	// 支持嵌套访问，如 "primary.light"
	parts := strings.Split(key, ".")

	var color Color
	switch len(parts) {
	case 1:
		color = t.getColorByName(parts[0])
	default:
		// 简单实现：只支持第一层
		color = t.getColorByName(parts[0])
	}

	if !color.IsNone() || t.Parent == nil {
		return color
	}

	// 从父主题获取
	return t.Parent.resolveColor(key)
}

func (t *Theme) getColorByName(name string) Color {
	switch strings.ToLower(name) {
	case "primary":
		return t.Colors.Primary
	case "secondary":
		return t.Colors.Secondary
	case "accent":
		return t.Colors.Accent
	case "success":
		return t.Colors.Success
	case "warning":
		return t.Colors.Warning
	case "error":
		return t.Colors.Error
	case "info":
		return t.Colors.Info
	case "background", "bg":
		return t.Colors.Background
	case "foreground", "fg":
		return t.Colors.Foreground
	case "muted":
		return t.Colors.Muted
	case "border":
		return t.Colors.Border
	case "focus":
		return t.Colors.Focus
	case "disabled":
		return t.Colors.Disabled
	case "hover":
		return t.Colors.Hover
	case "active":
		return t.Colors.Active
	default:
		return NoColor
	}
}

// GetStyle 获取样式（支持继承）
func (t *Theme) GetStyle(styleKey string) StyleConfig {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if style, ok := t.Styles[styleKey]; ok {
		return style
	}

	// 从父主题获取
	if t.Parent != nil {
		return t.Parent.GetStyle(styleKey)
	}

	return StyleConfig{}
}

// GetComponentStyle 获取组件样式
func (t *Theme) GetComponentStyle(componentID, state string) StyleConfig {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// 1. 查找组件特定样式
	if compStyle, ok := t.Components[componentID]; ok {
		if state != "" {
			if stateStyle, ok := compStyle.States[state]; ok {
				return stateStyle
			}
		}
		return compStyle.Base
	}

	// 2. 从父主题获取
	if t.Parent != nil {
		return t.Parent.GetComponentStyle(componentID, state)
	}

	return StyleConfig{}
}

// SetStyle 设置全局样式
func (t *Theme) SetStyle(key string, style StyleConfig) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.Styles == nil {
		t.Styles = make(map[string]StyleConfig)
	}
	t.Styles[key] = style
}

// SetComponentStyle 设置组件样式
func (t *Theme) SetComponentStyle(componentID string, base StyleConfig, states map[string]StyleConfig) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.Components == nil {
		t.Components = make(map[string]ComponentStyle)
	}

	t.Components[componentID] = ComponentStyle{
		Base:   base,
		States: states,
	}
}

// SetParent 设置父主题
func (t *Theme) SetParent(parent *Theme) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.Parent = parent
}

// Extend 创建继承自当前主题的新主题
func (t *Theme) Extend(name string) *Theme {
	child := NewTheme(name)
	child.Parent = t
	return child
}

// WithColor 返回修改指定颜色后的新主题
func (t *Theme) WithColor(colorName string, color Color) *Theme {
	newTheme := *t
	newTheme.Colors = t.Colors

	// 修改指定颜色
	switch strings.ToLower(colorName) {
	case "primary":
		newTheme.Colors.Primary = color
	case "secondary":
		newTheme.Colors.Secondary = color
	case "accent":
		newTheme.Colors.Accent = color
	case "success":
		newTheme.Colors.Success = color
	case "warning":
		newTheme.Colors.Warning = color
	case "error":
		newTheme.Colors.Error = color
	case "info":
		newTheme.Colors.Info = color
	case "background", "bg":
		newTheme.Colors.Background = color
	case "foreground", "fg":
		newTheme.Colors.Foreground = color
	case "muted":
		newTheme.Colors.Muted = color
	case "border":
		newTheme.Colors.Border = color
	case "focus":
		newTheme.Colors.Focus = color
	case "disabled":
		newTheme.Colors.Disabled = color
	case "hover":
		newTheme.Colors.Hover = color
	case "active":
		newTheme.Colors.Active = color
	}

	return &newTheme
}

// Clone 克隆主题
// 注意：不是线程安全的，调用者应确保没有并发写入
func (t *Theme) Clone() *Theme {
	clone := NewTheme(t.Name + "_clone")
	clone.Version = t.Version
	clone.Parent = t.Parent // 父主题共享引用，不深拷贝

	// 复制值类型
	clone.Colors = t.Colors
	clone.Spacing = t.Spacing

	// 深拷贝 Styles
	t.mu.RLock()
	if t.Styles != nil {
		clone.Styles = make(map[string]StyleConfig, len(t.Styles))
		for k, v := range t.Styles {
			clone.Styles[k] = v
		}
	}

	// 深拷贝 Components
	if t.Components != nil {
		clone.Components = make(map[string]ComponentStyle, len(t.Components))
		for k, v := range t.Components {
			clone.Components[k] = v
		}
	}
	t.mu.RUnlock()

	// 深拷贝 Metadata
	if t.Metadata != nil {
		clone.Metadata = make(map[string]interface{}, len(t.Metadata))
		for k, v := range t.Metadata {
			clone.Metadata[k] = v
		}
	}

	return clone
}

// Merge 合并另一个主题的样式
func (t *Theme) Merge(other *Theme) *Theme {
	t.mu.Lock()
	defer t.mu.Unlock()

	// 创建新主题而不调用 Clone()，避免死锁
	result := NewTheme(t.Name + "_merged")
	result.Version = t.Version
	result.Parent = t.Parent
	result.Colors = t.Colors
	result.Spacing = t.Spacing

	// 深拷贝 Styles
	if t.Styles != nil {
		result.Styles = make(map[string]StyleConfig, len(t.Styles))
		for k, v := range t.Styles {
			result.Styles[k] = v
		}
	}

	// 深拷贝 Components
	if t.Components != nil {
		result.Components = make(map[string]ComponentStyle, len(t.Components))
		for k, v := range t.Components {
			result.Components[k] = v
		}
	}

	// 深拷贝 Metadata
	if t.Metadata != nil {
		result.Metadata = make(map[string]interface{}, len(t.Metadata))
		for k, v := range t.Metadata {
			result.Metadata[k] = v
		}
	}

	// 合并颜色（other 优先）
	if other != nil {
		// 如果颜色不是 NoColor，则覆盖
		if !other.Colors.Primary.IsNone() {
			result.Colors.Primary = other.Colors.Primary
		}
		if !other.Colors.Secondary.IsNone() {
			result.Colors.Secondary = other.Colors.Secondary
		}
		if !other.Colors.Accent.IsNone() {
			result.Colors.Accent = other.Colors.Accent
		}
		if !other.Colors.Success.IsNone() {
			result.Colors.Success = other.Colors.Success
		}
		if !other.Colors.Warning.IsNone() {
			result.Colors.Warning = other.Colors.Warning
		}
		if !other.Colors.Error.IsNone() {
			result.Colors.Error = other.Colors.Error
		}
		if !other.Colors.Info.IsNone() {
			result.Colors.Info = other.Colors.Info
		}
		if !other.Colors.Background.IsNone() {
			result.Colors.Background = other.Colors.Background
		}
		if !other.Colors.Foreground.IsNone() {
			result.Colors.Foreground = other.Colors.Foreground
		}
		if !other.Colors.Muted.IsNone() {
			result.Colors.Muted = other.Colors.Muted
		}
		if !other.Colors.Border.IsNone() {
			result.Colors.Border = other.Colors.Border
		}
		if !other.Colors.Focus.IsNone() {
			result.Colors.Focus = other.Colors.Focus
		}
		if !other.Colors.Disabled.IsNone() {
			result.Colors.Disabled = other.Colors.Disabled
		}
		if !other.Colors.Hover.IsNone() {
			result.Colors.Hover = other.Colors.Hover
		}
		if !other.Colors.Active.IsNone() {
			result.Colors.Active = other.Colors.Active
		}

		// 合并样式
		for k, v := range other.Styles {
			result.Styles[k] = v
		}

		// 合并组件样式
		for k, v := range other.Components {
			result.Components[k] = v
		}
	}

	return result
}

// GetSpacing 获取间距值
func (t *Theme) GetSpacing(size string) int {
	t.mu.RLock()
	defer t.mu.RUnlock()

	switch strings.ToLower(size) {
	case "xs":
		return t.Spacing.XS
	case "sm":
		return t.Spacing.SM
	case "md":
		return t.Spacing.MD
	case "lg":
		return t.Spacing.LG
	case "xl":
		return t.Spacing.XL
	default:
		return 0
	}
}

// SetMetadata 设置元数据
func (t *Theme) SetMetadata(key string, value interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.Metadata == nil {
		t.Metadata = make(map[string]interface{})
	}
	t.Metadata[key] = value
}

// GetMetadata 获取元数据
func (t *Theme) GetMetadata(key string) (interface{}, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.Metadata == nil {
		return nil, false
	}
	val, ok := t.Metadata[key]
	return val, ok
}

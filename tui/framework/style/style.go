package style

import (
	"fmt"
	"strings"
)

// =============================================================================
// 全局样式获取 - 连接到主题系统
// =============================================================================

// getGlobalStyleGetter 获取全局样式获取函数
// 这是一个桥接函数，让 style 包可以访问 styling 包的全局样式
// 使用包变量函数避免循环导入
var getGlobalStyleGetter func() func(string, string) Style

// RegisterStyleGetter 注册全局样式获取函数
// 由主题系统在初始化时调用
func RegisterStyleGetter(getter func() func(string, string) Style) {
	getGlobalStyleGetter = getter
}

// GetStyle 获取组件样式
// 这是从组件获取样式的入口点
// 如果主题系统未初始化，返回空样式
func GetStyle(componentID, state string) Style {
	if getGlobalStyleGetter == nil {
		return Style{}
	}
	getter := getGlobalStyleGetter()
	if getter == nil {
		return Style{}
	}
	return getter(componentID, state)
}

// Color 颜色表示（简化版，与 theme.Color 兼容）
// 建议新代码使用 theme.Color
type Color string

const (
	Black   Color = "black"
	Red     Color = "red"
	Green   Color = "green"
	Yellow  Color = "yellow"
	Blue    Color = "blue"
	Magenta Color = "magenta"
	Cyan    Color = "cyan"
	White   Color = "white"

	BrightBlack   Color = "bright-black"
	BrightRed     Color = "bright-red"
	BrightGreen   Color = "bright-green"
	BrightYellow  Color = "bright-yellow"
	BrightBlue    Color = "bright-blue"
	BrightMagenta Color = "bright-magenta"
	BrightCyan    Color = "bright-cyan"
	BrightWhite   Color = "bright-white"

	NoColor Color = ""
)

// Style 样式定义
// 注意：此 Style 结构体与 theme.StyleConfig 不同
// 这是 framework 层的基础样式，theme.StyleConfig 是更高级的主题样式
type Style struct {
	FG            Color
	BG            Color
	isBold        bool
	isItalic      bool
	isUnderline   bool
	isStrikethrough bool
	isReverse     bool
	isBlink       bool
}

// NewStyle 创建默认样式
func NewStyle() Style {
	return Style{}
}

// Foreground 设置前景色
func (s Style) Foreground(c Color) Style {
	s.FG = c
	return s
}

// Background 设置背景色
func (s Style) Background(c Color) Style {
	s.BG = c
	return s
}

// Bold 设置粗体
func (s Style) Bold(on bool) Style {
	s.isBold = on
	return s
}

// IsBold 获取粗体状态
func (s Style) IsBold() bool {
	return s.isBold
}

// Italic 设置斜体
func (s Style) Italic(on bool) Style {
	s.isItalic = on
	return s
}

// IsItalic 获取斜体状态
func (s Style) IsItalic() bool {
	return s.isItalic
}

// Underline 设置下划线
func (s Style) Underline(on bool) Style {
	s.isUnderline = on
	return s
}

// IsUnderline 获取下划线状态
func (s Style) IsUnderline() bool {
	return s.isUnderline
}

// Strikethrough 设置删除线
func (s Style) Strikethrough(on bool) Style {
	s.isStrikethrough = on
	return s
}

// IsStrikethrough 获取删除线状态
func (s Style) IsStrikethrough() bool {
	return s.isStrikethrough
}

// Reverse 设置反白
func (s Style) Reverse(on bool) Style {
	s.isReverse = on
	return s
}

// IsReverse 获取反白状态
func (s Style) IsReverse() bool {
	return s.isReverse
}

// Blink 设置闪烁
func (s Style) Blink(on bool) Style {
	s.isBlink = on
	return s
}

// IsBlink 获取闪烁状态
func (s Style) IsBlink() bool {
	return s.isBlink
}

// Merge 合并样式
func (s Style) Merge(other Style) Style {
	result := s
	if other.FG != NoColor {
		result.FG = other.FG
	}
	if other.BG != NoColor {
		result.BG = other.BG
	}
	if other.isBold {
		result.isBold = true
	}
	if other.isItalic {
		result.isItalic = true
	}
	if other.isUnderline {
		result.isUnderline = true
	}
	if other.isStrikethrough {
		result.isStrikethrough = true
	}
	if other.isReverse {
		result.isReverse = true
	}
	if other.isBlink {
		result.isBlink = true
	}
	return result
}

// ToANSI 转换为 ANSI 转义码
func (s Style) ToANSI() string {
	var codes []string

	if fg, ok := colorCodes[string(s.FG)]; ok {
		codes = append(codes, fmt.Sprintf("%d", fg+30))
	}
	if bg, ok := colorCodes[string(s.BG)]; ok {
		codes = append(codes, fmt.Sprintf("%d", bg+40))
	}
	if s.isBold {
		codes = append(codes, "1")
	}
	if s.isItalic {
		codes = append(codes, "3")
	}
	if s.isUnderline {
		codes = append(codes, "4")
	}
	if s.isBlink {
		codes = append(codes, "5")
	}
	if s.isReverse {
		codes = append(codes, "7")
	}

	if len(codes) == 0 {
		return ""
	}
	return "\x1b[" + strings.Join(codes, ";") + "m"
}

// Apply 应用样式到文本
func (s Style) Apply(text string) string {
	if s == (Style{}) {
		return text
	}
	ansi := s.ToANSI()
	if ansi == "" {
		return text
	}
	return ansi + text + "\x1b[0m"
}

// colorCodes 颜色代码映射
var colorCodes = map[string]int{
	"black":         0,
	"red":           1,
	"green":         2,
	"yellow":        3,
	"blue":          4,
	"magenta":       5,
	"cyan":          6,
	"white":         7,
	"bright-black":   8,
	"bright-red":     9,
	"bright-green":  10,
	"bright-yellow": 11,
	"bright-blue":   12,
	"bright-magenta": 13,
	"bright-cyan":   14,
	"bright-white":  15,
}

// Builder 样式构建器
type Builder struct {
	style Style
}

// NewBuilder 创建样式构建器
func NewBuilder() *Builder {
	return &Builder{style: NewStyle()}
}

// Foreground 设置前景色
func (b *Builder) Foreground(color string) *Builder {
	b.style.FG = ParseColor(color)
	return b
}

// Background 设置背景色
func (b *Builder) Background(color string) *Builder {
	b.style.BG = ParseColor(color)
	return b
}

// Bold 设置粗体
func (b *Builder) Bold() *Builder {
	b.style.isBold = true
	return b
}

// Italic 设置斜体
func (b *Builder) Italic() *Builder {
	b.style.isItalic = true
	return b
}

// Underline 设置下划线
func (b *Builder) Underline() *Builder {
	b.style.isUnderline = true
	return b
}

// Reverse 设置反白
func (b *Builder) Reverse() *Builder {
	b.style.isReverse = true
	return b
}

// Build 构建样式
func (b *Builder) Build() Style {
	return b.style
}

// ParseColor 解析颜色字符串
func ParseColor(s string) Color {
	return Color(s)
}

// ParseRGB 解析 RGB 颜色
func ParseRGB(r, g, b int) Color {
	return Color(fmt.Sprintf("rgb(%d,%d,%d)", r, g, b))
}

// ParseHex 解析十六进制颜色
func ParseHex(hex string) Color {
	return Color(hex)
}

// =============================================================================
// 便捷函数
// =============================================================================

// WithForeground 返回带前景色的样式
func WithForeground(c Color) Style {
	return NewStyle().Foreground(c)
}

// WithBackground 返回带背景色的样式
func WithBackground(c Color) Style {
	return NewStyle().Background(c)
}

// WithBold 返回带粗体的样式
func WithBold() Style {
	return NewStyle().Bold(true)
}

// WithItalic 返回带斜体的样式
func WithItalic() Style {
	return NewStyle().Italic(true)
}

// WithUnderline 返回带下划线的样式
func WithUnderline() Style {
	return NewStyle().Underline(true)
}

// WithReverse 返回带反白的样式
func WithReverse() Style {
	return NewStyle().Reverse(true)
}

// Combine 组合多个样式
func Combine(styles ...Style) Style {
	result := NewStyle()
	for _, s := range styles {
		result = result.Merge(s)
	}
	return result
}

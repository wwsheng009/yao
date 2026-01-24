package style

import (
	"fmt"
	"strconv"
	"strings"
)

// Color 颜色表示
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
	"bright-magenta":13,
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
	return fmt.Sprintf("rgb(%d,%d,%d)", r, g, b)
}

// ParseHex 解析十六进制颜色
func ParseHex(hex string) Color {
	return Color(hex)
}

// Theme 主题
type Theme struct {
	Name string

	Primary   Color
	Secondary Color

	Success   Color
	Warning   Color
	Error     Color
	Info      Color

	Foreground Color
	Background Color
}

// DefaultTheme 默认主题
var DefaultTheme = &Theme{
	Name:       "default",
	Primary:    Blue,
	Secondary:  Cyan,
	Success:    Green,
	Warning:    Yellow,
	Error:      Red,
	Info:       Blue,
	Foreground: White,
	Background: Black,
}

// LightTheme 亮色主题
var LightTheme = &Theme{
	Name:       "light",
	Primary:    Blue,
	Secondary:  Cyan,
	Success:    Green,
	Warning:    Yellow,
	Error:      Red,
	Info:       Blue,
	Foreground: Black,
	Background: White,
}

// DraculaTheme Dracula 主题
var DraculaTheme = &Theme{
	Name:       "dracula",
	Primary:    ParseColor("#bd93f9"),
	Secondary:  ParseColor("#ff79c6"),
	Success:    ParseColor("#50fa7b"),
	Warning:    ParseColor("#f1fa8c"),
	Error:      ParseColor("#ff5555"),
	Info:       ParseColor("#8be9fd"),
	Foreground: ParseColor("#f8f8f2"),
	Background: ParseColor("#282a36"),
}

// NordTheme Nord 主题
var NordTheme = &Theme{
	Name:       "nord",
	Primary:    ParseColor("#88c0d0"),
	Secondary:  ParseColor("#81a1c1"),
	Success:    ParseColor("#a3be8c"),
	Warning:    ParseColor("#ebcb8b"),
	Error:      ParseColor("#bf616a"),
	Info:       ParseColor("#5e81ac"),
	Foreground: ParseColor("#eceff4"),
	Background: ParseColor("#2e3440"),
}

// GetTheme 获取主题
func GetTheme(name string) *Theme {
	switch name {
	case "light":
		return LightTheme
	case "dracula":
		return DraculaTheme
	case "nord":
		return NordTheme
	default:
		return DefaultTheme
	}
}

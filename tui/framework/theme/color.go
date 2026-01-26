package theme

import (
	"fmt"
	"strconv"
	"strings"
)

// ColorType 颜色类型
type ColorType int

const (
	ColorNamed ColorType = iota // 命名颜色 (如 "red", "blue")
	Color256                   // 256色模式 (0-255)
	ColorRGB                   // RGB 颜色
	ColorHex                   // 十六进制颜色
	ColorNone                  // 无颜色 (透明)
)

// Color 颜色表示（支持多种格式）
type Color struct {
	Type  ColorType
	Value interface{}
}

// 颜色常量
var (
	Black   Color = Color{Type: ColorNamed, Value: "black"}
	Red     Color = Color{Type: ColorNamed, Value: "red"}
	Green   Color = Color{Type: ColorNamed, Value: "green"}
	Yellow  Color = Color{Type: ColorNamed, Value: "yellow"}
	Blue    Color = Color{Type: ColorNamed, Value: "blue"}
	Magenta Color = Color{Type: ColorNamed, Value: "magenta"}
	Cyan    Color = Color{Type: ColorNamed, Value: "cyan"}
	White   Color = Color{Type: ColorNamed, Value: "white"}

	BrightBlack   Color = Color{Type: ColorNamed, Value: "bright-black"}
	BrightRed     Color = Color{Type: ColorNamed, Value: "bright-red"}
	BrightGreen   Color = Color{Type: ColorNamed, Value: "bright-green"}
	BrightYellow  Color = Color{Type: ColorNamed, Value: "bright-yellow"}
	BrightBlue    Color = Color{Type: ColorNamed, Value: "bright-blue"}
	BrightMagenta Color = Color{Type: ColorNamed, Value: "bright-magenta"}
	BrightCyan    Color = Color{Type: ColorNamed, Value: "bright-cyan"}
	BrightWhite   Color = Color{Type: ColorNamed, Value: "bright-white"}

	NoColor Color = Color{Type: ColorNone}
)

// ColorCodes 颜色代码映射 (命名颜色 -> ANSI)
var ColorCodes = map[string]int{
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

// NewColor 创建颜色
func NewColor(value interface{}) Color {
	switch v := value.(type) {
	case int:
		return Color{Type: Color256, Value: v}
	case string:
		return ParseColor(v)
	case [3]int:
		return Color{Type: ColorRGB, Value: v}
	case [3]uint8:
		return Color{Type: ColorRGB, Value: [3]int{int(v[0]), int(v[1]), int(v[2])}}
	default:
		return NoColor
	}
}

// ParseColor 解析颜色字符串
func ParseColor(s string) Color {
	if s == "" {
		return NoColor
	}

	s = strings.TrimSpace(s)

	// 检查十六进制
	if strings.HasPrefix(s, "#") {
		if rgb, err := parseHexColor(s); err == nil {
			return Color{Type: ColorRGB, Value: rgb}
		}
	}

	// 检查命名颜色
	if _, ok := ColorCodes[strings.ToLower(s)]; ok {
		return Color{Type: ColorNamed, Value: strings.ToLower(s)}
	}

	// 检查数字 (256色)
	if n, err := strconv.Atoi(s); err == nil && n >= 0 && n <= 255 {
		return Color{Type: Color256, Value: n}
	}

	// RGB 格式: "rgb(r,g,b)"
	if strings.HasPrefix(strings.ToLower(s), "rgb(") {
		if rgb, err := parseRGBColor(s); err == nil {
			return Color{Type: ColorRGB, Value: rgb}
		}
	}

	// 默认作为命名颜色
	return Color{Type: ColorNamed, Value: strings.ToLower(s)}
}

// parseHexColor 解析十六进制颜色
func parseHexColor(s string) ([3]int, error) {
	s = strings.TrimPrefix(s, "#")
	s = strings.ToLower(s)

	var r, g, b int
	// var err error

	switch len(s) {
	case 3: // #RGB
		r64, err := strconv.ParseInt(string(s[0])+"0", 16, 32)
		if err != nil {
			return [3]int{}, err
		}
		r = int(r64)
		g64, _ := strconv.ParseInt(string(s[1])+"0", 16, 32)
		g = int(g64)
		b64, _ := strconv.ParseInt(string(s[2])+"0", 16, 32)
		b = int(b64)
	case 6: // #RRGGBB
		r64, err := strconv.ParseInt(s[0:2], 16, 32)
		if err != nil {
			return [3]int{}, err
		}
		r = int(r64)
		g64, _ := strconv.ParseInt(s[2:4], 16, 32)
		g = int(g64)
		b64, _ := strconv.ParseInt(s[4:6], 16, 32)
		b = int(b64)
	default:
		return [3]int{}, fmt.Errorf("invalid hex color: #%s", s)
	}

	return [3]int{r, g, b}, nil
}

// parseRGBColor 解析 RGB 颜色
func parseRGBColor(s string) ([3]int, error) {
	s = strings.TrimPrefix(s, "rgb(")
	s = strings.TrimSuffix(s, ")")
	s = strings.ToLower(s)

	parts := strings.Split(s, ",")
	if len(parts) != 3 {
		return [3]int{}, fmt.Errorf("invalid rgb color: %s", s)
	}

	r, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
	g, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
	b, _ := strconv.Atoi(strings.TrimSpace(parts[2]))

	return [3]int{r, g, b}, nil
}

// toANSI 转换为 ANSI 转义码
func (c Color) toANSI(bg bool) string {
	if c.Type == ColorNone {
		return ""
	}

	switch c.Type {
	case ColorNamed:
		if name, ok := c.Value.(string); ok {
			if code, ok := ColorCodes[name]; ok {
				if bg {
					return fmt.Sprintf("%d", code+40)
				}
				return fmt.Sprintf("%d", code+30)
			}
		}

	case Color256:
		if code, ok := c.Value.(int); ok {
			if bg {
				return fmt.Sprintf("48;5;%d", code)
			}
			return fmt.Sprintf("38;5;%d", code)
		}

	case ColorRGB, ColorHex:
		if rgb, ok := c.Value.([3]int); ok {
			if bg {
				return fmt.Sprintf("48;2;%d;%d;%d", rgb[0], rgb[1], rgb[2])
			}
			return fmt.Sprintf("38;2;%d;%d;%d", rgb[0], rgb[1], rgb[2])
		}
	}

	return ""
}

// ToANSIString 返回 ANSI 转义码字符串
func (c Color) ToANSIString() string {
	return c.toANSI(false)
}

// ToANSIBGString 返回 ANSI 背景色转义码字符串
func (c Color) ToANSIBGString() string {
	return c.toANSI(true)
}

// String 字符串表示
func (c Color) String() string {
	switch c.Type {
	case ColorNamed:
		return c.Value.(string)
	case Color256:
		return fmt.Sprintf("ansi:%d", c.Value.(int))
	case ColorRGB, ColorHex:
		rgb := c.Value.([3]int)
		return fmt.Sprintf("#%02x%02x%02x", rgb[0], rgb[1], rgb[2])
	case ColorNone:
		return "none"
	default:
		return "unknown"
	}
}

// IsNone 检查是否为无颜色
func (c Color) IsNone() bool {
	return c.Type == ColorNone
}

// RGBValue 返回 RGB 值（如果不是 RGB 颜色，返回 0,0,0）
func (c Color) RGBValue() (r, g, b int) {
	if c.Type == ColorRGB || c.Type == ColorHex {
		rgb := c.Value.([3]int)
		return rgb[0], rgb[1], rgb[2]
	}
	return 0, 0, 0
}

// WithAlpha 返回带透明度的颜色字符串 (用于支持 truecolor 的终端)
func (c Color) WithAlpha(alpha int) string {
	if c.Type != ColorRGB && c.Type != ColorHex {
		return c.String()
	}
	r, g, b := c.RGBValue()
	return fmt.Sprintf("rgba(%d,%d,%d,%f)", r, g, b, float64(alpha)/255.0)
}

// Lighten 变亮颜色 (返回新颜色)
func (c Color) Lighten(percent int) Color {
	if c.Type != ColorRGB && c.Type != ColorHex {
		return c
	}

	r, g, b := c.RGBValue()
	factor := 1.0 + float64(percent)/100.0

	r = int(float64(r) * factor)
	g = int(float64(g) * factor)
	b = int(float64(b) * factor)

	// Clamp to 0-255
	if r > 255 {
		r = 255
	}
	if g > 255 {
		g = 255
	}
	if b > 255 {
		b = 255
	}

	return Color{Type: ColorRGB, Value: [3]int{r, g, b}}
}

// Darken 变暗颜色 (返回新颜色)
func (c Color) Darken(percent int) Color {
	if c.Type != ColorRGB && c.Type != ColorHex {
		return c
	}

	r, g, b := c.RGBValue()
	factor := 1.0 - float64(percent)/100.0

	r = int(float64(r) * factor)
	g = int(float64(g) * factor)
	b = int(float64(b) * factor)

	// Clamp to 0-255
	if r < 0 {
		r = 0
	}
	if g < 0 {
		g = 0
	}
	if b < 0 {
		b = 0
	}

	return Color{Type: ColorRGB, Value: [3]int{r, g, b}}
}

// Equals 检查颜色是否相等
func (c Color) Equals(other Color) bool {
	if c.Type != other.Type {
		return false
	}

	switch c.Type {
	case ColorNamed:
		return c.Value.(string) == other.Value.(string)
	case Color256:
		return c.Value.(int) == other.Value.(int)
	case ColorRGB, ColorHex:
		rgb1 := c.Value.([3]int)
		rgb2 := other.Value.([3]int)
		return rgb1[0] == rgb2[0] && rgb1[1] == rgb2[1] && rgb1[2] == rgb2[2]
	default:
		return true
	}
}

// ColorPalette 颜色调色板
type ColorPalette struct {
	// 主色
	Primary   Color
	Secondary Color
	Accent    Color

	// 功能色
	Success Color
	Warning Color
	Error   Color
	Info    Color

	// 中性色
	Background Color
	Foreground Color
	Muted      Color

	// 边框色
	Border Color
	Focus  Color

	// 状态色
	Disabled Color
	Hover    Color
	Active   Color
}

// NewColorPalette 创建默认颜色调色板
func NewColorPalette() ColorPalette {
	return ColorPalette{
		Primary:     Blue,
		Secondary:   Cyan,
		Accent:      Yellow,
		Success:     Green,
		Warning:     Yellow,
		Error:       Red,
		Info:        Blue,
		Background:  Black,
		Foreground:  White,
		Muted:       BrightBlack,
		Border:      BrightBlack,
		Focus:       Blue,
		Disabled:    BrightBlack,
		Hover:       Blue,
		Active:      Blue,
	}
}

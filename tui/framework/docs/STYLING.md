# Style System Design

## 概述

样式系统负责为 TUI 组件提供颜色、字体、边框等视觉样式。本文档详细描述了样式系统的设计和实现。

## 核心概念

### 颜色表示

```go
// 位于: tui/framework/style/color.go

package style

// Color 颜色表示
type Color struct {
    Type  ColorType
    Value interface{}
}

// ColorType 颜色类型
type ColorType int

const (
    ColorNamed ColorType = iota  // 命名颜色 (如 "red", "blue")
    Color256                    // 256色模式 (0-255)
    ColorRGB                    // RGB 颜色
    ColorHex                    // 十六进制颜色
    ColorNone                   // 无颜色 (透明)
)

// 颜色常量
const (
    // 标准色
    Black   Color = "black"
    Red     Color = "red"
    Green   Color = "green"
    Yellow  Color = "yellow"
    Blue    Color = "blue"
    Magenta Color = "magenta"
    Cyan    Color = "cyan"
    White   Color = "white"

    // 高亮色
    BrightBlack   Color = "bright-black"
    BrightRed     Color = "bright-red"
    BrightGreen   Color = "bright-green"
    BrightYellow  Color = "bright-yellow"
    BrightBlue    Color = "bright-blue"
    BrightMagenta Color = "bright-magenta"
    BrightCyan    Color = "bright-cyan"
    BrightWhite   Color = "bright-white"

    // 特殊
    NoColor Color = ""
)

// ParseColor 解析颜色字符串
func ParseColor(s string) Color {
    if s == "" {
        return Color{Type: ColorNone}
    }

    // 检查十六进制
    if strings.HasPrefix(s, "#") {
        rgb, err := parseHexColor(s)
        if err == nil {
            return Color{Type: ColorRGB, Value: rgb}
        }
    }

    // 检查命名颜色
    if _, ok := ColorCodes[s]; ok {
        return Color{Type: ColorNamed, Value: s}
    }

    // 检查数字 (256色)
    if n, err := strconv.Atoi(s); err == nil && n >= 0 && n <= 255 {
        return Color{Type: Color256, Value: n}
    }

    // RGB 格式: "rgb(r,g,b)"
    if strings.HasPrefix(s, "rgb(") {
        rgb, err := parseRGBColor(s)
        if err == nil {
            return Color{Type: ColorRGB, Value: rgb}
        }
    }

    return Color{Type: ColorNamed, Value: s}
}

// parseHexColor 解析十六进制颜色
func parseHexColor(s string) ([3]int, error) {
    s = strings.TrimPrefix(s, "#")

    var r, g, b int
    var err error

    switch len(s) {
    case 3:  // #RGB
        r, err = strconv.ParseInt(string(s[0])+"0", 16, 32)
        g, _ = strconv.ParseInt(string(s[1])+"0", 16, 32)
        b, _ = strconv.ParseInt(string(s[2])+"0", 16, 32)
    case 6:  // #RRGGBB
        r, err = strconv.ParseInt(s[0:2], 16, 32)
        g, _ = strconv.ParseInt(s[2:4], 16, 32)
        b, _ = strconv.ParseInt(s[4:6], 16, 32)
    default:
        return [3]int{}, fmt.Errorf("invalid hex color: %s", s)
    }

    return [3]int{r, g, b}, err
}

// parseRGBColor 解析 RGB 颜色
func parseRGBColor(s string) ([3]int, error) {
    s = strings.TrimPrefix(s, "rgb(")
    s = strings.TrimSuffix(s, ")")

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
        if code, ok := ColorCodes[c.Value.(string)]; ok {
            if bg {
                return fmt.Sprintf("%d", code+40)
            }
            return fmt.Sprintf("%d", code+30)
        }

    case Color256:
        code := c.Value.(int)
        if bg {
            return fmt.Sprintf("48;5;%d", code)
        }
        return fmt.Sprintf("38;5;%d", code)

    case ColorRGB, ColorHex:
        rgb := c.Value.([3]int)
        if bg {
            return fmt.Sprintf("48;2;%d;%d;%d", rgb[0], rgb[1], rgb[2])
        }
        return fmt.Sprintf("38;2;%d;%d;%d", rgb[0], rgb[1], rgb[2])
    }

    return ""
}

// ColorCodes 颜色代码映射
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
    "bright-magenta":13,
    "bright-cyan":   14,
    "bright-white":  15,
}
```

### 样式定义

```go
// 位于: tui/framework/style/style.go

package style

// Style 样式定义
type Style struct {
    // 颜色
    FG    Color
    BG    Color

    // 字体样式
    Bold       bool
    Italic     bool
    Underline  bool
    Strikethrough bool
    Blink      bool

    // 其他
    Reverse    bool  // 反白

    // 边框
    Border     BorderStyle

    // 间距
    Padding    BoxSpacing
    Margin     BoxSpacing

    // 宽度
    Width      int
    Height     int
}

// BoxSpacing 间距
type BoxSpacing struct {
    Top    int
    Right  int
    Bottom int
    Left   int
}

// NewStyle 创建默认样式
func NewStyle() Style {
    return Style{}
}

// DefaultStyle 默认样式
func DefaultStyle() Style {
    return Style{
        FG:   White,
        BG:   Black,
    }
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
    s.Bold = on
    return s
}

// Italic 设置斜体
func (s Style) Italic(on bool) Style {
    s.Italic = on
    return s
}

// Underline 设置下划线
func (s Style) Underline(on bool) Style {
    s.Underline = on
    return s
}

// Reverse 设置反白
func (s Style) Reverse(on bool) Style {
    s.Reverse = on
    return s
}

// Border 设置边框
func (s Style) Border(border BorderStyle) Style {
    s.Border = border
    return s
}

// Padding 设置内边距
func (s Style) Padding(all int) Style {
    s.Padding = BoxSpacing{
        Top:    all,
        Right:  all,
        Bottom: all,
        Left:   all,
    }
    return s
}

// PaddingV 设置垂直内边距
func (s Style) PaddingV(vertical int) Style {
    s.Padding.Top = vertical
    s.Padding.Bottom = vertical
    return s
}

// PaddingH 设置水平内边距
func (s Style) PaddingH(horizontal int) Style {
    s.Padding.Left = horizontal
    s.Padding.Right = horizontal
    return s
}

// Merge 合并样式
func (s Style) Merge(other Style) Style {
    result := s

    if other.FG != (Color{}) && other.FG != (Color{Type: ColorNone}) {
        result.FG = other.FG
    }
    if other.BG != (Color{}) && other.BG != (Color{Type: ColorNone}) {
        result.BG = other.BG
    }
    if other.Bold {
        result.Bold = true
    }
    if other.Italic {
        result.Italic = true
    }
    if other.Underline {
        result.Underline = true
    }
    if other.Reverse {
        result.Reverse = true
    }

    return result
}

// ToANSI 转换为 ANSI 转义码
func (s Style) ToANSI() string {
    var codes []string

    // 前景色
    if fg := s.FG.toANSI(false); fg != "" {
        codes = append(codes, fg)
    }

    // 背景色
    if bg := s.BG.toANSI(true); bg != "" {
        codes = append(codes, bg)
    }

    // 粗体
    if s.Bold {
        codes = append(codes, "1")
    }

    // 斜体
    if s.Italic {
        codes = append(codes, "3")
    }

    // 下划线
    if s.Underline {
        codes = append(codes, "4")
    }

    // 反白
    if s.Reverse {
        codes = append(codes, "7")
    }

    // 闪烁
    if s.Blink {
        codes = append(codes, "5")
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

// Clone 克隆样式
func (s Style) Clone() Style {
    return s
}

// IsEmpty 检查样式是否为空
func (s Style) IsEmpty() bool {
    return s == Style{}
}
```

### 边框样式

```go
// 位于: tui/framework/style/border.go

package style

// BorderStyle 边框样式
type BorderStyle struct {
    Enabled bool
    Top     bool
    Bottom  bool
    Left    bool
    Right   bool
    Style   BorderType
    FG      Color
    BG      Color
}

// BorderType 边框类型
type BorderType int

const (
    BorderNormal BorderType = iota
    BorderRounded
    BorderDouble
    BorderThick
    BorderHidden
    BorderDashed
    BorderDotted
)

// BorderChars 边框字符
var BorderChars = map[BorderType]BorderEdges{
    BorderNormal: {
        TL: "┌", T: "─", TR: "┐",
        L: "│", R: "│",
        BL: "└", B: "─", BR: "┘",
    },
    BorderRounded: {
        TL: "╭", T: "─", TR: "╮",
        L: "│", R: "│",
        BL: "╰", B: "─", BR: "╯",
    },
    BorderDouble: {
        TL: "╔", T: "═", TR: "╗",
        L: "║", R: "║",
        BL: "╚", B: "═", BR: "╝",
    },
    BorderThick: {
        TL: "┏", T: "━", TR: "┓",
        L: "┃", R: "┃",
        BL: "┗", B: "━", BR: "┛",
    },
    BorderDashed: {
        TL: "┌", T: "┄", TR: "┐",
        L: "┆", R: "┆",
        BL: "└", B: "┄", BR: "┘",
    },
    BorderDotted: {
        TL: "┌", T: "┈", TR: "┐",
        L: "┊", R: "┊",
        BL: "└", B: "┈", BR: "┘",
    },
}

// BorderEdges 边框边缘字符
type BorderEdges struct {
    TL string  // Top Left
    T  string  // Top
    TR string  // Top Right
    L  string  // Left
    R  string  // Right
    BL string  // Bottom Left
    B  string  // Bottom
    BR string  // Bottom Right
}

// NewBorder 创建边框
func NewBorder() BorderStyle {
    return BorderStyle{
        Enabled: true,
        Top:     true,
        Bottom:  true,
        Left:    true,
        Right:   true,
        Style:   BorderNormal,
    }
}

// All 启用所有边
func (b BorderStyle) All() BorderStyle {
    b.Top = true
    b.Bottom = true
    b.Left = true
    b.Right = true
    return b
}

// TopOnly 只启用顶部边
func (b BorderStyle) TopOnly() BorderStyle {
    b.Top = true
    b.Bottom = false
    b.Left = false
    b.Right = false
    return b
}

// BottomOnly 只启用底部边
func (b BorderStyle) BottomOnly() BorderStyle {
    b.Top = false
    b.Bottom = true
    b.Left = false
    b.Right = false
    return b
}

// Sides 只启用左右边
func (b BorderStyle) Sides() BorderStyle {
    b.Top = false
    b.Bottom = false
    b.Left = true
    b.Right = true
    return b
}

// WithStyle 设置边框类型
func (b BorderStyle) WithStyle(style BorderType) BorderStyle {
    b.Style = style
    return b
}

// WithColor 设置边框颜色
func (b BorderStyle) WithColor(fg Color) BorderStyle {
    b.FG = fg
    return b
}

// GetEdges 获取边框字符
func (b BorderStyle) GetEdges() BorderEdges {
    return BorderChars[b.Style]
}

// Render 渲染边框
func (b BorderStyle) Render(width, height int) []string {
    if !b.Enabled {
        return []string{}
    }

    edges := b.GetEdges()
    lines := make([]string, height)

    // 顶部边框
    if b.Top {
        top := edges.TL
        for i := 0; i < width-2; i++ {
            top += edges.T
        }
        top += edges.TR
        lines[0] = b.FG.Apply(top)
    }

    // 中间行
    for y := 1; y < height-1; y++ {
        line := ""
        if b.Left {
            line += edges.L
        }
        for i := 0; i < width-2; i++ {
            line += " "
        }
        if b.Right {
            line += edges.R
        }
        lines[y] = b.FG.Apply(line)
    }

    // 底部边框
    if b.Bottom && height > 1 {
        bottom := edges.BL
        for i := 0; i < width-2; i++ {
            bottom += edges.B
        }
        bottom += edges.BR
        lines[height-1] = b.FG.Apply(bottom)
    }

    return lines
}
```

### 样式构建器

```go
// 位于: tui/framework/style/builder.go

package style

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
    b.style.Bold = true
    return b
}

// Italic 设置斜体
func (b *Builder) Italic() *Builder {
    b.style.Italic = true
    return b
}

// Underline 设置下划线
func (b *Builder) Underline() *Builder {
    b.style.Underline = true
    return b
}

// Reverse 设置反白
func (b *Builder) Reverse() *Builder {
    b.style.Reverse = true
    return b
}

// Blink 设置闪烁
func (b *Builder) Blink() *Builder {
    b.style.Blink = true
    return b
}

// Padding 设置内边距
func (b *Builder) Padding(all int) *Builder {
    b.style.Padding = BoxSpacing{
        Top:    all,
        Right:  all,
        Bottom: all,
        Left:   all,
    }
    return b
}

// PaddingV 设置垂直内边距
func (b *Builder) PaddingV(vertical int) *Builder {
    b.style.Padding.Top = vertical
    b.style.Padding.Bottom = vertical
    return b
}

// PaddingH 设置水平内边距
func (b *Builder) PaddingH(horizontal int) *Builder {
    b.style.Padding.Left = horizontal
    b.style.Padding.Right = horizontal
    return b
}

// Border 设置边框
func (b *Builder) Border(borderType BorderType) *Builder {
    b.style.Border = BorderStyle{
        Enabled: true,
        Style:   borderType,
    }
    return b
}

// BorderColor 设置边框颜色
func (b *Builder) BorderColor(color string) *Builder {
    b.style.Border.FG = ParseColor(color)
    return b
}

// Width 设置宽度
func (b *Builder) Width(w int) *Builder {
    b.style.Width = w
    return b
}

// Height 设置高度
func (b *Builder) Height(h int) *Builder {
    b.style.Height = h
    return b
}

// Build 构建样式
func (b *Builder) Build() Style {
    return b.style
}
```

## 主题系统

### 主题定义

```go
// 位于: tui/framework/style/theme.go

package style

// Theme 主题
type Theme struct {
    // 名称
    Name string

    // 主色
    Primary   Color
    Secondary Color

    // 功能色
    Success   Color
    Warning   Color
    Error     Color
    Info      Color

    // 基础色
    Foreground Color
    Background Color

    // 组件样式
    Button    ButtonStyle
    Input     InputStyle
    List      ListStyle
    Table     TableStyle
    Border    BorderStyle
}

// ButtonStyle 按钮样式
type ButtonStyle struct {
    Normal    Style
    Focused   Style
    Active    Style
    Disabled  Style
}

// InputStyle 输入框样式
type InputStyle struct {
    Normal    Style
    Focused   Style
    Error     Style
    Cursor    Style
    Selection Style
}

// ListStyle 列表样式
type ListStyle struct {
    Normal    Style
    Focused   Style
    Selected  Style
    Dimmed    Style
}

// TableStyle 表格样式
type TableStyle struct {
    Header    Style
    Row       Style
    AltRow    Style
    Focused   Style
    Selected  Style
}

// NewTheme 创建主题
func NewTheme(name string) *Theme {
    return &Theme{
        Name: name,
    }
}

// GetColor 获取颜色
func (t *Theme) GetColor(name string) Color {
    switch name {
    case "primary":
        return t.Primary
    case "secondary":
        return t.Secondary
    case "success":
        return t.Success
    case "warning":
        return t.Warning
    case "error":
        return t.Error
    case "info":
        return t.Info
    case "fg", "foreground":
        return t.Foreground
    case "bg", "background":
        return t.Background
    default:
        return Color{}
    }
}

// GetStyle 获取样式
func (t *Theme) GetStyle(name string) Style {
    switch name {
    case "button":
        return t.Button.Normal
    case "button.focused":
        return t.Button.Focused
    case "input":
        return t.Input.Normal
    case "input.focused":
        return t.Input.Focused
    default:
        return Style{}
    }
}
```

### 预设主题

```go
// 位于: tui/framework/style/themes.go

package style

// DefaultTheme 默认主题
var DefaultTheme = &Theme{
    Name: "default",

    Primary:   Blue,
    Secondary: Cyan,

    Success:   Green,
    Warning:   Yellow,
    Error:     Red,
    Info:      Blue,

    Foreground: White,
    Background: Black,

    Button: ButtonStyle{
        Normal:   Style{}.Foreground(White).Background(Blue).Bold(),
        Focused:  Style{}.Foreground(Blue).Background(White).Bold(),
        Active:   Style{}.Foreground(Black).Background(White),
        Disabled: Style{}.Foreground(BrightBlack).Background(Black),
    },

    Input: InputStyle{
        Normal:    Style{}.Foreground(White).Background(Black),
        Focused:   Style{}.Foreground(White).Background(Blue),
        Error:     Style{}.Foreground(Red).Background(Black),
        Cursor:    Style{}.Foreground(Black).Background(White).Reverse(),
        Selection: Style{}.Foreground(Black).Background(Cyan),
    },

    List: ListStyle{
        Normal:   Style{}.Foreground(White),
        Focused:  Style{}.Foreground(Black).Background(Blue).Bold(),
        Selected: Style{}.Foreground(Black).Background(Cyan),
        Dimmed:   Style{}.Foreground(BrightBlack),
    },

    Table: TableStyle{
        Header:   Style{}.Foreground(Yellow).Bold(),
        Row:      Style{}.Foreground(White),
        AltRow:   Style{}.Foreground(BrightWhite),
        Focused:  Style{}.Foreground(Black).Background(Blue).Bold(),
        Selected: Style{}.Foreground(Black).Background(Cyan),
    },

    Border: BorderStyle{
        Enabled: true,
        Style:   BorderNormal,
        FG:      BrightBlack,
    },
}

// LightTheme 亮色主题
var LightTheme = &Theme{
    Name: "light",

    Primary:   Blue,
    Secondary: Cyan,

    Success:   Green,
    Warning:   Yellow,
    Error:     Red,
    Info:      Blue,

    Foreground: Black,
    Background: White,

    Button: ButtonStyle{
        Normal:   Style{}.Foreground(White).Background(Blue).Bold(),
        Focused:  Style{}.Foreground(Blue).Background(White).Bold(),
        Active:   Style{}.Foreground(Black).Background(White),
        Disabled: Style{}.Foreground(BrightBlack).Background(White),
    },

    Input: InputStyle{
        Normal:    Style{}.Foreground(Black).Background(White),
        Focused:   Style{}.Foreground(Black).Background(Cyan),
        Error:     Style{}.Foreground(Red).Background(White),
        Cursor:    Style{}.Foreground(White).Background(Black).Reverse(),
        Selection: Style{}.Foreground(White).Background(Blue),
    },

    List: ListStyle{
        Normal:   Style{}.Foreground(Black),
        Focused:  Style{}.Foreground(White).Background(Blue).Bold(),
        Selected: Style{}.Foreground(White).Background(Cyan),
        Dimmed:   Style{}.Foreground(BrightBlack),
    },

    Table: TableStyle{
        Header:   Style{}.Foreground(Blue).Bold(),
        Row:      Style{}.Foreground(Black),
        AltRow:   Style{}.Foreground(BrightBlack),
        Focused:  Style{}.Foreground(White).Background(Blue).Bold(),
        Selected: Style{}.Foreground(White).Background(Cyan),
    },

    Border: BorderStyle{
        Enabled: true,
        Style:   BorderNormal,
        FG:      BrightBlack,
    },
}

// DraculaTheme Dracula 主题
var DraculaTheme = &Theme{
    Name: "dracula",

    Primary:   ParseColor("#bd93f9"),  // Purple
    Secondary: ParseColor("#ff79c6"),  // Pink

    Success:   ParseColor("#50fa7b"),  // Green
    Warning:   ParseColor("#f1fa8c"),  // Yellow
    Error:     ParseColor("#ff5555"),  // Red
    Info:      ParseColor("#8be9fd"),  // Cyan

    Foreground: ParseColor("#f8f8f2"),  // White
    Background: ParseColor("#282a36"),  // Background

    Button: ButtonStyle{
        Normal:   Style{}.Foreground(ParseColor("#282a36")).Background(ParseColor("#bd93f9")).Bold(),
        Focused:  Style{}.Foreground(ParseColor("#bd93f9")).Background(ParseColor("#f8f8f2")).Bold(),
        Active:   Style{}.Foreground(ParseColor("#282a36")).Background(ParseColor("#f8f8f2")),
        Disabled: Style{}.Foreground(ParseColor("#6272a4")).Background(ParseColor("#282a36")),
    },

    Input: InputStyle{
        Normal:    Style{}.Foreground(ParseColor("#f8f8f2")).Background(ParseColor("#44475a")),
        Focused:   Style{}.Foreground(ParseColor("#f8f8f2")).Background(ParseColor("#6272a4")),
        Error:     Style{}.Foreground(ParseColor("#ff5555")).Background(ParseColor("#282a36")),
        Cursor:    Style{}.Foreground(ParseColor("#282a36")).Background(ParseColor("#f8f8f2")).Reverse(),
        Selection: Style{}.Foreground(ParseColor("#282a36")).Background(ParseColor("#bd93f9")),
    },

    List: ListStyle{
        Normal:   Style{}.Foreground(ParseColor("#f8f8f2")),
        Focused:  Style{}.Foreground(ParseColor("#282a36")).Background(ParseColor("#bd93f9")).Bold(),
        Selected: Style{}.Foreground(ParseColor("#282a36")).Background(ParseColor("#ff79c6")),
        Dimmed:   Style{}.Foreground(ParseColor("#6272a4")),
    },

    Table: TableStyle{
        Header:   Style{}.Foreground(ParseColor("#bd93f9")).Bold(),
        Row:      Style{}.Foreground(ParseColor("#f8f8f2")),
        AltRow:   Style{}.Foreground(ParseColor("#f8f8f2")),
        Focused:  Style{}.Foreground(ParseColor("#282a36")).Background(ParseColor("#bd93f9")).Bold(),
        Selected: Style{}.Foreground(ParseColor("#282a36")).Background(ParseColor("#ff79c6")),
    },

    Border: BorderStyle{
        Enabled: true,
        Style:   BorderNormal,
        FG:      ParseColor("#6272a4"),
    },
}

// NordTheme Nord 主题
var NordTheme = &Theme{
    Name: "nord",

    Primary:   ParseColor("#88c0d0"),  // Frost 8
    Secondary: ParseColor("#81a1c1"),  // Frost 7

    Success:   ParseColor("#a3be8c"),  // Aurora 5
    Warning:   ParseColor("#ebcb8b"),  // Aurora 3
    Error:     ParseColor("#bf616a"),  // Aurora 1
    Info:      ParseColor("#5e81ac"),  // Frost 4

    Foreground: ParseColor("#eceff4"),  // Polar 6
    Background: ParseColor("#2e3440"),  // Nordic 0

    Button: ButtonStyle{
        Normal:   Style{}.Foreground(ParseColor("#2e3440")).Background(ParseColor("#88c0d0")).Bold(),
        Focused:  Style{}.Foreground(ParseColor("#88c0d0")).Background(ParseColor("#eceff4")).Bold(),
        Active:   Style{}.Foreground(ParseColor("#2e3440")).Background(ParseColor("#eceff4")),
        Disabled: Style{}.Foreground(ParseColor("#4c566a")).Background(ParseColor("#2e3440")),
    },

    Input: InputStyle{
        Normal:    Style{}.Foreground(ParseColor("#eceff4")).Background(ParseColor("#3b4252")),
        Focused:   Style{}.Foreground(ParseColor("#eceff4")).Background(ParseColor("#434c5e")),
        Error:     Style{}.Foreground(ParseColor("#bf616a")).Background(ParseColor("#2e3440")),
        Cursor:    Style{}.Foreground(ParseColor("#2e3440")).Background(ParseColor("#eceff4")).Reverse(),
        Selection: Style{}.Foreground(ParseColor("#2e3440")).Background(ParseColor("#88c0d0")),
    },

    List: ListStyle{
        Normal:   Style{}.Foreground(ParseColor("#eceff4")),
        Focused:  Style{}.Foreground(ParseColor("#2e3440")).Background(ParseColor("#88c0d0")).Bold(),
        Selected: Style{}.Foreground(ParseColor("#2e3440")).Background(ParseColor("#81a1c1")),
        Dimmed:   Style{}.Foreground(ParseColor("#4c566a")),
    },

    Table: TableStyle{
        Header:   Style{}.Foreground(ParseColor("#88c0d0")).Bold(),
        Row:      Style{}.Foreground(ParseColor("#eceff4")),
        AltRow:   Style{}.Foreground(ParseColor("#d8dee9")),
        Focused:  Style{}.Foreground(ParseColor("#2e3440")).Background(ParseColor("#88c0d0")).Bold(),
        Selected: Style{}.Foreground(ParseColor("#2e3440")).Background(ParseColor("#81a1c1")),
    },

    Border: BorderStyle{
        Enabled: true,
        Style:   BorderNormal,
        FG:      ParseColor("#4c566a"),
    },
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
```

## 样式应用

### 组件样式

```go
// 位于: tui/framework/style/styled.go

package style

// Styled 可样式化接口
type Styled interface {
    SetStyle(style Style)
    GetStyle() Style
}

// StyledComponent 可样式化组件
type StyledComponent struct {
    BaseComponent
    style    Style
    theme    *Theme
}

// SetStyle 设置样式
func (c *StyledComponent) SetStyle(style Style) {
    c.style = style
}

// GetStyle 获取样式
func (c *StyledComponent) GetStyle() Style {
    return c.style
}

// SetTheme 设置主题
func (c *StyledComponent) SetTheme(theme *Theme) {
    c.theme = theme
}

// GetTheme 获取主题
func (c *StyledComponent) GetTheme() *Theme {
    return c.theme
}

// GetEffectiveStyle 获取有效样式 (合并主题和组件样式)
func (c *StyledComponent) GetEffectiveStyle() Style {
    if c.theme == nil {
        return c.style
    }

    // 从主题获取基础样式
    themeStyle := c.getThemeStyle()

    // 合并组件样式
    return themeStyle.Merge(c.style)
}

// getThemeStyle 从主题获取样式
func (c *StyledComponent) getThemeStyle() Style {
    // 由子类实现
    return Style{}
}
```

### 内联样式

```go
// 位于: tui/framework/style/inline.go

package style

// InlineStyle 内联样式
type InlineStyle struct {
    styles map[string]Style
}

// NewInlineStyle 创建内联样式
func NewInlineStyle() *InlineStyle {
    return &InlineStyle{
        styles: make(map[string]Style),
    }
}

// Set 设置样式
func (s *InlineStyle) Set(selector string, style Style) {
    s.styles[selector] = style
}

// Get 获取样式
func (s *InlineStyle) Get(selector string) Style {
    return s.styles[selector]
}

// Apply 应用样式到组件
func (s *InlineStyle) Apply(comp Component) {
    if styled, ok := comp.(Styled); ok {
        selector := getComponentSelector(comp)
        if style, ok := s.styles[selector]; ok {
            styled.SetStyle(style)
        }
    }
}

// getComponentSelector 获取组件选择器
func getComponentSelector(comp Component) string {
    // 生成选择器: "type#id.class"
    var parts []string

    parts = append(parts, strings.ToLower(comp.GetType()))

    if id := comp.GetID(); id != "" {
        parts = append(parts, "#"+id)
    }

    if class := comp.GetClass(); class != "" {
        parts = append(parts, "."+class)
    }

    return strings.Join(parts, "")
}
```

## 使用示例

```go
// 使用样式构建器
style := NewBuilder().
    Foreground("blue").
    Background("white").
    Bold().
    Underline().
    Build()

text := style.Apply("Hello, World!")
// 输出: "\x1b[34;47;1;4mHello, World!\x1b[0m"

// 使用主题
theme := GetTheme("dracula")
buttonStyle := theme.Button.Focused
text := buttonStyle.Apply("Click Me")

// 应用到组件
button := NewButton("Click Me")
button.SetStyle(style.NewBuilder().
    Foreground("green").
    Background("black").
    Border(style.BorderRounded).
    Build())

// 设置组件主题
app.SetTheme(LightTheme)
```

## 样式继承

```go
// 样式继承规则
// 1. 主题样式 (最低优先级)
// 2. 父组件样式
// 3. 组件默认样式
// 4. 内联样式 (最高优先级)

// 位于: tui/framework/style/inheritance.go

package style

// Inheritance 样式继承
type Inheritance struct {
    theme    *Theme
    parent   Style
    current  Style
}

// NewInheritance 创建样式继承
func NewInheritance(theme *Theme) *Inheritance {
    return &Inheritance{
        theme: theme,
    }
}

// SetParent 设置父组件样式
func (i *Inheritance) SetParent(style Style) {
    i.parent = style
}

// SetCurrent 设置当前组件样式
func (i *Inheritance) SetCurrent(style Style) {
    i.current = style
}

// Resolve 解析最终样式
func (i *Inheritance) Resolve() Style {
    result := Style{}

    // 1. 主题样式
    if i.theme != nil {
        result = i.theme.GetStyle("default")
    }

    // 2. 父组件样式
    result = result.Merge(i.parent)

    // 3. 当前组件样式
    result = result.Merge(i.current)

    return result
}
```

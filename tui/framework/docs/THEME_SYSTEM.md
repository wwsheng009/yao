# Theme System Design (V3)

> **优先级**: P0 (核心功能)
> **目标**: 支持运行时 UI 主题切换
> **关键特性**: 主题继承、样式覆盖、动画过渡

## 概述

主题系统负责管理 UI 的视觉外观，包括颜色、字体、间距等。V3 架构中，主题与组件解耦，支持运行时切换，并支持主题间的平滑过渡。

### 为什么需要主题系统？

**没有主题系统的问题**：
```go
// ❌ 样式硬编码
func (t *Text) Paint(ctx PaintContext, buf *CellBuffer) {
    style := lipgloss.NewStyle().Foreground(lipgloss.Color(" blue "))  // 硬编码
    buf.SetCell(x, y, char, style)
}

// 问题：
// - 无法切换主题
// - 无法统一管理样式
// - 维护困难
```

**有主题系统的优势**：
```go
// ✅ 主题驱动
func (t *Text) Paint(ctx PaintContext, buf *CellBuffer) {
    theme := ctx.GetTheme()
    style := theme.GetStyle("text.primary")
    buf.SetCell(x, y, char, style)
}

// 优势：
// - 支持主题切换
// - 样式集中管理
// - 易于维护和扩展
```

## 设计目标

1. **解耦**: 组件不依赖具体的样式库
2. **可切换**: 运行时切换主题，支持热切换
3. **可继承**: 支持主题继承和覆盖
4. **高性能**: 主题切换不重建组件树
5. **可扩展**: 支持自定义主题

## 核心类型定义

### 1. Theme 结构

```go
// 位于: tui/framework/theme/theme.go

package theme

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

    // 字体
    Fonts FontSet

    // 间距
    Spacing SpacingSet

    // 样式映射
    Styles map[string]StyleConfig

    // 组件特定配置
    Components map[string]ComponentStyle

    // 元数据
    Metadata map[string]interface{}
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
    Muted     Color

    // 边框色
    Border Color
    Focus  Color

    // 状态色
    Disabled Color
    Hover    Color
    Active   Color
}

// FontSet 字体集
type FontSet struct {
    Default FontStyle
    Bold    FontStyle
    Italic  FontStyle
    Code    FontStyle
}

// FontStyle 字体样式
type FontStyle struct {
    Family string
    Size   int
    Weight int
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
    Padding    *[4]int  // top, right, bottom, left
    Margin     *[4]int
    Width      *int
    Height     *int
}

// ComponentStyle 组件样式
type ComponentStyle struct {
    Base   StyleConfig
    States map[string]StyleConfig  // hover, focus, disabled, etc.
}
```

### 2. Color 类型

```go
// 位于: tui/framework/theme/color.go

package theme

// Color 颜色（支持 ANSI 和 RGB）
type Color struct {
    // ANSI 颜色（0-255）
    ANSI *uint8

    // RGB 颜色
    RGB *[3]uint8  // [R, G, B]

    // 十六进制表示
    Hex string
}

// Predefined ANSI colors
const (
    ColorBlack  ColorValue = iota + 90
    ColorRed
    ColorGreen
    ColorYellow
    ColorBlue
    ColorMagenta
    ColorCyan
    ColorWhite
)

// ColorValue 颜色值类型
type ColorValue int

// NewColor 创建颜色
func NewColor(value interface{}) Color {
    switch v := value.(type) {
    case ColorValue:
        return Color{ANSI: uint8(v)}
    case string:
        return ParseHexColor(v)
    case [3]int:
        return Color{RGB: &[3]uint8{uint8(v[0]), uint8(v[1]), uint8(v[2])}}
    case [3]uint8:
        return Color{RGB: &[3]uint8{v[0], v[1], v[2]}}
    default:
        return Color{}  // 默认颜色
    }
}

// ParseHexColor 解析十六进制颜色
func ParseHexColor(hex string) Color {
    // 支持 #RGB 或 #RRGGBB
    return Color{Hex: hex}
}

// ToLipgloss 转换为 lipgloss.Color
func (c Color) ToLipgloss() lipgloss.Color {
    if c.ANSI != nil {
        return lipgloss.Color(*c.ANSI)
    }
    if c.RGB != nil {
        return lipgloss.Color(c.RGB[0], c.RGB[1], c.RGB[2])
    }
    return lipgloss.Color("")
}

// String 字符串表示
func (c Color) String() string {
    if c.Hex != "" {
        return c.Hex
    }
    if c.ANSI != nil {
        return fmt.Sprintf("ansi:%d", *c.ANSI)
    }
    if c.RGB != nil {
        return fmt.Sprintf("rgb:%d,%d,%d", c.RGB[0], c.RGB[1], c.RGB[2])
    }
    return "default"
}
```

### 3. Theme Manager

```go
// 位于: tui/framework/theme/manager.go

package theme

// Manager 主题管理器
type Manager struct {
    mu sync.RWMutex

    // 当前主题
    current *Theme

    // 主题注册表
    themes map[string]*Theme

    // 主题变化监听器
    listeners []ThemeChangeListener

    // 切换动画配置
    transitionDuration time.Duration
}

// ThemeChangeListener 主题变化监听器
type ThemeChangeListener func(old, new *Theme)

// NewManager 创建主题管理器
func NewManager() *Manager {
    return &Manager{
        themes: make(map[string]*Theme),
        listeners: make([]ThemeChangeListener, 0),
        transitionDuration: 300 * time.Millisecond,
    }
}

// Register 注册主题
func (m *Manager) Register(theme *Theme) {
    m.mu.Lock()
    defer m.mu.Unlock()

    m.themes[theme.Name] = theme

    // 如果是第一个主题，设为默认
    if m.current == nil {
        m.current = theme
    }
}

// Get 获取主题
func (m *Manager) Get(name string) (*Theme, bool) {
    m.mu.RLock()
    defer m.mu.RUnlock()

    theme, ok := m.themes[name]
    return theme, ok
}

// Current 获取当前主题
func (m *Manager) Current() *Theme {
    m.mu.RLock()
    defer m.mu.RUnlock()

    return m.current
}

// Set 设置当前主题
func (m *Manager) Set(name string) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    theme, ok := m.themes[name]
    if !ok {
        return fmt.Errorf("theme not found: %s", name)
    }

    old := m.current
    m.current = theme

    // 通知监听器
    m.notify(old, theme)

    return nil
}

// Toggle 切换到下一个主题
func (m *Manager) Toggle() error {
    m.mu.Lock()
    defer m.mu.Unlock()

    names := m.themeNames()
    if len(names) == 0 {
        return nil
    }

    currentName := ""
    if m.current != nil {
        currentName = m.current.Name
    }

    // 找到下一个主题
    for i, name := range names {
        if name == currentName {
            if i < len(names)-1 {
                return m.Set(names[i+1])
            } else {
                return m.Set(names[0])
            }
        }
    }

    return m.Set(names[0])
}

// Subscribe 订阅主题变化
func (m *Manager) Subscribe(listener ThemeChangeListener) func() {
    m.mu.Lock()
    defer m.mu.Unlock()

    m.listeners = append(m.listeners, listener)

    return func() {
        m.Unsubscribe(listener)
    }
}

// Unsubscribe 取消订阅
func (m *Manager) Unsubscribe(listener ThemeChangeListener) {
    m.mu.Lock()
    defer m.mu.Unlock()

    for i, l := range m.listeners {
        if reflect.ValueOf(l).Pointer() == reflect.ValueOf(listener).Pointer() {
            m.listeners = append(m.listeners[:i], m.listeners[i+1:]...)
            break
        }
    }
}

// notify 通知监听器
func (m *Manager) notify(old, new *Theme) {
    listeners := make([]ThemeChangeListener, len(m.listeners))
    copy(listeners, m.listeners)

    for _, listener := range listeners {
        listener(old, new)
    }
}

// themeNames 获取所有主题名称
func (m *Manager) themeNames() []string {
    names := make([]string, 0, len(m.themes))
    for name := range m.themes {
        names = append(names, name)
    }
    sort.Strings(names)
    return names
}

// GetStyle 获取样式
func (m *Manager) GetStyle(componentID, styleKey string) StyleConfig {
    m.mu.RLock()
    defer m.mu.RUnlock()

    // 1. 查找组件特定样式
    if compStyle, ok := m.current.Components[componentID]; ok {
        if baseStyle, ok := compStyle.States[styleKey]; ok {
            return baseStyle
        }
        return compStyle.Base
    }

    // 2. 查找全局样式
    if style, ok := m.current.Styles[styleKey]; ok {
        return style
    }

    // 3. 查找父主题
    if m.current.Parent != nil {
        return m.resolveStyle(m.current.Parent, componentID, styleKey)
    }

    // 默认样式
    return DefaultStyleConfig()
}

// resolveStyle 递归解析样式
func (m *Manager) resolveStyle(theme *Theme, componentID, styleKey string) StyleConfig {
    // 1. 查找组件特定样式
    if compStyle, ok := theme.Components[componentID]; ok {
        if baseStyle, ok := compStyle.States[styleKey]; ok {
            return baseStyle
        }
        if compStyle.Base.Foreground != nil || compStyle.Base.Background != nil {
            return compStyle.Base
        }
    }

    // 2. 查找全局样式
    if style, ok := theme.Styles[styleKey]; ok {
        return style
    }

    // 3. 递归查找父主题
    if theme.Parent != nil {
        return m.resolveStyle(theme.Parent, componentID, styleKey)
    }

    return DefaultStyleConfig()
}

// GetColor 获取颜色
func (m *Manager) GetColor(colorKey string) Color {
    m.mu.RLock()
    defer m.mu.RUnlock()

    theme := m.current
    for theme != nil {
        if color := theme.resolveColor(colorKey); color != (Color{}) {
            return color
        }
        theme = theme.Parent
    }

    return Color{}
}

// resolveColor 解析颜色
func (t *Theme) resolveColor(key string) Color {
    // 支持嵌套访问，如 "primary.light"
    parts := strings.Split(key, ".")

    var current interface{} = t.Colors
    for _, part := range parts {
        switch v := current.(type) {
        case Color:
            return v
        case map[string]interface{}:
            var ok bool
            current, ok = v[part]
            if !ok {
                return Color{}
            }
        default:
            return Color{}
        }
    }

    if color, ok := current.(Color); ok {
        return color
    }
    return Color{}
}

// DefaultStyleConfig 默认样式配置
func DefaultStyleConfig() StyleConfig {
    return StyleConfig{}
}
```

## 预定义主题

### 内置主题

```go
// 位于: tui/framework/theme/builtin.go

package theme

// LightTheme 亮色主题
var LightTheme = &Theme{
    Name:    "light",
    Version: "1.0.0",
    Colors: ColorPalette{
        Primary:   Color{RGB: &[3]uint8{66, 133, 244},  // 蓝色
        Secondary: Color{RGB: &[3]uint8{236, 72, 153}},   // 紫色
        Accent:    Color{RGB: &[3]uint8{255, 193, 7}},    // 橙色
        Success:   Color{RGB: &[3]uint8{76, 175, 80}},    // 绿色
        Warning:   Color{RGB: &[3]uint8{255, 193, 7}},    // 橙色
        Error:     Color{RGB: &[3]uint8{244, 67, 54}},     // 红色
        Info:      Color{RGB: &[3]uint8{66, 165, 245}},    // 青色
        Background: Color{RGB: &[3]uint8{255, 255, 255}}, // 白色
        Foreground: Color{RGB: &[3]uint8{30, 41, 59}},     // 深灰
        Muted:     Color{RGB: &[3]uint8{148, 163, 184}},   // 灰色
        Border:     Color{RGB: &[3]uint8{227, 233, 240}},   // 浅灰
        Focus:      Color{RGB: &[3]uint8{66, 133, 244}},   // 蓝色
        Disabled:   Color{RGB: &[3]uint8{201, 203, 207}},   // 灰色
        Hover:      Color{RGB: &[3]uint8{66, 133, 244}},   // 蓝色
        Active:     Color{RGB: &[3]uint8{66, 133, 244}},   // 蓝色
    },
    Styles: map[string]StyleConfig{
        "text.primary": {
            Foreground: &Color{RGB: &[3]uint8{30, 41, 59}},
        },
        "text.secondary": {
            Foreground: &Color{RGB: &[3]uint8{148, 163, 184}},
        },
        "border.default": {
            Foreground: &Color{RGB: &[3]uint8{227, 233, 240}},
        },
    },
}

// DarkTheme 深色主题
var DarkTheme = &Theme{
    Name:    "dark",
    Version: "1.0.0",
    Colors: ColorPalette{
        Primary:   Color{RGB: &[3]uint8{97, 175, 239},   // 亮蓝
        Secondary: Color{RGB: &[3]uint8{224, 108, 117}},  // 粉红
        Accent:    Color{RGB: &[3]uint8{255, 213, 79}},    // 金色
        Success:   Color{RGB: &[3]uint8{134, 239, 172}},   // 绿色
        Warning:   Color{RGB: &[3]uint8{255, 213, 79}},    // 金色
        Error:     Color{RGB: &[3]uint8{239, 68, 68}},     // 红色
        Info:      ColorRGB: &[3]uint8{66, 165, 245}},    // 青色
        Background: ColorRGB: &[3]uint8{17, 24, 39}},     // 深蓝黑
        Foreground: ColorRGB: &[3]uint8{227, 233, 240}},   // 浅灰
        Muted:     ColorRGB: &[3]uint8{161, 161, 170}},    // 灰色
        Border:     ColorRGB: &[3]uint8{55, 65, 81}},      // 深灰
        Focus:      ColorRGB: &[3]uint8{97, 175, 239}},   // 亮蓝
        Disabled:   ColorRGB: &[3]uint8{86, 95, 105}},    // 暗灰
        Hover:      ColorRGB: &[3]uint8{97, 175, 239}},   // 亮蓝
        Active:     ColorRGB: &[3]uint8{97, 175, 239}},   // 亮蓝
    },
    Styles: map[string]StyleConfig{
        "text.primary": {
            Foreground: &Color{RGB: &[3]uint8{227, 233, 240}},
        },
        "text.secondary": {
            Foreground: &Color{RGB: &[3]uint8{161, 161, 170}},
        },
        "border.default": {
            Foreground: &Color{RGB: &[3]uint8{55, 65, 81}},
        },
    },
}
```

## 组件集成

### Themed 组件接口

```go
// 位于: tui/framework/component/themed.go

package component

// Themed 有主题的组件
type Themed interface {
    Node
    GetThemeStyle(styleKey string) StyleConfig
}

// ThemeHolder 主题持有者（嵌入组件）
type ThemeHolder struct {
    themeMgr *theme.Manager
    styleOverrides map[string]StyleConfig
}

// NewThemeHolder 创建主题持有者
func NewThemeHolder(themeMgr *theme.Manager) *ThemeHolder {
    return &ThemeHolder{
        themeMgr:      themeMgr,
        styleOverrides: make(map[string]StyleConfig),
    }
}

// GetStyle 获取样式
func (h *ThemeHolder) GetStyle(componentID, styleKey string) StyleConfig {
    // 1. 检查组件级覆盖
    if override, ok := h.styleOverrides[styleKey]; ok {
        return override
    }

    // 2. 从主题管理器获取
    return h.themeMgr.GetStyle(componentID, styleKey)
}

// OverrideStyle 覆盖样式
func (h *ThemeHolder) OverrideStyle(styleKey string, config StyleConfig) {
    h.styleOverrides[styleKey] = config
}

// GetColor 获取颜色
func (h *ThemeHolder) GetColor(colorKey string) theme.Color {
    return h.themeMgr.GetColor(colorKey)
}
```

### 示例：Themed Button

```go
// 位于: tui/framework/component/button.go

package component

// Button 按钮组件
type Button struct {
    BaseComponent
    *ThemeHolder

    label    string
    disabled bool
    focused  bool
}

// Paint 绘制按钮
func (b *Button) Paint(ctx PaintContext, buf *CellBuffer) {
    // 获取按钮样式
    var styleKey string
    if b.disabled {
        styleKey = "button.disabled"
    } else if b.focused {
        styleKey = "button.focused"
    } else {
        styleKey = "button.default"
    }

    styleConfig := b.GetStyle(b.ID(), styleKey)
    style := b.buildStyle(styleConfig)

    // 绘制按钮背景
    for i := 0; i < b.width; i++ {
        buf.SetCell(b.x+i, b.y, ' ', style)
    }

    // 绘制标签
    textX := b.x + (b.width - len(b.label)) / 2
    for i, ch := range b.label {
        buf.SetCell(textX+i, b.y, ch, style)
    }
}

// buildStyle 构建样式
func (b *Button) buildStyle(config StyleConfig) Style {
    style := lipgloss.NewStyle()

    if config.Foreground != nil {
        style = style.Foreground(config.Foreground.ToLipgloss())
    }
    if config.Background != nil {
        style = style.Background(config.Background.ToLipgloss())
    }
    if config.Bold {
        style = style.Bold(true)
    }

    return style
}
```

### 主题切换示例

```go
// 位于: tui/application/app.go

func (a *App) InitTheme() {
    themeMgr := theme.NewManager()

    // 注册内置主题
    themeMgr.Register(theme.LightTheme)
    themeMgr.Register(theme.DarkTheme)

    // 注册自定义主题
    customTheme := &theme.Theme{
        Name: "custom",
        Colors: theme.ColorPalette{
            Primary: theme.Color{RGB: &[3]uint8{255, 0, 128}},
        },
    }
    themeMgr.Register(customTheme)

    // 设置初始主题
    themeMgr.Set("light")

    a.themeMgr = themeMgr
}

// 切换主题
func (a *App) ToggleTheme() error {
    return a.themeMgr.Toggle()
}

// 订阅主题变化
func (a *App) OnThemeChange() {
    a.themeMgr.Subscribe(func(old, new *theme.Theme) {
        // 标记所有组件为 dirty，触发重新渲染
        a.dirty.MarkAll()
    })
}
```

## 主题配置

### YAML 配置格式

```yaml
# themes/default.yaml
name: default
version: "1.0.0"
extends: light  # 继承自 light 主题

colors:
  primary: "#4285f4"
  secondary: "#ec4899"
  success: "#4caf50"
  warning: "#ffc107"
  error: "#f44336"
  background: "#ffffff"
  foreground: "#1e293b"

fonts:
  default:
    family: "default"
    size: 14
  bold:
    family: "default"
    weight: bold

spacing:
  xs: 4
  sm: 8
  md: 16
  lg: 24
  xl: 32

styles:
  button.default:
    background: primary
    foreground: white
    padding: [8, 16, 8, 16]

  button.disabled:
    background: muted
    foreground: background

  button.focused:
    border: focus
    padding: [8, 16, 8, 16]

components:
  input:
    default:
      border: default
      padding: [8, 12, 8, 12]
    focused:
      border: focus
      padding: [8, 12, 8, 12]
```

### 加载配置

```go
// 位于: tui/framework/theme/loader.go

package theme

// LoadYAML 从 YAML 加载主题
func LoadYAML(path string) (*Theme, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    var config ThemeConfig
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, err
    }

    return BuildTheme(config)
}

// BuildTheme 从配置构建主题
func BuildTheme(config ThemeConfig) (*Theme, error) {
    theme := &Theme{
        Name:    config.Name,
        Version: config.Version,
        Colors:  buildColorPalette(config.Colors),
        Fonts:   buildFontSet(config.Fonts),
        Spacing: buildSpacingSet(config.Spacing),
        Styles:  buildStyles(config.Styles),
    }

    // 处理继承
    if config.Extends != "" {
        if parent, ok := defaultThemes[config.Extends]; ok {
            theme.Parent = parent
        }
    }

    return theme, nil
}
```

## 主题切换动画

### 过渡动画

```go
// 位于: tui/framework/theme/transition.go

package theme

// Transition 主题切换过渡
type Transition struct {
    from      *Theme
    to        *Theme
    progress  float64
    duration  time.Duration
    startedAt time.Time
}

// NewTransition 创建过渡
func NewTransition(from, to *Theme, duration time.Duration) *Transition {
    return &Transition{
        from:     from,
        to:       to,
        duration: duration,
        startedAt: time.Now(),
    }
}

// Update 更新过渡
func (t *Transition) Update(dt time.Duration) (done bool, progress float64) {
    t.progress = float64(time.Since(t.startedAt)) / float64(t.duration)

    if t.progress >= 1.0 {
        t.progress = 1.0
        return true, t.progress
    }

    return false, t.progress
}

// InterpolateColor 插值颜色
func (t *Transition) InterpolateColor(colorKey string) Color {
    fromColor := t.from.GetColor(colorKey)
    toColor := t.to.GetColor(colorKey)

    if fromColor.RGB != nil && toColor.RGB != nil {
        return Color{
            RGB: &[3]uint8{
                uint8(float64(fromColor.RGB[0]) + (float64(toColor.RGB[0])-float64(fromColor.RGB[0]))*t.progress),
                uint8(float64(fromColor.RGB[1]) + (float64(toColor.RGB[1])-float64(fromColor.RGB[1]))*t.progress),
                uint8(float64(fromColor.RGB[2]) + (float64(toColor.RGB[2])-float64(fromColor.RGB[2]))*t.progress),
            },
        }
    }

    return toColor
}

// GetStyle 获取过渡样式
func (t *Transition) GetStyle(componentID, styleKey string) StyleConfig {
    // 在过渡期间，使用插值样式
    return StyleConfig{}
}
```

## 测试

```go
// 位于: tui/framework/theme/theme_test.go

package theme

func TestThemeManager(t *testing.T) {
    mgr := NewManager()
    mgr.Register(LightTheme)
    mgr.Register(DarkTheme)

    // 设置初始主题
    err := mgr.Set("light")
    assert.NoError(t, err)
    assert.Equal(t, "light", mgr.Current().Name)

    // 切换主题
    err = mgr.Set("dark")
    assert.NoError(t, err)
    assert.Equal(t, "dark", mgr.Current().Name)

    // Toggle
    err = mgr.Toggle()
    assert.NoError(t, err)
    assert.Equal(t, "light", mgr.Current().Name)
}

func TestThemeInheritance(t *testing.T) {
    parent := &Theme{
        Name: "parent",
        Colors: ColorPalette{
            Primary: Color{RGB: &[3]uint8{100, 100, 100}},
        },
    }

    child := &Theme{
        Name:   "child",
        Parent: parent,
        Colors: ColorPalette{
            Secondary: Color{RGB: &[3]uint8{200, 200, 200}},
        },
    }

    mgr := NewManager()
    mgr.Register(child)

    // 可以获取父主题的颜色
    color := mgr.GetColor("primary")
    assert.Equal(t, uint8(100), color.RGB[0])
}

func TestStyleOverride(t *testing.T) {
    theme := LightTheme
    mgr := NewManager()
    mgr.Register(theme)

    // 获取原始样式
    original := mgr.GetStyle("button", "default")

    // 覆盖样式
    custom := StyleConfig{
        Background: &Color{RGB: &[3]uint8{255, 0, 0}},
    }
    // 设置组件级覆盖...

    // 验证覆盖生效
    // ...
}
```

## 总结

主题系统提供：

1. **解耦**: 组件不依赖具体样式库
2. **可切换**: 运行时切换，支持热切换
3. **可继承**: 支持主题继承和覆盖
4. **高性能**: 切换不重建组件树
5. **可扩展**: 支持自定义主题和配置

## 相关文档

- [ARCHITECTURE.md](./ARCHITECTURE.md) - 架构概览
- [COMPONENTS.md](./COMPONENTS.md) - 组件系统
- [RENDERING.md](./RENDERING.md) - 渲染系统
- [ARCHITECTURE_INVARIANTS.md](./ARCHITECTURE_INVARIANTS.md) - 架构不变量

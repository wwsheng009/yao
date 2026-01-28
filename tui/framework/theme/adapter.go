package theme

import (
	"fmt"

	"github.com/yaoapp/yao/tui/runtime/style"
	"github.com/yaoapp/yao/tui/framework/styling"
)

// =============================================================================
// 主题到样式提供者的适配器
// =============================================================================

// ThemeStyleProvider 主题样式提供者
// 将 theme.Manager 适配为 styling.StyleProvider 接口
// 这实现了适配器模式，让主题系统可以作为样式提供者使用
type ThemeStyleProvider struct {
	manager *Manager
}

// NewThemeStyleProvider 创建主题样式提供者
func NewThemeStyleProvider(manager *Manager) *ThemeStyleProvider {
	return &ThemeStyleProvider{
		manager: manager,
	}
}

// GetStyle 实现 styling.StyleProvider 接口
// 直接返回 style.Style，供组件渲染使用
func (p *ThemeStyleProvider) GetStyle(componentID, state string) style.Style {
	if p.manager == nil {
		return style.Style{}
	}

	// 从主题获取样式配置
	config := p.manager.GetStyle(componentID, state)

	// 转换为 style.Style（一次性转换，渲染层直接使用）
	return p.configToStyle(config)
}

// Name 实现 styling.ThemeProvider 接口
func (p *ThemeStyleProvider) Name() string {
	if p.manager != nil && p.manager.Current() != nil {
		return p.manager.Current().Name
	}
	return "unknown"
}

// SetTheme 实现 styling.ThemeProvider 接口
func (p *ThemeStyleProvider) SetTheme(name string) error {
	if p.manager == nil {
		return styling.ErrThemeNotFound
	}
	return p.manager.Set(name)
}

// CurrentTheme 实现 styling.ThemeProvider 接口
func (p *ThemeStyleProvider) CurrentTheme() string {
	if p.manager != nil && p.manager.Current() != nil {
		return p.manager.Current().Name
	}
	return "default"
}

// configToStyle 将 StyleConfig 转换为 style.Style
// 这是唯一需要转换的地方 - 从主题配置到渲染样式
func (p *ThemeStyleProvider) configToStyle(config StyleConfig) style.Style {
	s := style.Style{}

	if config.Foreground != nil {
		s.FG = p.colorToStyleColor(*config.Foreground)
	}
	if config.Background != nil {
		s.BG = p.colorToStyleColor(*config.Background)
	}
	if config.Bold {
		s = s.Bold(true)
	}
	if config.Italic {
		s = s.Italic(true)
	}
	if config.Underline {
		s = s.Underline(true)
	}
	if config.Strikethrough {
		s = s.Strikethrough(true)
	}
	if config.Reverse {
		s = s.Reverse(true)
	}
	if config.Blink {
		s = s.Blink(true)
	}

	return s
}

// colorToStyleColor 将 theme.Color 转换为 style.Color
func (p *ThemeStyleProvider) colorToStyleColor(c Color) style.Color {
	if c.IsNone() {
		return style.NoColor
	}

	switch c.Type {
	case ColorRGB, ColorHex:
		r, g, b := c.RGBValue()
		return p.rgbToNamedColor(r, g, b)

	case ColorNamed:
		if name, ok := c.Value.(string); ok {
			return p.namedColorToStyle(name)
		}

	case Color256:
		// 256色模式简化处理
		if code, ok := c.Value.(int); ok {
			return p.color256ToStyle(code)
		}
	}

	return style.NoColor
}

// rgbToNamedColor 将RGB转换为最接近的命名颜色
func (p *ThemeStyleProvider) rgbToNamedColor(r, g, b int) style.Color {
	// 计算亮度
	brightness := (r + g + b) / 3

	// 黑色/深色
	if brightness < 50 {
		return style.Black
	}

	// 白色/浅色
	if brightness > 230 {
		return style.White
	}

	// 改进的颜色匹配 - 按主导色分类
	maxVal := r
	if g > maxVal {
		maxVal = g
	}
	if b > maxVal {
		maxVal = b
	}
	minVal := r
	if g < minVal {
		minVal = g
	}
	if b < minVal {
		minVal = b
	}

	// 颜色范围
	rRange := maxVal - minVal
	isColorful := rRange > 50

	// 灰度/低饱和度
	if !isColorful || rRange < 30 {
		if brightness > 160 {
			return style.White // 亮灰
		}
		if brightness > 100 {
			return style.BrightBlack // 中灰
		}
		return style.Black // 暗灰
	}

	// 有颜色的部分
	if r >= g && r >= b {
		// 红色主导
		if r > 180 && g < 150 {
			return style.Red
		}
		if r > 200 && g > 150 {
			return style.Yellow // 偏黄的红色
		}
		if r > 200 && g < 100 && b > 150 {
			return style.Magenta // 偏紫的红色
		}
		return style.BrightRed
	}

	if g >= r && g >= b {
		// 绿色主导
		if g > 180 && b < 150 {
			return style.Green
		}
		if g > 180 && b > 150 {
			return style.Cyan // 青色（绿+蓝）
		}
		if r > 150 {
			return style.Yellow // 黄色（红+绿）
		}
		return style.BrightGreen
	}

	if b >= r && b >= g {
		// 蓝色主导
		if b > 200 && r < 150 && g < 150 {
			return style.Blue
		}
		if b > 180 && g > 150 {
			return style.BrightBlue // 亮蓝
		}
		if b > 180 && r > 150 {
			return style.Magenta // 品红（红+蓝）
		}
		if b > 150 && g > 150 {
			return style.Cyan // 青色（蓝+绿）
		}
		return style.BrightBlue
	}

	// 默认根据亮度返回
	if brightness > 160 {
		return style.White
	}
	if brightness > 80 {
		return style.BrightBlack
	}
	return style.Black
}

// namedColorToStyle 命名颜色转style.Color
func (p *ThemeStyleProvider) namedColorToStyle(name string) style.Color {
	// 直接使用 ColorCodes 映射
	if code, ok := ColorCodes[name]; ok {
		// 将ANSI代码转为style.Color
		return p.ansiCodeToColor(code)
	}
	return style.NoColor
}

// ansiCodeToColor ANSI代码转style.Color
func (p *ThemeStyleProvider) ansiCodeToColor(code int) style.Color {
	colors := map[int]style.Color{
		0:  style.Black,
		1:  style.Red,
		2:  style.Green,
		3:  style.Yellow,
		4:  style.Blue,
		5:  style.Magenta,
		6:  style.Cyan,
		7:  style.White,
		8:  style.BrightBlack,
		9:  style.BrightRed,
		10: style.BrightGreen,
		11: style.BrightYellow,
		12: style.BrightBlue,
		13: style.BrightMagenta,
		14: style.BrightCyan,
		15: style.BrightWhite,
	}
	if c, ok := colors[code]; ok {
		return c
	}
	return style.NoColor
}

// color256ToStyle 256色转style.Color
func (p *ThemeStyleProvider) color256ToStyle(code int) style.Color {
	// 简化处理：将256色映射到16色
	// 0-7: 标准色, 8-15: 亮色, 16-231: 216色, 232-255: 灰度
	switch {
	case code <= 7:
		return p.ansiCodeToColor(code)
	case code <= 15:
		return p.ansiCodeToColor(code)
	case code >= 232 && code <= 255:
		// 灰度，根据亮度返回
		if code < 244 {
			return style.BrightBlack
		}
		return style.White
	default:
		// 216色区域，简化处理
		return style.White
	}
}

// =============================================================================
// 全局注册辅助函数
// =============================================================================

// RegisterAsGlobal 将主题管理器注册为全局样式提供者
// 这是应用初始化时的便捷方法
// 同时注册到 styling 包和 style 包的桥接
func (m *Manager) RegisterAsGlobal() {
	provider := NewThemeStyleProvider(m)
	styling.SetProvider(provider)

	// 注册到 style 包的桥接，让 style.GetStyle() 可以工作
	// 包装为 func() StyleGetter 以匹配 style.RegisterStyleGetter 的签名
	style.RegisterStyleGetter(func() func(string, string) style.Style {
		return styling.GetStyleGetter()
	})
}

// SetTheme 设置主题并更新全局提供者
// 便捷方法：切换主题
func (m *Manager) SetThemeGlobal(name string) error {
	if err := m.Set(name); err != nil {
		return err
	}

	// 更新全局提供者（如果当前已注册）
	if current := styling.GetProvider(); current != nil {
		if tsp, ok := current.(*ThemeStyleProvider); ok && tsp.manager == m {
			styling.SetProvider(NewThemeStyleProvider(m))
		}
	}

	return nil
}

// =============================================================================
// 便捷主题切换入口
// =============================================================================

// InitThemes 初始化主题系统
// 创建管理器，注册内置主题，设置为全局提供者
// 注意：多次调用此函数是安全的，只有第一次会执行注册
func InitThemes(initialTheme string) (*Manager, error) {
	// 如果已经初始化过，直接返回默认管理器
	if styling.IsThemeInitialized() {
		mgr := NewManager()
		mgr.RegisterMultiple(BuiltinThemes())
		if initialTheme != "" {
			if err := mgr.Set(initialTheme); err != nil {
				return mgr, fmt.Errorf("failed to set initial theme: %w", err)
			}
		}
		return mgr, nil
	}

	// 首次初始化
	mgr := NewManager()
	mgr.RegisterMultiple(BuiltinThemes())

	if initialTheme != "" {
		if err := mgr.Set(initialTheme); err != nil {
			return mgr, fmt.Errorf("failed to set initial theme: %w", err)
		}
	}

	mgr.RegisterAsGlobal()
	styling.MarkThemeInitialized()
	return mgr, nil
}

package theme

import (
	"github.com/yaoapp/yao/tui/framework/styling"
)

// =============================================================================
// 主题到样式提供者的适配器
// =============================================================================

// ThemeProvider 主题样式提供者
// 将 theme.Manager 适配为 styling.StyleProvider 接口
// 这实现了适配器模式，让主题系统可以作为样式提供者使用
type ThemeProvider struct {
	manager *Manager
}

// NewThemeProvider 创建主题样式提供者
func NewThemeProvider(manager *Manager) *ThemeProvider {
	return &ThemeProvider{
		manager: manager,
	}
}

// GetStyle 实现 styling.StyleProvider 接口
func (p *ThemeProvider) GetStyle(componentID, state string) styling.StyleConfig {
	if p.manager == nil {
		return styling.StyleConfig{}
	}

	// 从主题获取样式配置
	config := p.manager.GetComponentStyle(componentID, state)

	// 转换为 styling.StyleConfig
	return p.convertToStylingConfig(config)
}

// convertToStylingConfig 将 theme.StyleConfig 转换为 styling.StyleConfig
func (p *ThemeProvider) convertToStylingConfig(config StyleConfig) styling.StyleConfig {
	result := styling.StyleConfig{}

	if config.Foreground != nil {
		result.Foreground = p.convertColor(*config.Foreground)
	}
	if config.Background != nil {
		result.Background = p.convertColor(*config.Background)
	}
	result.Bold = config.Bold
	result.Italic = config.Italic
	result.Underline = config.Underline
	result.Strikethrough = config.Strikethrough
	result.Reverse = config.Reverse
	result.Blink = config.Blink

	return result
}

// convertColor 转换颜色
func (p *ThemeProvider) convertColor(c Color) *styling.Color {
	if c.IsNone() {
		return nil
	}

	sc := &styling.Color{
		Type: styling.ColorType(c.Type),
	}

	// 根据 color 类型提取值
	switch c.Type {
	case ColorRGB, ColorHex:
		if rgb, ok := c.Value.([3]int); ok {
			sc.Value = rgb
		}
	case Color256:
		if code, ok := c.Value.(int); ok {
			// 将 256 色码转换为 RGB 简化处理
			// 实际应用中可能需要更复杂的转换
			sc.Value = [3]int{code, code, code}
		}
	case ColorNamed:
		if name, ok := c.Value.(string); ok {
			// 命名颜色转为预设的 RGB 值
			sc.Value = namedColorToRGB(name)
		}
	}

	return sc
}

// namedColorToRGB 将命名颜色转换为 RGB 值
func namedColorToRGB(name string) [3]int {
	// 基本颜色映射
	colors := map[string][3]int{
		"black":         {0, 0, 0},
		"red":           {255, 0, 0},
		"green":         {0, 255, 0},
		"yellow":        {255, 255, 0},
		"blue":          {0, 0, 255},
		"magenta":       {255, 0, 255},
		"cyan":          {0, 255, 255},
		"white":         {255, 255, 255},
		"bright-black":   {128, 128, 128},
		"bright-red":     {255, 128, 128},
		"bright-green":   {128, 255, 128},
		"bright-yellow":  {255, 255, 128},
		"bright-blue":    {128, 128, 255},
		"bright-magenta": {255, 128, 255},
		"bright-cyan":    {128, 255, 255},
		"bright-white":   {255, 255, 255},
	}
	if rgb, ok := colors[name]; ok {
		return rgb
	}
	return [3]int{255, 255, 255}
}

// =============================================================================
// 全局注册辅助函数
// =============================================================================

// RegisterAsGlobal 将主题管理器注册为全局样式提供者
// 这是应用初始化时的便捷方法
func (m *Manager) RegisterAsGlobal() {
	provider := NewThemeProvider(m)
	styling.RegisterGlobalProvider(provider)
}

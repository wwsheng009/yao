package runtime

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/yaoapp/yao/tui/framework/theme"
)

// VisualStyle extends the base Style with visual properties (colors, borders, etc.)
// This bridges the Runtime Style system with lipgloss for rich visual styling.
type VisualStyle struct {
	// Base style provides layout properties
	Style

	// Colors
	Foreground string
	Background string

	// Border styling
	Border     lipgloss.Border
	BorderForeground string
	BorderBackground string

	// Text styling
	Bold      bool
	Italic    bool
	Underline bool
	Strikethrough bool

	// Alignment
	Align     lipgloss.Position
	MarginTop    int
	MarginRight  int
	MarginBottom int
	MarginLeft   int

	// Width/Height constraints (can use lipgloss for these)
	MaxWidth  int
	MaxHeight int

	// Whether border is enabled
	HasBorder bool
}

// NewVisualStyle creates a default VisualStyle
func NewVisualStyle() VisualStyle {
	return VisualStyle{
		Style:         NewStyle(),
		Foreground:    "",
		Background:    "",
		Border:        lipgloss.Border{},
		BorderForeground: "",
		BorderBackground: "",
		Bold:          false,
		Italic:        false,
		Underline:     false,
		Strikethrough: false,
		Align:         lipgloss.Position(0), // Left
		MaxWidth:      0,
		MaxHeight:     0,
		HasBorder:     false,
	}
}

// ToLipgloss converts VisualStyle to a lipgloss.Style
func (vs VisualStyle) ToLipgloss() lipgloss.Style {
	style := lipgloss.NewStyle()

	// Apply colors
	if vs.Foreground != "" {
		style = style.Foreground(lipgloss.Color(vs.Foreground))
	}
	if vs.Background != "" {
		style = style.Background(lipgloss.Color(vs.Background))
	}

	// Apply border
	if vs.HasBorder {
		style = style.Border(vs.Border)
	}
	if vs.BorderForeground != "" {
		style = style.BorderForeground(lipgloss.Color(vs.BorderForeground))
	}
	if vs.BorderBackground != "" {
		style = style.BorderBackground(lipgloss.Color(vs.BorderBackground))
	}

	// Apply text styling
	if vs.Bold {
		style = style.Bold(true)
	}
	if vs.Italic {
		style = style.Italic(true)
	}
	if vs.Underline {
		style = style.Underline(true)
	}
	if vs.Strikethrough {
		style = style.Strikethrough(true)
	}

	// Apply alignment
	style = style.Align(vs.Align)

	// Apply margins
	if vs.MarginTop > 0 || vs.MarginRight > 0 || vs.MarginBottom > 0 || vs.MarginLeft > 0 {
		style = style.Margin(
			vs.MarginTop,
			vs.MarginRight,
			vs.MarginBottom,
			vs.MarginLeft,
		)
	}

	// Apply padding from base style
	if vs.Style.Padding.Top > 0 || vs.Style.Padding.Right > 0 ||
	   vs.Style.Padding.Bottom > 0 || vs.Style.Padding.Left > 0 {
		style = style.Padding(
			vs.Style.Padding.Top,
			vs.Style.Padding.Right,
			vs.Style.Padding.Bottom,
			vs.Style.Padding.Left,
		)
	}

	// Apply width/height constraints
	if vs.MaxWidth > 0 {
		style = style.MaxWidth(vs.MaxWidth)
	}
	if vs.MaxHeight > 0 {
		style = style.MaxHeight(vs.MaxHeight)
	}

	// Apply explicit width/height if set
	if vs.Style.Width > 0 && vs.Style.Width != AutoSize {
		style = style.Width(vs.Style.Width)
	}
	if vs.Style.Height > 0 && vs.Style.Height != AutoSize {
		style = style.Height(vs.Style.Height)
	}

	return style
}

// WithForeground sets the foreground color
func (vs VisualStyle) WithForeground(color string) VisualStyle {
	vs.Foreground = color
	return vs
}

// WithBackground sets the background color
func (vs VisualStyle) WithBackground(color string) VisualStyle {
	vs.Background = color
	return vs
}

// WithBorder sets the border style
func (vs VisualStyle) WithBorder(border lipgloss.Border) VisualStyle {
	vs.Border = border
	vs.HasBorder = true
	return vs
}

// WithBorderType sets a border by type (normal, rounded, thick, double)
func (vs VisualStyle) WithBorderType(borderType string) VisualStyle {
	vs.HasBorder = true
	switch borderType {
	case "normal":
		vs.Border = lipgloss.NormalBorder()
	case "rounded":
		vs.Border = lipgloss.RoundedBorder()
	case "thick":
		vs.Border = lipgloss.ThickBorder()
	case "double":
		vs.Border = lipgloss.DoubleBorder()
	case "hidden":
		vs.Border = lipgloss.HiddenBorder()
	default:
		vs.Border = lipgloss.NormalBorder()
	}
	return vs
}

// WithBorderForeground sets the border foreground color
func (vs VisualStyle) WithBorderForeground(color string) VisualStyle {
	vs.BorderForeground = color
	return vs
}

// WithBorderBackground sets the border background color
func (vs VisualStyle) WithBorderBackground(color string) VisualStyle {
	vs.BorderBackground = color
	return vs
}

// WithBold enables bold text
func (vs VisualStyle) WithBold(enabled bool) VisualStyle {
	vs.Bold = enabled
	return vs
}

// WithItalic enables italic text
func (vs VisualStyle) WithItalic(enabled bool) VisualStyle {
	vs.Italic = enabled
	return vs
}

// WithUnderline enables underline
func (vs VisualStyle) WithUnderline(enabled bool) VisualStyle {
	vs.Underline = enabled
	return vs
}

// WithStrikethrough enables strikethrough
func (vs VisualStyle) WithStrikethrough(enabled bool) VisualStyle {
	vs.Strikethrough = enabled
	return vs
}

// WithAlign sets text alignment
func (vs VisualStyle) WithAlign(align string) VisualStyle {
	switch align {
	case "left":
		vs.Align = lipgloss.Left
	case "right":
		vs.Align = lipgloss.Right
	case "center":
		vs.Align = lipgloss.Center
	default:
		vs.Align = lipgloss.Left
	}
	return vs
}

// WithMargin sets margin on all sides
func (vs VisualStyle) WithMargin(top, right, bottom, left int) VisualStyle {
	vs.MarginTop = top
	vs.MarginRight = right
	vs.MarginBottom = bottom
	vs.MarginLeft = left
	return vs
}

// WithMarginTop sets top margin
func (vs VisualStyle) WithMarginTop(margin int) VisualStyle {
	vs.MarginTop = margin
	return vs
}

// WithMarginRight sets right margin
func (vs VisualStyle) WithMarginRight(margin int) VisualStyle {
	vs.MarginRight = margin
	return vs
}

// WithMarginBottom sets bottom margin
func (vs VisualStyle) WithMarginBottom(margin int) VisualStyle {
	vs.MarginBottom = margin
	return vs
}

// WithMarginLeft sets left margin
func (vs VisualStyle) WithMarginLeft(margin int) VisualStyle {
	vs.MarginLeft = margin
	return vs
}

// WithMaxWidth sets maximum width
func (vs VisualStyle) WithMaxWidth(width int) VisualStyle {
	vs.MaxWidth = width
	return vs
}

// WithMaxHeight sets maximum height
func (vs VisualStyle) WithMaxHeight(height int) VisualStyle {
	vs.MaxHeight = height
	return vs
}

// Render applies the visual style to content
func (vs VisualStyle) Render(content string) string {
	return vs.ToLipgloss().Render(content)
}

// ApplyBorderStyle is a helper to apply common border styles
func ApplyBorderStyle(base VisualStyle, borderStyle string) VisualStyle {
	switch borderStyle {
	case "normal":
		return base.WithBorder(lipgloss.NormalBorder())
	case "rounded":
		return base.WithBorder(lipgloss.RoundedBorder())
	case "thick":
		return base.WithBorder(lipgloss.ThickBorder())
	case "double":
		return base.WithBorder(lipgloss.DoubleBorder())
	default:
		return base
	}
}

// CommonColorPalettes provides predefined color schemes
var CommonColorPalettes = map[string]struct {
	Primary, Secondary, Accent, Muted, Success, Warning, Error string
}{
	"default": {
		Primary:   "#007AFF",
		Secondary: "#5856D6",
		Accent:    "#FF9500",
		Muted:     "#8E8E93",
		Success:   "#34C759",
		Warning:   "#FFCC00",
		Error:     "#FF3B30",
	},
	"dracula": {
		Primary:   "#BD93F9",
		Secondary: "#FF79C6",
		Accent:    "#50FA7B",
		Muted:     "#6272A4",
		Success:   "#50FA7B",
		Warning:   "#F1FA8C",
		Error:     "#FF5555",
	},
	"nord": {
		Primary:   "#88C0D0",
		Secondary: "#81A1C1",
		Accent:    "#8FBCBB",
		Muted:     "#4C566A",
		Success:   "#A3BE8C",
		Warning:   "#EBCB8B",
		Error:     "#BF616A",
	},
	"monokai": {
		Primary:   "#A6E22E",
		Secondary: "#66D9EF",
		Accent:    "#F92672",
		Muted:     "#75715E",
		Success:   "#A6E22E",
		Warning:   "#E6DB74",
		Error:     "#F92672",
	},
}

// GetColorPalette returns a color palette by name
func GetColorPalette(name string) (colors struct {
	Primary, Secondary, Accent, Muted, Success, Warning, Error string
}) {
	if palette, ok := CommonColorPalettes[name]; ok {
		return palette
	}
	return CommonColorPalettes["default"]
}

// =============================================================================
// 与 theme 系统的集成
// =============================================================================

// FromTheme 从 theme.Theme 创建 VisualStyle
func FromTheme(t *theme.Theme) VisualStyle {
	if t == nil {
		return NewVisualStyle()
	}

	vs := NewVisualStyle()

	// 从主题获取颜色
	vs.Foreground = t.Colors.Foreground.String()
	vs.Background = t.Colors.Background.String()
	vs.BorderForeground = t.Colors.Border.String()

	return vs
}

// FromThemeWithStyle 从 theme.StyleConfig 创建 VisualStyle
func FromThemeWithStyle(config theme.StyleConfig) VisualStyle {
	vs := NewVisualStyle()

	if config.Foreground != nil {
		vs.Foreground = config.Foreground.String()
	}
	if config.Background != nil {
		vs.Background = config.Background.String()
	}
	if config.Bold {
		vs.Bold = true
	}
	if config.Italic {
		vs.Italic = true
	}
	if config.Underline {
		vs.Underline = true
	}
	if config.Strikethrough {
		vs.Strikethrough = true
	}
	if config.Border != nil {
		vs.HasBorder = config.Border.Enabled
		vs.BorderForeground = config.Border.FG.String()
		vs.BorderBackground = config.Border.BG.String()
	}

	return vs
}

// WithThemeColor 使用主题颜色设置前景色
func (vs VisualStyle) WithThemeColor(color theme.Color) VisualStyle {
	vs.Foreground = color.String()
	return vs
}

// WithThemeBackground 使用主题颜色设置背景色
func (vs VisualStyle) WithThemeBackground(color theme.Color) VisualStyle {
	vs.Background = color.String()
	return vs
}

// WithThemePalette 使用主题调色板设置颜色
func (vs VisualStyle) WithThemePalette(palette theme.ColorPalette, colorName string) VisualStyle {
	var color theme.Color
	switch colorName {
	case "primary":
		color = palette.Primary
	case "secondary":
		color = palette.Secondary
	case "accent":
		color = palette.Accent
	case "success":
		color = palette.Success
	case "warning":
		color = palette.Warning
	case "error":
		color = palette.Error
	case "info":
		color = palette.Info
	case "foreground", "fg":
		color = palette.Foreground
	case "background", "bg":
		color = palette.Background
	case "muted":
		color = palette.Muted
	case "border":
		color = palette.Border
	case "focus":
		color = palette.Focus
	default:
		return vs
	}
	vs.Foreground = color.String()
	return vs
}

// WithBuiltinTheme 使用内置主题创建 VisualStyle
func WithBuiltinTheme(themeName string) VisualStyle {
	t := theme.GetBuiltinTheme(themeName)
	return FromTheme(t)
}

// ThemeVisualStyle 主题驱动的 VisualStyle
// 当主题改变时自动更新
type ThemeVisualStyle struct {
	themeMgr *theme.Manager
	styleKey string
	componentID string
	state string
}

// NewThemeVisualStyle 创建主题驱动的 VisualStyle
func NewThemeVisualStyle(manager *theme.Manager, componentID, styleKey, state string) *ThemeVisualStyle {
	return &ThemeVisualStyle{
		themeMgr:    manager,
		componentID: componentID,
		styleKey:    styleKey,
		state:       state,
	}
}

// ToVisualStyle 转换为 VisualStyle
func (tvs *ThemeVisualStyle) ToVisualStyle() VisualStyle {
	if tvs.themeMgr == nil {
		return NewVisualStyle()
	}

	var config theme.StyleConfig
	if tvs.componentID != "" {
		config = tvs.themeMgr.GetComponentStyle(tvs.componentID, tvs.state)
	} else if tvs.styleKey != "" {
		config = tvs.themeMgr.GetStyle("", tvs.styleKey)
	}

	vs := FromThemeWithStyle(config)

	// 如果没有配置样式，使用当前主题的基础颜色
	if vs.Foreground == "" {
		vs.Foreground = tvs.themeMgr.GetColor("foreground").String()
	}
	if vs.Background == "" {
		vs.Background = tvs.themeMgr.GetColor("background").String()
	}

	return vs
}

// SetManager 设置主题管理器
func (tvs *ThemeVisualStyle) SetManager(manager *theme.Manager) {
	tvs.themeMgr = manager
}

// UpdateState 更新状态
func (tvs *ThemeVisualStyle) UpdateState(state string) {
	tvs.state = state
}


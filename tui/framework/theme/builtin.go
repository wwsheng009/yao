package theme

// LightTheme 亮色主题
var LightTheme = &Theme{
	Name:    "light",
	Version: "1.0.0",
	Colors: ColorPalette{
		Primary:   Color{Type: ColorRGB, Value: [3]int{66, 133, 244}},   // 蓝色
		Secondary: Color{Type: ColorRGB, Value: [3]int{236, 72, 153}},    // 紫色
		Accent:    Color{Type: ColorRGB, Value: [3]int{255, 193, 7}},     // 橙色
		Success:   Color{Type: ColorRGB, Value: [3]int{76, 175, 80}},     // 绿色
		Warning:   Color{Type: ColorRGB, Value: [3]int{255, 193, 7}},     // 橙色
		Error:     Color{Type: ColorRGB, Value: [3]int{244, 67, 54}},      // 红色
		Info:      Color{Type: ColorRGB, Value: [3]int{66, 165, 245}},     // 青色
		Background: Color{Type: ColorRGB, Value: [3]int{255, 255, 255}},  // 白色
		Foreground: Color{Type: ColorRGB, Value: [3]int{30, 41, 59}},      // 深灰
		Muted:      Color{Type: ColorRGB, Value: [3]int{148, 163, 184}},   // 灰色
		Border:     Color{Type: ColorRGB, Value: [3]int{227, 233, 240}},   // 浅灰
		Focus:      Color{Type: ColorRGB, Value: [3]int{66, 133, 244}},    // 蓝色
		Disabled:   Color{Type: ColorRGB, Value: [3]int{201, 203, 207}},   // 灰色
		Hover:      Color{Type: ColorRGB, Value: [3]int{66, 133, 244}},    // 蓝色
		Active:     Color{Type: ColorRGB, Value: [3]int{66, 133, 244}},    // 蓝色
	},
	Styles: map[string]StyleConfig{
		"text.primary": {
			Foreground: &Color{Type: ColorRGB, Value: [3]int{30, 41, 59}},
		},
		"text.secondary": {
			Foreground: &Color{Type: ColorRGB, Value: [3]int{148, 163, 184}},
		},
		"border.default": {
			Foreground: &Color{Type: ColorRGB, Value: [3]int{227, 233, 240}},
		},
	},
	Components: map[string]ComponentStyle{
		"input": {
			Base: StyleConfig{
				Foreground: &Color{Type: ColorRGB, Value: [3]int{30, 41, 59}},
			},
			States: map[string]StyleConfig{
				"focus": {
					Foreground: &Color{Type: ColorRGB, Value: [3]int{66, 133, 244}},
				},
				"placeholder": {
					Foreground: &Color{Type: ColorRGB, Value: [3]int{148, 163, 184}},
				},
			},
		},
	},
	Spacing: DefaultSpacingSet(),
}

// DarkTheme 深色主题
var DarkTheme = &Theme{
	Name:    "dark",
	Version: "1.0.0",
	Colors: ColorPalette{
		Primary:    Color{Type: ColorRGB, Value: [3]int{97, 175, 239}},   // 亮蓝
		Secondary:  Color{Type: ColorRGB, Value: [3]int{224, 108, 117}},  // 粉红
		Accent:     Color{Type: ColorRGB, Value: [3]int{255, 213, 79}},   // 金色
		Success:    Color{Type: ColorRGB, Value: [3]int{134, 239, 172}},  // 绿色
		Warning:    Color{Type: ColorRGB, Value: [3]int{255, 213, 79}},   // 金色
		Error:      Color{Type: ColorRGB, Value: [3]int{239, 68, 68}},     // 红色
		Info:       Color{Type: ColorRGB, Value: [3]int{66, 165, 245}},   // 青色
		Background: Color{Type: ColorRGB, Value: [3]int{17, 24, 39}},     // 深蓝黑
		Foreground: Color{Type: ColorRGB, Value: [3]int{227, 233, 240}},  // 浅灰
		Muted:      Color{Type: ColorRGB, Value: [3]int{161, 161, 170}},  // 灰色
		Border:     Color{Type: ColorRGB, Value: [3]int{55, 65, 81}},     // 深灰
		Focus:      Color{Type: ColorRGB, Value: [3]int{97, 175, 239}},   // 亮蓝
		Disabled:   Color{Type: ColorRGB, Value: [3]int{86, 95, 105}},    // 暗灰
		Hover:      Color{Type: ColorRGB, Value: [3]int{97, 175, 239}},   // 亮蓝
		Active:     Color{Type: ColorRGB, Value: [3]int{97, 175, 239}},   // 亮蓝
	},
	Styles: map[string]StyleConfig{
		"text.primary": {
			Foreground: &Color{Type: ColorRGB, Value: [3]int{227, 233, 240}},
		},
		"text.secondary": {
			Foreground: &Color{Type: ColorRGB, Value: [3]int{161, 161, 170}},
		},
		"border.default": {
			Foreground: &Color{Type: ColorRGB, Value: [3]int{55, 65, 81}},
		},
	},
	Components: map[string]ComponentStyle{
		"input": {
			Base: StyleConfig{
				Foreground: &Color{Type: ColorRGB, Value: [3]int{227, 233, 240}},
			},
			States: map[string]StyleConfig{
				"focus": {
					Foreground: &Color{Type: ColorRGB, Value: [3]int{97, 175, 239}},
				},
				"placeholder": {
					Foreground: &Color{Type: ColorRGB, Value: [3]int{161, 161, 170}},
				},
			},
		},
	},
	Spacing: DefaultSpacingSet(),
}

// DraculaTheme Dracula 主题
var DraculaTheme = &Theme{
	Name:    "dracula",
	Version: "1.0.0",
	Colors: ColorPalette{
		Primary:    Color{Type: ColorRGB, Value: [3]int{189, 147, 249}},  // 紫色
		Secondary:  Color{Type: ColorRGB, Value: [3]int{255, 121, 198}},  // 粉色
		Accent:     Color{Type: ColorRGB, Value: [3]int{80, 250, 123}},   // 绿色
		Success:    Color{Type: ColorRGB, Value: [3]int{80, 250, 123}},   // 绿色
		Warning:    Color{Type: ColorRGB, Value: [3]int{241, 250, 140}},  // 黄色
		Error:      Color{Type: ColorRGB, Value: [3]int{255, 85, 85}},    // 红色
		Info:       Color{Type: ColorRGB, Value: [3]int{139, 233, 253}},  // 青色
		Background: Color{Type: ColorRGB, Value: [3]int{40, 42, 54}},    // 深色背景
		Foreground: Color{Type: ColorRGB, Value: [3]int{248, 248, 242}},  // 白色
		Muted:      Color{Type: ColorRGB, Value: [3]int{98, 114, 164}},   // 紫灰
		Border:     Color{Type: ColorRGB, Value: [3]int{98, 114, 164}},   // 紫灰
		Focus:      Color{Type: ColorRGB, Value: [3]int{189, 147, 249}},  // 紫色
		Disabled:   Color{Type: ColorRGB, Value: [3]int{98, 114, 164}},   // 紫灰
		Hover:      Color{Type: ColorRGB, Value: [3]int{255, 121, 198}},  // 粉色
		Active:     Color{Type: ColorRGB, Value: [3]int{189, 147, 249}},  // 紫色
	},
	Styles: map[string]StyleConfig{
		"text.primary": {
			Foreground: &Color{Type: ColorRGB, Value: [3]int{248, 248, 242}},
		},
		"text.secondary": {
			Foreground: &Color{Type: ColorRGB, Value: [3]int{98, 114, 164}},
		},
		"border.default": {
			Foreground: &Color{Type: ColorRGB, Value: [3]int{98, 114, 164}},
		},
		"button.primary": {
			Foreground: &Color{Type: ColorRGB, Value: [3]int{40, 42, 54}},
			Background: &Color{Type: ColorRGB, Value: [3]int{189, 147, 249}},
			Bold:       true,
		},
	},
	Components: map[string]ComponentStyle{
		"input": {
			Base: StyleConfig{
				Foreground: &Color{Type: ColorRGB, Value: [3]int{248, 248, 242}},
			},
			States: map[string]StyleConfig{
				"focus": {
					Foreground: &Color{Type: ColorRGB, Value: [3]int{189, 147, 249}},
				},
				"placeholder": {
					Foreground: &Color{Type: ColorRGB, Value: [3]int{98, 114, 164}},
				},
			},
		},
	},
	Spacing: DefaultSpacingSet(),
}

// NordTheme Nord 主题
var NordTheme = &Theme{
	Name:    "nord",
	Version: "1.0.0",
	Colors: ColorPalette{
		Primary:    Color{Type: ColorRGB, Value: [3]int{136, 192, 208}}, // 冰蓝
		Secondary:  Color{Type: ColorRGB, Value: [3]int{129, 161, 193}}, // 天蓝
		Accent:     Color{Type: ColorRGB, Value: [3]int{143, 188, 187}}, // 青绿
		Success:    Color{Type: ColorRGB, Value: [3]int{163, 190, 140}}, // 绿色
		Warning:    Color{Type: ColorRGB, Value: [3]int{235, 203, 139}}, // 黄色
		Error:      Color{Type: ColorRGB, Value: [3]int{191, 97, 106}},  // 红色
		Info:       Color{Type: ColorRGB, Value: [3]int{94, 129, 172}},  // 深蓝
		Background: Color{Type: ColorRGB, Value: [3]int{46, 52, 64}},   // 深灰蓝
		Foreground: Color{Type: ColorRGB, Value: [3]int{236, 239, 244}}, // 浅灰白
		Muted:      Color{Type: ColorRGB, Value: [3]int{76, 86, 106}},   // 灰蓝
		Border:     Color{Type: ColorRGB, Value: [3]int{59, 66, 82}},    // 深灰蓝
		Focus:      Color{Type: ColorRGB, Value: [3]int{136, 192, 208}}, // 冰蓝
		Disabled:   Color{Type: ColorRGB, Value: [3]int{59, 66, 82}},    // 深灰蓝
		Hover:      Color{Type: ColorRGB, Value: [3]int{129, 161, 193}}, // 天蓝
		Active:     Color{Type: ColorRGB, Value: [3]int{136, 192, 208}}, // 冰蓝
	},
	Styles: map[string]StyleConfig{
		"text.primary": {
			Foreground: &Color{Type: ColorRGB, Value: [3]int{236, 239, 244}},
		},
		"text.secondary": {
			Foreground: &Color{Type: ColorRGB, Value: [3]int{216, 222, 233}},
		},
		"border.default": {
			Foreground: &Color{Type: ColorRGB, Value: [3]int{59, 66, 82}},
		},
	},
	Components: map[string]ComponentStyle{
		"input": {
			Base: StyleConfig{
				Foreground: &Color{Type: ColorRGB, Value: [3]int{236, 239, 244}},
			},
			States: map[string]StyleConfig{
				"focus": {
					Foreground: &Color{Type: ColorRGB, Value: [3]int{136, 192, 208}},
				},
				"placeholder": {
					Foreground: &Color{Type: ColorRGB, Value: [3]int{216, 222, 233}},
				},
			},
		},
	},
	Spacing: DefaultSpacingSet(),
}

// MonokaiTheme Monokai 主题
var MonokaiTheme = &Theme{
	Name:    "monokai",
	Version: "1.0.0",
	Colors: ColorPalette{
		Primary:    Color{Type: ColorRGB, Value: [3]int{166, 226, 46}},   // 绿色
		Secondary:  Color{Type: ColorRGB, Value: [3]int{102, 217, 239}},  // 青色
		Accent:     Color{Type: ColorRGB, Value: [3]int{249, 38, 114}},    // 粉红
		Success:    Color{Type: ColorRGB, Value: [3]int{166, 226, 46}},   // 绿色
		Warning:    Color{Type: ColorRGB, Value: [3]int{230, 219, 116}},  // 黄色
		Error:      Color{Type: ColorRGB, Value: [3]int{249, 38, 114}},    // 粉红
		Info:       Color{Type: ColorRGB, Value: [3]int{102, 217, 239}},  // 青色
		Background: Color{Type: ColorRGB, Value: [3]int{39, 40, 34}},     // 深灰
		Foreground: Color{Type: ColorRGB, Value: [3]int{248, 248, 242}},  // 白色
		Muted:      Color{Type: ColorRGB, Value: [3]int{117, 113, 94}},   // 灰色
		Border:     Color{Type: ColorRGB, Value: [3]int{73, 72, 62}},     // 深灰
		Focus:      Color{Type: ColorRGB, Value: [3]int{166, 226, 46}},   // 绿色
		Disabled:   Color{Type: ColorRGB, Value: [3]int{73, 72, 62}},     // 深灰
		Hover:      Color{Type: ColorRGB, Value: [3]int{102, 217, 239}},  // 青色
		Active:     Color{Type: ColorRGB, Value: [3]int{166, 226, 46}},   // 绿色
	},
	Styles: map[string]StyleConfig{
		"text.primary": {
			Foreground: &Color{Type: ColorRGB, Value: [3]int{248, 248, 242}},
		},
		"text.secondary": {
			Foreground: &Color{Type: ColorRGB, Value: [3]int{117, 113, 94}},
		},
		"border.default": {
			Foreground: &Color{Type: ColorRGB, Value: [3]int{73, 72, 62}},
		},
		"code.keyword": {
			Foreground: &Color{Type: ColorRGB, Value: [3]int{249, 38, 114}},
		},
		"code.string": {
			Foreground: &Color{Type: ColorRGB, Value: [3]int{230, 219, 116}},
		},
		"code.comment": {
			Foreground: &Color{Type: ColorRGB, Value: [3]int{117, 113, 94}},
			Italic:     true,
		},
	},
	Components: map[string]ComponentStyle{
		"input": {
			Base: StyleConfig{
				Foreground: &Color{Type: ColorRGB, Value: [3]int{248, 248, 242}},
			},
			States: map[string]StyleConfig{
				"focus": {
					Foreground: &Color{Type: ColorRGB, Value: [3]int{166, 226, 46}},
				},
				"placeholder": {
					Foreground: &Color{Type: ColorRGB, Value: [3]int{117, 113, 94}},
				},
			},
		},
	},
	Spacing: DefaultSpacingSet(),
}

// TokyoNightTheme Tokyo Night 主题
var TokyoNightTheme = &Theme{
	Name:    "tokyo-night",
	Version: "1.0.0",
	Colors: ColorPalette{
		Primary:    Color{Type: ColorRGB, Value: [3]int{39, 105, 187}},   // 蓝色
		Secondary:  Color{Type: ColorRGB, Value: [3]int{224, 107, 229}},  // 紫色
		Accent:     Color{Type: ColorRGB, Value: [3]int{235, 135, 95}},   // 橙色
		Success:    Color{Type: ColorRGB, Value: [3]int{35, 174, 148}},   // 绿色
		Warning:    Color{Type: ColorRGB, Value: [3]int{235, 135, 95}},   // 橙色
		Error:      Color{Type: ColorRGB, Value: [3]int{227, 103, 115}},  // 红色
		Info:       Color{Type: ColorRGB, Value: [3]int{39, 105, 187}},   // 蓝色
		Background: Color{Type: ColorRGB, Value: [3]int{26, 27, 38}},     // 深蓝黑
		Foreground: Color{Type: ColorRGB, Value: [3]int{169, 177, 214}},  // 浅灰
		Muted:      Color{Type: ColorRGB, Value: [3]int{113, 124, 180}},  // 紫灰
		Border:     Color{Type: ColorRGB, Value: [3]int{77, 87, 114}},    // 灰蓝
		Focus:      Color{Type: ColorRGB, Value: [3]int{39, 105, 187}},   // 蓝色
		Disabled:   Color{Type: ColorRGB, Value: [3]int{77, 87, 114}},    // 灰蓝
		Hover:      Color{Type: ColorRGB, Value: [3]int{187, 154, 247}},  // 紫色
		Active:     Color{Type: ColorRGB, Value: [3]int{39, 105, 187}},   // 蓝色
	},
	Styles: map[string]StyleConfig{
		"text.primary": {
			Foreground: &Color{Type: ColorRGB, Value: [3]int{169, 177, 214}},
		},
		"text.secondary": {
			Foreground: &Color{Type: ColorRGB, Value: [3]int{113, 124, 180}},
		},
		"border.default": {
			Foreground: &Color{Type: ColorRGB, Value: [3]int{77, 87, 114}},
		},
	},
	Components: map[string]ComponentStyle{
		"input": {
			Base: StyleConfig{
				Foreground: &Color{Type: ColorRGB, Value: [3]int{169, 177, 214}},
			},
			States: map[string]StyleConfig{
				"focus": {
					Foreground: &Color{Type: ColorRGB, Value: [3]int{39, 105, 187}},
				},
				"placeholder": {
					Foreground: &Color{Type: ColorRGB, Value: [3]int{113, 124, 180}},
				},
			},
		},
	},
	Spacing: DefaultSpacingSet(),
}

// GetBuiltinTheme 获取内置主题
func GetBuiltinTheme(name string) *Theme {
	switch name {
	case "light":
		return LightTheme
	case "dark":
		return DarkTheme
	case "dracula":
		return DraculaTheme
	case "nord":
		return NordTheme
	case "monokai":
		return MonokaiTheme
	case "tokyo-night":
		return TokyoNightTheme
	default:
		return DarkTheme // 默认返回深色主题
	}
}

// BuiltinThemes 返回所有内置主题
func BuiltinThemes() []*Theme {
	return []*Theme{
		LightTheme,
		DarkTheme,
		DraculaTheme,
		NordTheme,
		MonokaiTheme,
		TokyoNightTheme,
	}
}

// BuiltinThemeNames 返回所有内置主题名称
func BuiltinThemeNames() []string {
	return []string{
		"light",
		"dark",
		"dracula",
		"nord",
		"monokai",
		"tokyo-night",
	}
}

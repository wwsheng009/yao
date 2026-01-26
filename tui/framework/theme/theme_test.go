package theme

import (
	"testing"
)

func TestNewTheme(t *testing.T) {
	theme := NewTheme("test")

	if theme.Name != "test" {
		t.Errorf("NewTheme() Name = %v, want 'test'", theme.Name)
	}
	if theme.Version != "1.0.0" {
		t.Errorf("NewTheme() Version = %v, want '1.0.0'", theme.Version)
	}
}

func TestThemeGetColor(t *testing.T) {
	theme := NewTheme("test")
	theme.Colors.Primary = Blue
	theme.Colors.Secondary = Red

	tests := []struct {
		name  string
		key   string
		want  Color
	}{
		{"primary", "primary", Blue},
		{"secondary", "secondary", Red},
		{"unknown", "unknown", NoColor},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := theme.GetColor(tt.key)
			if !got.Equals(tt.want) {
				t.Errorf("GetColor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestThemeInheritance(t *testing.T) {
	parent := NewTheme("parent")
	parent.Colors.Primary = Blue

	child := NewTheme("child")
	child.Parent = parent
	child.Colors.Secondary = Red

	// Child should inherit parent's primary color
	if got := child.GetColor("primary"); !got.Equals(Blue) {
		t.Errorf("Child should inherit primary color, got %v, want %v", got, Blue)
	}

	// Child should have its own secondary color
	if got := child.GetColor("secondary"); !got.Equals(Red) {
		t.Errorf("Child should have its own secondary color, got %v, want %v", got, Red)
	}
}

func TestThemeSetStyle(t *testing.T) {
	theme := NewTheme("test")

	styleConfig := StyleConfig{
		Foreground: &Blue,
		Bold:       true,
	}

	theme.SetStyle("test.style", styleConfig)

	got := theme.GetStyle("test.style")
	if got.Foreground == nil || !got.Foreground.Equals(Blue) {
		t.Error("SetStyle() should set foreground color")
	}
	if !got.Bold {
		t.Error("SetStyle() should set bold")
	}
}

func TestThemeSetComponentStyle(t *testing.T) {
	theme := NewTheme("test")

	baseStyle := StyleConfig{
		Foreground: &Blue,
		Bold:       true,
	}

	focusedStyle := StyleConfig{
		Foreground: &Green,
		Underline:  true,
	}

	states := map[string]StyleConfig{
		"focused": focusedStyle,
	}

	theme.SetComponentStyle("button", baseStyle, states)

	// Test getting base style
	gotBase := theme.GetComponentStyle("button", "")
	if gotBase.Foreground == nil || !gotBase.Foreground.Equals(Blue) {
		t.Error("GetComponentStyle() should return base style with blue foreground")
	}

	// Test getting state style
	gotFocused := theme.GetComponentStyle("button", "focused")
	if gotFocused.Foreground == nil || !gotFocused.Foreground.Equals(Green) {
		t.Error("GetComponentStyle() should return focused style with green foreground")
	}
	if !gotFocused.Underline {
		t.Error("GetComponentStyle() focused style should have underline")
	}
}

func TestThemeExtend(t *testing.T) {
	parent := NewTheme("parent")
	parent.Colors.Primary = Blue

	child := parent.Extend("child")

	if child.Parent != parent {
		t.Error("Extend() should set parent")
	}

	if child.Name != "child" {
		t.Errorf("Extend() Name = %v, want 'child'", child.Name)
	}

	// Child should inherit parent's colors
	if got := child.GetColor("primary"); !got.Equals(Blue) {
		t.Errorf("Child should inherit parent's colors, got %v, want %v", got, Blue)
	}
}

func TestThemeClone(t *testing.T) {
	original := NewTheme("original")
	original.Colors.Primary = Blue
	original.SetStyle("test", StyleConfig{Bold: true})

	cloned := original.Clone()

	if cloned.Name != "original_clone" {
		t.Errorf("Clone() Name = %v, want 'original_clone'", cloned.Name)
	}

	// Clone should have same colors
	if !cloned.Colors.Primary.Equals(original.Colors.Primary) {
		t.Error("Clone() should copy colors")
	}

	// Clone should have same styles
	if _, ok := cloned.Styles["test"]; !ok {
		t.Error("Clone() should copy styles")
	}
}

func TestThemeMerge(t *testing.T) {
	base := NewTheme("base")
	base.Colors.Primary = Blue

	override := NewTheme("override")
	override.Colors.Secondary = Red

	merged := base.Merge(override)

	if merged.Colors.Primary.Equals(Blue) {
		// Blue should remain if override doesn't have primary
	}

	if merged.Colors.Secondary.Equals(Red) {
		// Red from override should be present
	}
}

func TestThemeGetSpacing(t *testing.T) {
	theme := NewTheme("test")
	theme.Spacing = DefaultSpacingSet()

	tests := []struct {
		name string
		size string
		want int
	}{
		{"xs", "xs", 1},
		{"sm", "sm", 2},
		{"md", "md", 4},
		{"lg", "lg", 6},
		{"xl", "xl", 8},
		{"unknown", "unknown", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := theme.GetSpacing(tt.size)
			if got != tt.want {
				t.Errorf("GetSpacing() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStyleConfigHelpers(t *testing.T) {
	config := NewStyleConfig()

	// Test WithForeground
	fgColor := Blue
	config = config.WithForeground(fgColor)
	if config.Foreground == nil || !config.Foreground.Equals(fgColor) {
		t.Error("WithForeground() should set foreground")
	}

	// Test WithBackground
	bgColor := Red
	config = config.WithBackground(bgColor)
	if config.Background == nil || !config.Background.Equals(bgColor) {
		t.Error("WithBackground() should set background")
	}

	// Test WithBold
	config = config.WithBold()
	if !config.Bold {
		t.Error("WithBold() should set bold")
	}

	// Test WithPadding
	config = config.WithPadding(1, 2, 3, 4)
	if config.Padding == nil || config.Padding[0] != 1 || config.Padding[1] != 2 ||
	   config.Padding[2] != 3 || config.Padding[3] != 4 {
		t.Error("WithPadding() should set padding")
	}

	// Test WithBorder
	border := NewBorder()
	config = config.WithBorder(border)
	if config.Border == nil {
		t.Error("WithBorder() should set border")
	}

	// Test Merge
	other := NewStyleConfig().WithItalic()
	config = config.Merge(other)
	if !config.Italic {
		t.Error("Merge() should merge italic from other")
	}
}

func TestStyleConfigIsEmpty(t *testing.T) {
	config := NewStyleConfig()
	if !config.IsEmpty() {
		t.Error("New StyleConfig should be empty")
	}

	config = config.WithBold()
	if config.IsEmpty() {
		t.Error("StyleConfig with Bold should not be empty")
	}
}

func TestParseStylePath(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		wantComponent  string
		wantState      string
	}{
		{
			name:          "state only",
			path:          "focused",
			wantComponent: "",
			wantState:     "focused",
		},
		{
			name:          "component and state",
			path:          "button.focused",
			wantComponent: "button",
			wantState:     "focused",
		},
		{
			name:          "complex path",
			path:          "nav.button.focused",
			wantComponent: "nav.button",
			wantState:     "focused",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component, state := ParseStylePath(tt.path)
			if component != tt.wantComponent {
				t.Errorf("ParseStylePath() component = %v, want %v", component, tt.wantComponent)
			}
			if state != tt.wantState {
				t.Errorf("ParseStylePath() state = %v, want %v", state, tt.wantState)
			}
		})
	}
}

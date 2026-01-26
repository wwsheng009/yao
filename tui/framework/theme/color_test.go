package theme

import (
	"testing"
)

func TestParseColor(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantType ColorType
	}{
		{"named color", "red", ColorNamed},
		{"bright color", "bright-red", ColorNamed},
		{"hex color", "#ff0000", ColorRGB},
		{"hex short", "#f00", ColorRGB},
		{"256 color", "214", Color256},
		{"rgb color", "rgb(255,0,0)", ColorRGB},
		{"empty string", "", ColorNone},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseColor(tt.input)
			if got.Type != tt.wantType {
				t.Errorf("ParseColor() type = %v, want %v", got.Type, tt.wantType)
			}
		})
	}
}

func TestColorString(t *testing.T) {
	tests := []struct {
		name  string
		color Color
		want  string
	}{
		{"named color", Red, "red"},
		{"hex color", Color{Type: ColorRGB, Value: [3]int{255, 0, 0}}, "#ff0000"},
		{"256 color", Color{Type: Color256, Value: 214}, "ansi:214"},
		{"none color", NoColor, "none"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.color.String(); got != tt.want {
				t.Errorf("Color.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestColorEquals(t *testing.T) {
	tests := []struct {
		name  string
		c1, c2 Color
		want  bool
	}{
		{"same named", Red, Red, true},
		{"different named", Red, Blue, false},
		{"same rgb", Color{Type: ColorRGB, Value: [3]int{255, 0, 0}}, Color{Type: ColorRGB, Value: [3]int{255, 0, 0}}, true},
		{"different rgb", Color{Type: ColorRGB, Value: [3]int{255, 0, 0}}, Color{Type: ColorRGB, Value: [3]int{0, 0, 255}}, false},
		{"different type", Red, Color{Type: Color256, Value: 1}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c1.Equals(tt.c2); got != tt.want {
				t.Errorf("Color.Equals() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestColorLighten(t *testing.T) {
	c := Color{Type: ColorRGB, Value: [3]int{100, 100, 100}}
	lightened := c.Lighten(50)

	if lightened.Type != ColorRGB {
		t.Errorf("Lighten() type = %v, want %v", lightened.Type, ColorRGB)
	}

	r, g, b := lightened.RGBValue()
	if r != 150 || g != 150 || b != 150 {
		t.Errorf("Lighten() RGB = %d,%d,%d, want 150,150,150", r, g, b)
	}
}

func TestColorDarken(t *testing.T) {
	c := Color{Type: ColorRGB, Value: [3]int{200, 200, 200}}
	darkened := c.Darken(25)

	if darkened.Type != ColorRGB {
		t.Errorf("Darken() type = %v, want %v", darkened.Type, ColorRGB)
	}

	r, g, b := darkened.RGBValue()
	if r != 150 || g != 150 || b != 150 {
		t.Errorf("Darken() RGB = %d,%d,%d, want 150,150,150", r, g, b)
	}
}

func TestNewColorPalette(t *testing.T) {
	palette := NewColorPalette()

	if palette.Primary.IsNone() {
		t.Error("Primary color should not be none")
	}
	if palette.Background.IsNone() {
		t.Error("Background color should not be none")
	}
	if palette.Foreground.IsNone() {
		t.Error("Foreground color should not be none")
	}
}

package theme

import (
	"testing"
)

func TestNewBorder(t *testing.T) {
	border := NewBorder()

	if !border.Enabled {
		t.Error("Border should be enabled by default")
	}
	if !border.Top || !border.Bottom || !border.Left || !border.Right {
		t.Error("All sides should be enabled by default")
	}
	if border.Style != BorderNormal {
		t.Errorf("Border style should be Normal, got %v", border.Style)
	}
}

func TestBorderStyleModifiers(t *testing.T) {
	border := NewBorder()

	// Test TopOnly
	topOnly := border.TopOnly()
	if !topOnly.Top || topOnly.Bottom || topOnly.Left || topOnly.Right {
		t.Error("TopOnly should only enable top")
	}

	// Test BottomOnly
	bottomOnly := border.BottomOnly()
	if bottomOnly.Top || !bottomOnly.Bottom || bottomOnly.Left || bottomOnly.Right {
		t.Error("BottomOnly should only enable bottom")
	}

	// Test Sides
	sides := border.Sides()
	if sides.Top || sides.Bottom || !sides.Left || !sides.Right {
		t.Error("Sides should only enable left and right")
	}

	// Test WithStyle
	thick := border.WithStyle(BorderThick)
	if thick.Style != BorderThick {
		t.Errorf("WithStyle should set style to Thick, got %v", thick.Style)
	}

	// Test WithColor
	customColor := Red
	withColor := border.WithColor(customColor)
	if !withColor.FG.Equals(customColor) {
		t.Error("WithColor should set foreground color")
	}

	// Test Disable
	disabled := border.Disable()
	if disabled.Enabled {
		t.Error("Disable should disable border")
	}

	// Test Enable
	enabled := disabled.Enable()
	if !enabled.Enabled {
		t.Error("Enable should enable border")
	}
}

func TestBorderGetEdges(t *testing.T) {
	tests := []struct {
		name      string
		border    BorderStyle
		wantTL    string
		wantT     string
		wantTR    string
	}{
		{
			name:   "normal border",
			border: BorderStyle{Style: BorderNormal},
			wantTL: "┌",
			wantT:  "─",
			wantTR: "┐",
		},
		{
			name:   "rounded border",
			border: BorderStyle{Style: BorderRounded},
			wantTL: "╭",
			wantT:  "─",
			wantTR: "╮",
		},
		{
			name:   "double border",
			border: BorderStyle{Style: BorderDouble},
			wantTL: "╔",
			wantT:  "═",
			wantTR: "╗",
		},
		{
			name:   "thick border",
			border: BorderStyle{Style: BorderThick},
			wantTL: "┏",
			wantT:  "━",
			wantTR: "┓",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edges := tt.border.GetEdges()
			if edges.TL != tt.wantTL {
				t.Errorf("TL = %v, want %v", edges.TL, tt.wantTL)
			}
			if edges.T != tt.wantT {
				t.Errorf("T = %v, want %v", edges.T, tt.wantT)
			}
			if edges.TR != tt.wantTR {
				t.Errorf("TR = %v, want %v", edges.TR, tt.wantTR)
			}
		})
	}
}

func TestBorderRender(t *testing.T) {
	border := NewBorder()
	lines := border.Render(10, 5)

	if len(lines) != 5 {
		t.Errorf("Render() should return 5 lines, got %d", len(lines))
	}

	// Check that top border contains corner characters (ignoring ANSI codes)
	if !containsRunes(lines[0], "┌") || !containsRunes(lines[0], "┐") {
		t.Error("Top border should contain corner characters")
	}

	// All lines should be non-empty
	for i, line := range lines {
		if line == "" {
			t.Errorf("Line %d should not be empty", i)
		}
	}
}

func TestGetContentWidth(t *testing.T) {
	tests := []struct {
		name        string
		border      BorderStyle
		width       int
		wantContent int
	}{
		{
			name:        "disabled border",
			border:      BorderStyle{Enabled: false},
			width:       10,
			wantContent: 10,
		},
		{
			name:        "normal border",
			border:      NewBorder(),
			width:       10,
			wantContent: 8,
		},
		{
			name:        "small width with border",
			border:      NewBorder(),
			width:       2,
			wantContent: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.border.GetContentWidth(tt.width)
			if got != tt.wantContent {
				t.Errorf("GetContentWidth() = %v, want %v", got, tt.wantContent)
			}
		})
	}
}

func TestGetContentHeight(t *testing.T) {
	tests := []struct {
		name        string
		border      BorderStyle
		height      int
		wantContent int
	}{
		{
			name:        "disabled border",
			border:      BorderStyle{Enabled: false},
			height:      10,
			wantContent: 10,
		},
		{
			name:        "normal border",
			border:      NewBorder(),
			height:      10,
			wantContent: 8,
		},
		{
			name:        "small height with border",
			border:      NewBorder(),
			height:      2,
			wantContent: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.border.GetContentHeight(tt.height)
			if got != tt.wantContent {
				t.Errorf("GetContentHeight() = %v, want %v", got, tt.wantContent)
			}
		})
	}
}

func TestBorderThickness(t *testing.T) {
	border := NewBorder()
	top, right, bottom, left := border.BorderThickness()

	if top != 1 || right != 1 || bottom != 1 || left != 1 {
		t.Errorf("BorderThickness() = %d,%d,%d,%d, want 1,1,1,1", top, right, bottom, left)
	}

	// Test disabled border
	border = border.Disable()
	top, right, bottom, left = border.BorderThickness()

	if top != 0 || right != 0 || bottom != 0 || left != 0 {
		t.Errorf("Disabled BorderThickness() = %d,%d,%d,%d, want 0,0,0,0", top, right, bottom, left)
	}

	// Test partial border
	border = NewBorder().TopOnly()
	top, right, bottom, left = border.BorderThickness()

	if top != 1 || right != 0 || bottom != 0 || left != 0 {
		t.Errorf("TopOnly BorderThickness() = %d,%d,%d,%d, want 1,0,0,0", top, right, bottom, left)
	}
}

func containsRunes(s string, substr string) bool {
	for _, r := range substr {
		if !containsRune(s, r) {
			return false
		}
	}
	return true
}

func containsRune(s string, r rune) bool {
	for _, c := range s {
		if c == r {
			return true
		}
	}
	return false
}

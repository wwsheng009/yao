package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yaoapp/yao/tui/ui/components"
	runtimepkg "github.com/yaoapp/yao/tui/runtime"
)

// TestRuntimeVisualStyle tests the enhanced visual styling system
func TestRuntimeVisualStyle(t *testing.T) {
	config := &Config{
		Name:      "Visual Style Test",
		UseRuntime: true,
		Layout: Layout{
			Direction: "column",
			Children: []Component{
				{
					Type: "text",
					ID:   "styled-text",
					Props: map[string]interface{}{
						"content": "Styled Text",
					},
				},
			},
		},
	}

	// Create model with config
	model := NewModel(config, nil)
	model.Width = 80
	model.Height = 24

	// Initialize the model
	model.Init()

	// Send WindowSizeMsg to trigger layout
	windowMsg := tea.WindowSizeMsg{Width: 80, Height: 24}
	newModel, _ := model.Update(windowMsg)
	model = newModel.(*Model)

	// Get the Text component and apply styles
	root := model.RuntimeRoot
	if root == nil || len(root.Children) == 0 {
		t.Fatal("RuntimeRoot should have children")
	}

	textNode := root.Children[0]
	if textNode.Component == nil {
		t.Fatal("Text node should have a component")
	}

	// Unwrap NativeComponentWrapper to get actual component
	textWrapper, ok := textNode.Component.Instance.(*NativeComponentWrapper)
	if !ok {
		t.Fatalf("Expected NativeComponentWrapper, got %T", textNode.Component.Instance)
	}

	textComp, ok := textWrapper.Component.(*components.TextComponent)
	if !ok {
		t.Fatalf("Expected TextComponent, got %T", textWrapper.Component)
	}

	// Test 1: Bold text
	boldView := textComp.WithBold(true).View()
	if !strings.Contains(boldView, "Styled Text") {
		t.Error("Bold view should contain text")
	}
	t.Logf("Bold text: %q", boldView)

	// Test 2: Colored text
	coloredView := textComp.
		WithForeground("#FF0000").
		WithBackground("#0000FF").
		WithBold(true).
		View()
	if !strings.Contains(coloredView, "Styled Text") {
		t.Error("Colored view should contain text")
	}
	t.Logf("Colored text: %q", coloredView)

	// Test 3: Text with border
	borderedView := textComp.
		WithBorder("rounded").
		WithBorderForeground("#00FF00").
		View()
	if !strings.Contains(borderedView, "Styled Text") {
		t.Error("Bordered view should contain text")
	}
	t.Logf("Bordered text (first 100 chars): %s", truncateString(borderedView, 100))

	// Test 4: Centered text
	centeredView := textComp.
		WithAlign("center").
		WithWidth(40).
		View()
	if !strings.Contains(centeredView, "Styled Text") {
		t.Error("Centered view should contain text")
	}
	t.Logf("Centered text: %q", centeredView)

	// Test 5: Text with margin
	marginedView := textComp.
		WithMargin(1, 2, 1, 2).
		View()
	if !strings.Contains(marginedView, "Styled Text") {
		t.Error("Margined view should contain text")
	}
	t.Logf("Margined text: %q", marginedView)

	// Test 6: Text with padding
	paddedView := textComp.
		WithPadding(1, 2, 1, 2).
		View()
	if !strings.Contains(paddedView, "Styled Text") {
		t.Error("Padded view should contain text")
	}
	t.Logf("Padded text: %q", paddedView)

	// Test 7: Combined styles (bold, italic, underline, colored)
	combinedView := textComp.
		WithContent("Combined Styles").
		WithBold(true).
		WithItalic(true).
		WithUnderline(true).
		WithForeground("#FF00FF").
		WithBackground("#333333").
		WithBorder("thick").
		WithBorderForeground("#FFFF00").
		WithMargin(1, 1, 1, 1).
		WithPadding(1, 1, 1, 1).
		View()
	if !strings.Contains(combinedView, "Combined Styles") {
		t.Error("Combined view should contain text")
	}
	t.Logf("Combined styles (first 150 chars): %s", truncateString(combinedView, 150))

	// Test 8: VisualStyle direct manipulation
	vs := textComp.GetVisualStyle()
	updatedStyle := vs.WithForeground("#00FF00").WithBold(true)
	updatedStyle.Style.Width = 30
	textComp.SetVisualStyle(updatedStyle)
	directView := textComp.View()
	if !strings.Contains(directView, "Combined Styles") {
		t.Error("Direct style manipulation should work")
	}
	t.Logf("Direct style: %q", directView)

	t.Log("Visual style test passed!")
}

// TestRuntimeVisualStylePresets tests common visual style presets
func TestRuntimeVisualStylePresets(t *testing.T) {
	text := components.NewTextComponent("Preset Test")

	// Preset 1: Error style (red background, white text, bold)
	errorView := text.
		WithForeground("#FFFFFF").
		WithBackground("#FF0000").
		WithBold(true).
		WithPadding(1, 2, 1, 2).
		View()
	if !strings.Contains(errorView, "Preset Test") {
		t.Error("Error preset should contain text")
	}
	t.Logf("Error preset: %q", errorView)

	// Preset 2: Success style (green background, white text)
	successView := text.
		WithForeground("#FFFFFF").
		WithBackground("#00FF00").
		WithBold(true).
		WithPadding(1, 2, 1, 2).
		View()
	if !strings.Contains(successView, "Preset Test") {
		t.Error("Success preset should contain text")
	}
	t.Logf("Success preset: %q", successView)

	// Preset 3: Warning style (yellow background, black text, rounded border)
	warningView := text.
		WithForeground("#000000").
		WithBackground("#FFFF00").
		WithBold(true).
		WithBorder("rounded").
		WithBorderForeground("#FFA500").
		WithPadding(1, 2, 1, 2).
		View()
	if !strings.Contains(warningView, "Preset Test") {
		t.Error("Warning preset should contain text")
	}
	t.Logf("Warning preset (first 100 chars): %s", truncateString(warningView, 100))

	// Preset 4: Info style (blue background, white text, rounded border)
	infoView := text.
		WithForeground("#FFFFFF").
		WithBackground("#0000FF").
		WithBorder("rounded").
		WithBorderForeground("#87CEEB").
		WithPadding(1, 2, 1, 2).
		View()
	if !strings.Contains(infoView, "Preset Test") {
		t.Error("Info preset should contain text")
	}
	t.Logf("Info preset (first 100 chars): %s", truncateString(infoView, 100))

	t.Log("Visual style presets test passed!")
}

// TestRuntimeVisualStyleColorPalettes tests the predefined color palettes
func TestRuntimeVisualStyleColorPalettes(t *testing.T) {
	text := components.NewTextComponent("Color Palette Test")

	// Test default palette
	defaultColors := runtimepkg.GetColorPalette("default")
	if defaultColors.Primary == "" {
		t.Error("Default palette should have colors")
	}

	// Test Dracula palette
	draculaColors := runtimepkg.GetColorPalette("dracula")
	if draculaColors.Primary == "" {
		t.Error("Dracula palette should have colors")
	}

	// Test Nord palette
	nordColors := runtimepkg.GetColorPalette("nord")
	if nordColors.Primary == "" {
		t.Error("Nord palette should have colors")
	}

	// Test Monokai palette
	monokaiColors := runtimepkg.GetColorPalette("monokai")
	if monokaiColors.Primary == "" {
		t.Error("Monokai palette should have colors")
	}

	// Apply palette colors
	paletteView := text.
		WithForeground(draculaColors.Primary).
		WithBackground(draculaColors.Muted).
		WithBold(true).
		View()
	if !strings.Contains(paletteView, "Color Palette Test") {
		t.Error("Palette view should contain text")
	}
	t.Logf("Palette (Dracula) view: %q", paletteView)

	t.Log("Color palettes test passed!")
}

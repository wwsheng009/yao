// Package runtime provides export functionality demo.
//
// This test file demonstrates the export functionality by creating
// a sample layout and exporting it to various formats.
package runtime

import (
	"os"
	"path/filepath"
	"testing"
)

// TestExportDemo creates a demo layout and exports it to all formats.
// Run with: go test -v -run TestExportDemo ./tui/runtime
func TestExportDemo(t *testing.T) {
	// Create a 80x24 frame with a chat interface
	buf := NewCellBuffer(80, 24)

	// Draw outer border first (80x24)
	// Top border with title
	for x := 1; x < 79; x++ {
		buf.SetContent(x, 0, 0, '─', CellStyle{}, "")
	}
	// Left border
	for y := 1; y < 23; y++ {
		buf.SetContent(0, y, 0, '│', CellStyle{}, "")
	}
	// Right border
	for y := 1; y < 23; y++ {
		buf.SetContent(79, y, 0, '│', CellStyle{}, "")
	}
	// Bottom border
	for x := 1; x < 79; x++ {
		buf.SetContent(x, 23, 0, '─', CellStyle{}, "")
	}
	// Corners
	buf.SetContent(0, 0, 0, '┌', CellStyle{}, "")
	buf.SetContent(79, 0, 0, '┐', CellStyle{}, "")
	buf.SetContent(0, 23, 0, '└', CellStyle{}, "")
	buf.SetContent(79, 23, 0, '┘', CellStyle{}, "")

	// Header title (inside top border)
	renderText(buf, 2, 0, " Gemini Chat - Export Demo ", CellStyle{Bold: true})

	// Second line (separator below header)
	for x := 1; x < 79; x++ {
		buf.SetContent(x, 2, 0, '─', CellStyle{}, "")
	}
	// Left/right connection for separator
	buf.SetContent(0, 2, 0, '├', CellStyle{}, "")
	buf.SetContent(79, 2, 0, '┤', CellStyle{}, "")

	// Sidebar section (left side, inside border)
	sidebarY := 4
	renderText(buf, 2, sidebarY, "Recent", CellStyle{Bold: true})
	sidebarY++
	renderText(buf, 2, sidebarY, "", CellStyle{})
	sidebarY++
	renderText(buf, 4, sidebarY, "Today", CellStyle{Bold: true})
	sidebarY++
	renderText(buf, 4, sidebarY, "Project Planning", CellStyle{})
	sidebarY++
	renderText(buf, 4, sidebarY, "Code Review", CellStyle{})
	sidebarY++
	renderText(buf, 2, sidebarY, "Yesterday", CellStyle{Bold: true})
	sidebarY++
	renderText(buf, 4, sidebarY, "Bug Discussion", CellStyle{})
	sidebarY++
	renderText(buf, 2, sidebarY, "Previous 7 days", CellStyle{})

	// Chat area divider (vertical line)
	for y := 3; y < 23; y++ {
		buf.SetContent(21, y, 0, '│', CellStyle{}, "")
	}
	// Top and bottom T for divider
	buf.SetContent(21, 2, 0, '┼', CellStyle{}, "")
	buf.SetContent(21, 23, 0, '┴', CellStyle{}, "")

	// Chat area (right side, inside border)
	chatY := 4

	chatLines := []struct {
		text  string
		style CellStyle
	}{
		{"Today", CellStyle{Bold: true}},
		{"", CellStyle{}},
		{"Project Planning", CellStyle{Bold: true}},
		{"  👤 You: Can you help me understand the Yao TUI", CellStyle{}},
		{"          layout system?", CellStyle{}},
		{"  💎 Gem: Of course! The Yao TUI runtime uses a", CellStyle{Italic: true}},
		{"          flex layout algorithm with three phases:", CellStyle{Italic: true}},
		{"          Measure, Layout, and Render.", CellStyle{Italic: true}},
		{"", CellStyle{}},
		{"Code Review", CellStyle{Bold: true}},
		{"  👤 You: That sounds interesting. How do I create", CellStyle{}},
		{"          a simple layout?", CellStyle{}},
		{"", CellStyle{}},
		{"Yesterday", CellStyle{Bold: true}},
		{"Bug Discussion", CellStyle{Bold: true}},
		{"  💎 Gem: Use a row container with a fixed-width", CellStyle{Italic: true}},
		{"          sidebar for the main content area.", CellStyle{Italic: true}},
		{"", CellStyle{}},
		{"  👤 You: Great! And what about nested layouts?", CellStyle{}},
		{"", CellStyle{}},
		{"Previous 7 days", CellStyle{Bold: true}},
		{"", CellStyle{}},
		{"[Type a message...]", CellStyle{Italic: true}},
	}

	for _, line := range chatLines {
		if chatY < 22 {
			renderText(buf, 22, chatY, line.text, line.style)
			chatY++
		}
	}

	// Create frame
	frame := Frame{
		Buffer: buf,
		Width:  80,
		Height: 24,
		Dirty:  true,
	}

	// Create layout result (empty for demo)
	result := LayoutResult{
		Boxes:      []LayoutBox{},
		Dirty:      true,
		RootWidth:  80,
		RootHeight: 24,
	}

	// Create export directory
	exportDir := "exports"
	if err := os.MkdirAll(exportDir, 0755); err != nil {
		t.Fatalf("Failed to create export directory: %v", err)
	}

	// Create exporter
	exporter := NewExporter(&frame, &result)

	// Export to TXT
	txtPath := filepath.Join(exportDir, "demo.txt")
	if err := exporter.SaveToTXT(txtPath); err != nil {
		t.Fatalf("Failed to export to TXT: %v", err)
	}
	t.Logf("✓ Exported to TXT: %s", txtPath)

	// Export to SVG
	svgPath := filepath.Join(exportDir, "demo.svg")
	if err := exporter.SaveToSVG(svgPath); err != nil {
		t.Fatalf("Failed to export to SVG: %v", err)
	}
	t.Logf("✓ Exported to SVG: %s", svgPath)

	// Export to PNG (dark theme)
	pngPath := filepath.Join(exportDir, "demo.png")
	if err := exporter.SaveToPNG(pngPath); err != nil {
		t.Fatalf("Failed to export to PNG: %v", err)
	}
	t.Logf("✓ Exported to PNG: %s", pngPath)

	// Export to PNG (light theme)
	exporter.SetColorScheme(LightColorScheme())
	pngLightPath := filepath.Join(exportDir, "demo-light.png")
	if err := exporter.SaveToPNG(pngLightPath); err != nil {
		t.Fatalf("Failed to export to PNG (light): %v", err)
	}
	t.Logf("✓ Exported to PNG (light): %s", pngLightPath)

	// Export to HTML
	htmlPath := filepath.Join(exportDir, "demo.html")
	htmlContent, err := exporter.ToHTML()
	if err != nil {
		t.Fatalf("Failed to generate HTML: %v", err)
	}
	if err := os.WriteFile(htmlPath, []byte(htmlContent), 0644); err != nil {
		t.Fatalf("Failed to save HTML: %v", err)
	}
	t.Logf("✓ Exported to HTML: %s", htmlPath)

	// Debug info
	debug := DebugFrame(&frame, &result)
	t.Logf("\n=== Export Summary ===")
	t.Logf("%s", debug.Summary)
	t.Logf("Buffer: %dx%d, %d/%d cells non-empty (%.1f%%)",
		frame.Buffer.Width(), frame.Buffer.Height(),
		debug.BufferInfo.NonEmpty, debug.BufferInfo.TotalCells,
		float64(debug.BufferInfo.NonEmpty)*100/float64(debug.BufferInfo.TotalCells))
}

// TestExportBase64PNG demonstrates base64 PNG export.
func TestExportBase64PNG(t *testing.T) {
	buf := NewCellBuffer(40, 10)

	lines := []struct {
		text  string
		style CellStyle
	}{
		{"Hello, World!", CellStyle{Bold: true, Underline: true}},
		{"", CellStyle{}},
		{"This is a demo of the", CellStyle{}},
		{"Yao TUI export system.", CellStyle{}},
		{"", CellStyle{}},
		{"Supports: TXT, SVG, PNG, HTML", CellStyle{Italic: true}},
	}

	y := 0
	for _, line := range lines {
		renderText(buf, 0, y, line.text, line.style)
		y++
	}

	frame := Frame{
		Buffer: buf,
		Width:  40,
		Height: 10,
		Dirty:  true,
	}

	result := LayoutResult{
		Boxes:      []LayoutBox{},
		Dirty:      true,
		RootWidth:  40,
		RootHeight: 10,
	}

	exporter := NewExporter(&frame, &result)
	base64, err := exporter.ToBase64PNG()
	if err != nil {
		t.Fatalf("Failed to generate base64 PNG: %v", err)
	}

	t.Logf("Base64 PNG length: %d bytes", len(base64))
	t.Logf("Base64 prefix: %s...", truncateString(base64, 50))
}

// TestExportSimple demonstrates a simple export example.
func TestExportSimple(t *testing.T) {
	buf := NewCellBuffer(30, 5)

	// Draw a simple box
	renderLine(buf, 0, 0, 30)
	renderLine(buf, 0, 4, 30)
	for y := 1; y < 4; y++ {
		buf.SetContent(0, y, 0, '│', CellStyle{}, "")
		buf.SetContent(29, y, 0, '│', CellStyle{}, "")
	}
	renderText(buf, 12, 2, "BOX", CellStyle{Bold: true})

	frame := Frame{
		Buffer: buf,
		Width:  30,
		Height: 5,
		Dirty:  true,
	}

	result := LayoutResult{
		Boxes:      []LayoutBox{},
		Dirty:      true,
		RootWidth:  30,
		RootHeight: 5,
	}

	exportDir := "exports"
	os.MkdirAll(exportDir, 0755)

	exporter := NewExporter(&frame, &result)

	svgPath := filepath.Join(exportDir, "simple.svg")
	if err := exporter.SaveToSVG(svgPath); err != nil {
		t.Fatalf("Failed to export simple SVG: %v", err)
	}
	t.Logf("✓ Exported simple box to: %s", svgPath)
}

// Helper functions

func renderText(buf *CellBuffer, x, y int, text string, style CellStyle) {
	for i, char := range text {
		if x+i < buf.Width() {
			buf.SetContent(x+i, y, 0, char, style, "")
		}
	}
}

func renderLine(buf *CellBuffer, x, y, width int) {
	for i := 0; i < width && x+i < buf.Width(); i++ {
		buf.SetContent(x+i, y, 0, '─', CellStyle{}, "")
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

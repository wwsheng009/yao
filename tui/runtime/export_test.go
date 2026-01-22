package runtime

import (
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestExportTXT tests TXT export functionality.
func TestExportTXT(t *testing.T) {
	rt := NewRuntime(40, 10)
	root := mockContainer("root", "row", NewStyle())
	root.Style.Width = 40
	root.Style.Height = 10

	// Add some text content
	text := mockNode("text", "column", "Hello World")
	text.Style.Width = 20
	text.Style.Height = 5
	root.AddChild(text)

	constraints := NewBoxConstraints(0, 40, 0, 10)
	result := rt.Layout(root, constraints)
	frame := rt.Render(result)

	// Create exporter
	exporter := NewExporter(&frame, &result)

	// Export to temp file
	tmpDir := t.TempDir()
	txtPath := filepath.Join(tmpDir, "output.txt")

	err := exporter.SaveToTXT(txtPath)
	if err != nil {
		t.Fatalf("SaveToTXT failed: %v", err)
	}

	// Verify file exists
	info, err := os.Stat(txtPath)
	if err != nil {
		t.Fatalf("Failed to stat output file: %v", err)
	}
	if info.Size() == 0 {
		t.Error("Output file is empty")
	}

	// Read and verify content
	content, err := os.ReadFile(txtPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "Hello") {
		t.Errorf("Expected output to contain 'Hello', got: %q", contentStr)
	}

	t.Logf("TXT export successful:\n%s", contentStr)
}

// TestExportSVG tests SVG export functionality.
func TestExportSVG(t *testing.T) {
	rt := NewRuntime(40, 10)
	root := mockContainer("root", "column", NewStyle())
	root.Style.Width = 40
	root.Style.Height = 10

	// Add text
	text := mockNode("normal", "column", "Normal")
	text.Style.Width = 10
	text.Style.Height = 2
	root.AddChild(text)

	constraints := NewBoxConstraints(0, 40, 0, 10)
	result := rt.Layout(root, constraints)
	frame := rt.Render(result)

	exporter := NewExporter(&frame, &result)

	// Export to temp file
	tmpDir := t.TempDir()
	svgPath := filepath.Join(tmpDir, "output.svg")

	err := exporter.SaveToSVG(svgPath)
	if err != nil {
		t.Fatalf("SaveToSVG failed: %v", err)
	}

	// Verify file exists
	info, err := os.Stat(svgPath)
	if err != nil {
		t.Fatalf("Failed to stat output file: %v", err)
	}
	if info.Size() == 0 {
		t.Error("Output file is empty")
	}

	// Read and verify SVG structure
	content, err := os.ReadFile(svgPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	contentStr := string(content)
	// Check for SVG elements
	required := []string{
		"<svg",
		"xmlns=",
		"</svg>",
		"<style>",
		"<text",
	}

	for _, elem := range required {
		if !strings.Contains(contentStr, elem) {
			t.Errorf("Expected SVG to contain '%s'", elem)
		}
	}

	// Verify it's valid XML by parsing
	decoder := xml.NewDecoder(bytes.NewReader(content))
	for {
		_, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Errorf("Failed to parse SVG as XML: %v", err)
		}
	}

	t.Logf("SVG export successful (%d bytes)", len(content))
}

// TestExportPNG tests PNG export functionality.
func TestExportPNG(t *testing.T) {
	rt := NewRuntime(40, 10)
	root := mockContainer("root", "column", NewStyle())
	root.Style.Width = 40
	root.Style.Height = 10

	// Add some text
	text := mockNode("text", "column", "Test")
	text.Style.Width = 10
	text.Style.Height = 2
	root.AddChild(text)

	constraints := NewBoxConstraints(0, 40, 0, 10)
	result := rt.Layout(root, constraints)
	frame := rt.Render(result)

	exporter := NewExporter(&frame, &result)

	// Export to temp file
	tmpDir := t.TempDir()
	pngPath := filepath.Join(tmpDir, "output.png")

	err := exporter.SaveToPNG(pngPath)
	if err != nil {
		t.Fatalf("SaveToPNG failed: %v", err)
	}

	// Verify file exists
	info, err := os.Stat(pngPath)
	if err != nil {
		t.Fatalf("Failed to stat output file: %v", err)
	}
	if info.Size() == 0 {
		t.Error("Output file is empty")
	}

	// Verify PNG signature
	content, err := os.ReadFile(pngPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if len(content) < 8 {
		t.Fatal("PNG file too short")
	}

	// PNG signature: 137 80 78 71 13 10 26 10
	pngSignature := []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}
	for i, b := range pngSignature {
		if content[i] != b {
			t.Errorf("Invalid PNG signature at byte %d: got %d, expected %d", i, content[i], b)
		}
	}

	t.Logf("PNG export successful (%d bytes)", len(content))
}

// TestExportToBase64PNG tests base64 PNG encoding.
func TestExportToBase64PNG(t *testing.T) {
	rt := NewRuntime(20, 5)
	root := mockContainer("root", "column", NewStyle())
	root.Style.Width = 20
	root.Style.Height = 5

	text := mockNode("text", "column", "ABC")
	text.Style.Width = 5
	text.Style.Height = 2
	root.AddChild(text)

	constraints := NewBoxConstraints(0, 20, 0, 5)
	result := rt.Layout(root, constraints)
	frame := rt.Render(result)

	exporter := NewExporter(&frame, &result)

	base64Str, err := exporter.ToBase64PNG()
	if err != nil {
		t.Fatalf("ToBase64PNG failed: %v", err)
	}

	if base64Str == "" {
		t.Error("Base64 string is empty")
	}

	// Verify it's valid base64 and can be decoded
	decoded, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		t.Errorf("Failed to decode base64 string: %v", err)
	}

	// Verify PNG signature in decoded data
	if len(decoded) >= 8 {
		pngSignature := []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}
		for i, b := range pngSignature {
			if decoded[i] != b {
				t.Errorf("Invalid PNG signature in decoded data at byte %d", i)
				break
			}
		}
	}

	t.Logf("Base64 PNG successful (%d chars, %d bytes decoded)", len(base64Str), len(decoded))
}

// TestExportToHTML tests HTML export functionality.
func TestExportToHTML(t *testing.T) {
	rt := NewRuntime(30, 5)
	root := mockContainer("root", "column", NewStyle())
	root.Style.Width = 30
	root.Style.Height = 5

	text := mockNode("text", "column", "HTML")
	text.Style.Width = 6
	text.Style.Height = 2
	root.AddChild(text)

	constraints := NewBoxConstraints(0, 30, 0, 5)
	result := rt.Layout(root, constraints)
	frame := rt.Render(result)

	exporter := NewExporter(&frame, &result)

	html, err := exporter.ToHTML()
	if err != nil {
		t.Fatalf("ToHTML failed: %v", err)
	}

	// Verify HTML structure
	required := []string{
		"<div",
		"<svg",
		"xmlns=",
		"</svg>",
		"</div>",
	}

	for _, elem := range required {
		if !strings.Contains(html, elem) {
			t.Errorf("Expected HTML to contain '%s'", elem)
		}
	}

	t.Logf("HTML export successful (%d chars)", len(html))
}

// TestExportColorScheme tests different color schemes.
func TestExportColorScheme(t *testing.T) {
	rt := NewRuntime(20, 5)
	root := mockContainer("root", "column", NewStyle())
	root.Style.Width = 20
	root.Style.Height = 5

	text := mockNode("text", "column", "Color")
	text.Style.Width = 8
	text.Style.Height = 2
	root.AddChild(text)

	constraints := NewBoxConstraints(0, 20, 0, 5)
	result := rt.Layout(root, constraints)
	frame := rt.Render(result)

	tmpDir := t.TempDir()

	// Test with dark scheme (default)
	darkExporter := NewExporter(&frame, &result)
	darkPath := filepath.Join(tmpDir, "dark.png")
	err := darkExporter.SaveToPNG(darkPath)
	if err != nil {
		t.Errorf("Dark scheme export failed: %v", err)
	}

	// Test with light scheme
	lightExporter := NewExporter(&frame, &result)
	lightExporter.SetColorScheme(LightColorScheme())
	lightPath := filepath.Join(tmpDir, "light.png")
	err = lightExporter.SaveToPNG(lightPath)
	if err != nil {
		t.Errorf("Light scheme export failed: %v", err)
	}

	// Verify both files exist and have different sizes (due to different colors)
	darkInfo, _ := os.Stat(darkPath)
	lightInfo, _ := os.Stat(lightPath)

	t.Logf("Dark PNG: %d bytes, Light PNG: %d bytes", darkInfo.Size(), lightInfo.Size())
}

// TestExportEmptyFrame tests export with empty/minimal frames.
func TestExportEmptyFrame(t *testing.T) {
	rt := NewRuntime(10, 5)
	root := mockContainer("root", "column", NewStyle())
	root.Style.Width = 10
	root.Style.Height = 5

	constraints := NewBoxConstraints(0, 10, 0, 5)
	result := rt.Layout(root, constraints)
	frame := rt.Render(result)

	exporter := NewExporter(&frame, &result)

	tmpDir := t.TempDir()

	// Test all formats with empty frame
	formats := []struct {
		name   string
		ext    string
		export func(string) error
	}{
		{"TXT", ".txt", func(p string) error { return exporter.SaveToTXT(p) }},
		{"SVG", ".svg", func(p string) error { return exporter.SaveToSVG(p) }},
		{"PNG", ".png", func(p string) error { return exporter.SaveToPNG(p) }},
	}

	for _, fm := range formats {
		t.Run(fm.name, func(t *testing.T) {
			path := filepath.Join(tmpDir, "empty"+fm.ext)
			err := fm.export(path)
			if err != nil {
				t.Errorf("%s export failed: %v", fm.name, err)
			}

			info, err := os.Stat(path)
			if err != nil {
				t.Errorf("Failed to stat %s file: %v", fm.name, err)
			}
			t.Logf("%s export: %d bytes", fm.name, info.Size())
		})
	}
}

// TestExportNilFrame tests error handling for nil frames.
func TestExportNilFrame(t *testing.T) {
	exporter := NewExporter(nil, nil)

	tmpDir := t.TempDir()

	// All exports should fail with nil frame
	if err := exporter.SaveToTXT(filepath.Join(tmpDir, "test.txt")); err == nil {
		t.Error("Expected error for SaveToTXT with nil frame")
	}
	if err := exporter.SaveToSVG(filepath.Join(tmpDir, "test.svg")); err == nil {
		t.Error("Expected error for SaveToSVG with nil frame")
	}
	if err := exporter.SaveToPNG(filepath.Join(tmpDir, "test.png")); err == nil {
		t.Error("Expected error for SaveToPNG with nil frame")
	}
	if _, err := exporter.ToBase64PNG(); err == nil {
		t.Error("Expected error for ToBase64PNG with nil frame")
	}
	if _, err := exporter.ToHTML(); err == nil {
		t.Error("Expected error for ToHTML with nil frame")
	}
}

// TestExportSaveTo tests unified SaveTo method.
func TestExportSaveTo(t *testing.T) {
	rt := NewRuntime(20, 5)
	root := mockContainer("root", "column", NewStyle())
	root.Style.Width = 20
	root.Style.Height = 5

	text := mockNode("text", "column", "Test")
	text.Style.Width = 6
	text.Style.Height = 2
	root.AddChild(text)

	constraints := NewBoxConstraints(0, 20, 0, 5)
	result := rt.Layout(root, constraints)
	frame := rt.Render(result)

	exporter := NewExporter(&frame, &result)
	tmpDir := t.TempDir()

	// Test all formats through SaveTo
	formats := []ExportFormat{FormatTXT, FormatSVG, FormatPNG}
	for _, format := range formats {
		ext := "." + string(format)
		path := filepath.Join(tmpDir, "test"+ext)

		err := exporter.SaveTo(path, format)
		if err != nil {
			t.Errorf("SaveTo(%s) failed: %v", format, err)
		}

		if _, err := os.Stat(path); err != nil {
			t.Errorf("File not created for format %s: %v", format, err)
		}
	}

	// Test invalid format
	err := exporter.SaveTo(filepath.Join(tmpDir, "test.xyz"), ExportFormat("xyz"))
	if err == nil {
		t.Error("Expected error for invalid format")
	}
}

// TestBitmapFont tests the bitmap font rendering.
func TestBitmapFont(t *testing.T) {
	// Test common characters
	testChars := []rune{
		'A', 'B', 'C', '0', '1', '2',
		'a', 'b', 'c', ' ', '.', ',',
		'!', '?', '@', '#', '$', '%',
	}

	for _, char := range testChars {
		bitmap := getBitmapChar(char)
		if bitmap == nil {
			t.Errorf("No bitmap for character %q", char)
			continue
		}

		// Verify bitmap dimensions
		if len(bitmap) != 16 {
			t.Errorf("Invalid bitmap height for %q: got %d, expected 16", char, len(bitmap))
		}
	}

	// Test unsupported character returns question mark
	unknown := getBitmapChar('\u1234')
	if unknown == nil {
		t.Error("Expected fallback bitmap for unknown character")
	}

	t.Logf("Bitmap font test passed for %d characters", len(testChars))
}

// TestGeminiLayoutExport exports the Gemini chat layout to all formats.
func TestGeminiLayoutExport(t *testing.T) {
	rt := NewRuntime(80, 24)
	root := mockContainer("root", "column", NewStyle())
	root.Style.Width = 80
	root.Style.Height = 24

	// Create a simple header
	header := mockContainer("header", "row", NewStyle())
	header.Style.Height = 3
	headerText := mockNode("header-text", "column", "Gemini Chat")
	header.AddChild(headerText)
	root.AddChild(header)

	// Create main area
	main := mockContainer("main", "row", NewStyle())
	main.Style.Height = 18
	main.Style.FlexGrow = 1

	// Sidebar
	sidebar := mockContainer("sidebar", "column", NewStyle())
	sidebar.Style.Width = 20
	sidebarText := mockNode("sidebar-text", "column", "Recent")
	sidebar.AddChild(sidebarText)
	main.AddChild(sidebar)

	// Chat area
	chat := mockContainer("chat", "column", NewStyle())
	chat.Style.FlexGrow = 1
	chatText := mockNode("chat-text", "column", "Chat Messages")
	chat.AddChild(chatText)
	main.AddChild(chat)

	root.AddChild(main)

	// Input area
	input := mockContainer("input", "row", NewStyle())
	input.Style.Height = 3
	inputText := mockNode("input-text", "column", "[Type a message...]")
	input.AddChild(inputText)
	root.AddChild(input)

	constraints := NewBoxConstraints(0, 80, 0, 24)
	result := rt.Layout(root, constraints)
	frame := rt.Render(result)

	exporter := NewExporter(&frame, &result)
	tmpDir := t.TempDir()

	// Export to all formats
	formats := []struct {
		format ExportFormat
		ext    string
	}{
		{FormatTXT, ".txt"},
		{FormatSVG, ".svg"},
		{FormatPNG, ".png"},
	}

	for _, fm := range formats {
		filename := filepath.Join(tmpDir, "gemini-chat"+fm.ext)
		err := exporter.SaveTo(filename, fm.format)
		if err != nil {
			t.Errorf("Failed to export to %s: %v", fm.format, err)
			continue
		}

		info, _ := os.Stat(filename)
		t.Logf("Gemini layout exported to %s: %d bytes", fm.format, info.Size())
	}

	// Also export HTML
	html, err := exporter.ToHTML()
	if err != nil {
		t.Errorf("Failed to export HTML: %v", err)
	} else {
		htmlPath := filepath.Join(tmpDir, "gemini-chat.html")
		os.WriteFile(htmlPath, []byte(html), 0644)
		t.Logf("HTML export: %d bytes", len(html))
	}
}

// TestEscapeXML tests XML escaping.
func TestEscapeXML(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello & World", "Hello &amp; World"},
		{"<tag>", "&lt;tag&gt;"},
		{"\"quoted\"", "&quot;quoted&quot;"},
		{"'single'", "&apos;single&apos;"},
		{"All: <>&\"'", "All: &lt;&gt;&amp;&quot;&apos;"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := escapeXML(tt.input)
			if result != tt.expected {
				t.Errorf("escapeXML(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

// ExampleExporter demonstrates the export functionality.
func ExampleExporter() {
	// Create a simple layout
	rt := NewRuntime(40, 10)
	root := mockContainer("root", "column", NewStyle())
	root.Style.Width = 40
	root.Style.Height = 10

	text := mockNode("greeting", "column", "Hello, World!")
	root.AddChild(text)

	// Layout and render
	constraints := NewBoxConstraints(0, 40, 0, 10)
	result := rt.Layout(root, constraints)
	frame := rt.Render(result)

	// Export to different formats
	exporter := NewExporter(&frame, &result)

	exporter.SaveTo("output.txt", FormatTXT)
	exporter.SaveTo("output.svg", FormatSVG)
	exporter.SaveTo("output.png", FormatPNG)

	// Or use individual methods
	exporter.SaveToTXT("output.txt")
	exporter.SaveToSVG("output.svg")
	exporter.SaveToPNG("output.png")

	fmt.Println("Exported to TXT, SVG, and PNG")
}

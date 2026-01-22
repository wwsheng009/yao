package dsl

import (
	"testing"
)

// TestParseJSONC tests parsing JSONC (JSON with comments).
func TestParseJSONC(t *testing.T) {
	// JSONC with line comments
	jsoncLineComments := `{
		// This is a line comment
		"name": "Test Layout",  // inline comment
		// Another comment before layout
		"layout": {
			"type": "row",
			"children": [
				{"type": "text", "props": {"content": "Hello"}},
				// Comment between items
				{"type": "text", "props": {"content": "World"}}
			]
		}
	}`

	cfg, err := ParseJSONC([]byte(jsoncLineComments))
	if err != nil {
		t.Fatalf("ParseJSONC with line comments failed: %v", err)
	}

	if cfg.Name != "Test Layout" {
		t.Errorf("Expected name 'Test Layout', got '%s'", cfg.Name)
	}

	if cfg.Layout == nil {
		t.Fatal("Layout is nil")
	}

	if len(cfg.Layout.Children) != 2 {
		t.Errorf("Expected 2 children, got %d", len(cfg.Layout.Children))
	}
}

// TestParseJSONCBlockComments tests parsing JSONC with block comments.
func TestParseJSONCBlockComments(t *testing.T) {
	// JSONC with block comments
	jsoncBlockComments := `{
		/* This is a
		   multi-line
		   block comment */
		"name": "Block Test",
		/* Nested layout with block comment */
		"layout": {
			"type": "row",
			"children": [
				/* First child */
				{"type": "text", "props": {"content": "A"}},
				/* Second child */
				{"type": "text", "props": {"content": "B"}}
			]
		}
	}`

	cfg, err := ParseJSONC([]byte(jsoncBlockComments))
	if err != nil {
		t.Fatalf("ParseJSONC with block comments failed: %v", err)
	}

	if cfg.Name != "Block Test" {
		t.Errorf("Expected name 'Block Test', got '%s'", cfg.Name)
	}

	if len(cfg.Layout.Children) != 2 {
		t.Errorf("Expected 2 children, got %d", len(cfg.Layout.Children))
	}
}

// TestParseJSONCTrailingCommas tests parsing JSONC with trailing commas.
func TestParseJSONCTrailingCommas(t *testing.T) {
	// JSONC with trailing commas (not valid in standard JSON)
	jsoncTrailingCommas := `{
		"name": "Trailing Comma Test",
		"layout": {
			"type": "row",
			"children": [
				{"type": "text", "props": {"content": "First"}},
				{"type": "text", "props": {"content": "Second"}},
			]
		}
	}`

	cfg, err := ParseJSONC([]byte(jsoncTrailingCommas))
	if err != nil {
		t.Fatalf("ParseJSONC with trailing commas failed: %v", err)
	}

	if cfg.Name != "Trailing Comma Test" {
		t.Errorf("Expected name 'Trailing Comma Test', got '%s'", cfg.Name)
	}

	if len(cfg.Layout.Children) != 2 {
		t.Errorf("Expected 2 children, got %d", len(cfg.Layout.Children))
	}
}

// TestParseJSONCMixed tests parsing JSONC with mixed comment styles.
func TestParseJSONCMixed(t *testing.T) {
	jsoncMixed := `{
		// Line comment
		"name": "Mixed Test", /* inline block */
		"layout": {
			"type": "column",
			"direction": "column", // direction property
			"children": [
				/* Child 1 */
				{
					"type": "header",
					"props": {"title": "Title"} // title
				},
				// Child 2
				{
					"type": "text",
					"props": {"content": "Content"}
				},
			]
		}
	}`

	cfg, err := ParseJSONC([]byte(jsoncMixed))
	if err != nil {
		t.Fatalf("ParseJSONC with mixed comments failed: %v", err)
	}

	if cfg.Name != "Mixed Test" {
		t.Errorf("Expected name 'Mixed Test', got '%s'", cfg.Name)
	}

	if cfg.Layout.Direction != "column" {
		t.Errorf("Expected direction 'column', got '%s'", cfg.Layout.Direction)
	}
}

// TestParseJSONCPreservesStrings tests that strings containing comment patterns are preserved.
func TestParseJSONCPreservesStrings(t *testing.T) {
	jsoncStringPatterns := `{
		"name": "String Patterns Test",
		"layout": {
			"type": "row",
			"children": [
				{"type": "text", "props": {"content": "URL: http://example.com"}},
				{"type": "text", "props": {"content": "Path: C://Windows/path"}},
				{"type": "text", "props": {"content": "Ratio: 16/9"}}
			]
		}
	}`

	cfg, err := ParseJSONC([]byte(jsoncStringPatterns))
	if err != nil {
		t.Fatalf("ParseJSONC with string patterns failed: %v", err)
	}

	if len(cfg.Layout.Children) != 3 {
		t.Fatalf("Expected 3 children, got %d", len(cfg.Layout.Children))
	}

	// Check that string content is preserved
	expectedContents := []string{
		"URL: http://example.com",
		"Path: C://Windows/path",
		"Ratio: 16/9",
	}

	for i, child := range cfg.Layout.Children {
		if child.Props == nil {
			t.Errorf("Child %d has nil props", i)
			continue
		}
		content, ok := child.Props["content"].(string)
		if !ok {
			t.Errorf("Child %d content is not a string", i)
			continue
		}
		if content != expectedContents[i] {
			t.Errorf("Child %d: expected content %q, got %q", i, expectedContents[i], content)
		}
	}
}

// TestParseJSONCEmptyComments tests parsing JSONC with empty comments.
func TestParseJSONCEmptyComments(t *testing.T) {
	jsoncEmpty := `{
		//
		"name": "Empty Comments",
		/**/
		"layout": {
			"type": "row"
		}
	}`

	cfg, err := ParseJSONC([]byte(jsoncEmpty))
	if err != nil {
		t.Fatalf("ParseJSONC with empty comments failed: %v", err)
	}

	if cfg.Name != "Empty Comments" {
		t.Errorf("Expected name 'Empty Comments', got '%s'", cfg.Name)
	}
}

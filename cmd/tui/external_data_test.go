package tui

import (
	"encoding/json"
	"strings"
	"testing"

	tuipkg "github.com/yaoapp/yao/tui/tui"
)

// TestParseTUIMetadataArgs tests the parsing of TUI metadata with external data arguments
// This test validates the :: syntax for passing JSON data to TUIs
func TestParseTUIMetadataArgs(t *testing.T) {
	testCases := []struct {
		name         string
		args         []string
		expectedData map[string]interface{}
		expectError  bool
	}{
		{
			name:         "Simple JSON object",
			args:         []string{"test-tui", `::{"title":"External Title","count":42}`},
			expectedData: map[string]interface{}{"title": "External Title", "count": 42.0},
			expectError:  false,
		},
		{
			name: "Nested JSON object",
			args: []string{"test-tui", `::{"user":{"id":123,"name":"John"},"items":["a","b"]}`},
			expectedData: map[string]interface{}{
				"user":  map[string]interface{}{"id": 123.0, "name": "John"},
				"items": []interface{}{"a", "b"},
			},
			expectError: false,
		},
		{
			name:         "Multiple :: arguments",
			args:         []string{"test-tui", `::{"foo":"bar"}`, `::{"baz":"qux"}`},
			expectedData: map[string]interface{}{"foo": "bar", "baz": "qux"},
			expectError:  false,
		},
		{
			name:         "Escaped :: prefix",
			args:         []string{"test-tui", `\::test-string`},
			expectedData: map[string]interface{}{"_args": []interface{}{"::test-string"}},
			expectError:  false,
		},
		{
			name:         "Invalid JSON",
			args:         []string{"test-tui", `::{invalid json}`},
			expectedData: nil,
			expectError:  true,
		},
		{
			name:         "Regular string argument",
			args:         []string{"test-tui", "regular-string"},
			expectedData: map[string]interface{}{"_args": []interface{}{"regular-string"}},
			expectError:  false,
		},
		{
			name: "Array data",
			args: []string{"test-tui", `::{"items":["a","b","c"],"config":{"color":"red"}}`},
			expectedData: map[string]interface{}{
				"items":  []interface{}{"a", "b", "c"},
				"config": map[string]interface{}{"color": "red"},
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test command runner
			var parsedData map[string]interface{}
			var parseError error

			// Parse arguments using the same logic as tui command
			for i, arg := range tc.args {
				if i == 0 {
					continue // Skip tuiID
				}

				if strings.HasPrefix(arg, "::") {
					arg := strings.TrimPrefix(arg, "::")
					var v map[string]interface{}
					err := JSONUnmarshal([]byte(arg), &v)
					if err != nil {
						parseError = err
						break
					}

					if parsedData == nil {
						parsedData = v
					} else {
						for k, val := range v {
							parsedData[k] = val
						}
					}
				} else if strings.HasPrefix(arg, "\\::") {
					arg := "::" + strings.TrimPrefix(arg, "\\::")
					if parsedData == nil {
						parsedData = make(map[string]interface{})
					}
					if argsKey, exists := parsedData["_args"]; !exists {
						parsedData["_args"] = []interface{}{arg}
					} else {
						parsedData["_args"] = append(argsKey.([]interface{}), arg)
					}
				} else {
					if parsedData == nil {
						parsedData = make(map[string]interface{})
					}
					if argsKey, exists := parsedData["_args"]; !exists {
						parsedData["_args"] = []interface{}{arg}
					} else {
						parsedData["_args"] = append(argsKey.([]interface{}), arg)
					}
				}
			}

			if tc.expectError {
				if parseError == nil {
					t.Errorf("Expected parse error but got none")
				}
				return
			}

			if parseError != nil {
				t.Errorf("Unexpected parse error: %v", parseError)
				return
			}

			if !mapsEqual(parsedData, tc.expectedData) {
				t.Errorf("Parsed data mismatch:\n got: %+v\nwant: %+v", parsedData, tc.expectedData)
			}
		})
	}
}

// TestTUIExternalDataMerge tests that external data is correctly merged into TUI config
func TestTUIExternalDataMerge(t *testing.T) {
	cfg := &tuipkg.Config{
		ID:   "test",
		Name: "Test TUI",
		Data: map[string]interface{}{
			"staticTitle": "Static Title",
			"staticCount": 10,
			"items":       []interface{}{"a"},
		},
		Layout: tuipkg.Layout{
			Direction: "vertical",
		},
	}

	externalData := map[string]interface{}{
		"staticTitle":   "External Title", // Should override static data
		"externalField": "External Value",
		"items":         []interface{}{"x", "y", "z"}, // Should override static data
	}

	// Merge external data into config.Data (same logic as tui command)
	if cfg.Data == nil {
		cfg.Data = make(map[string]interface{})
	}
	for k, v := range externalData {
		cfg.Data[k] = v
	}

	// Verify merged data
	model := tuipkg.NewModel(cfg, nil)

	// Check state after merge
	if model.State["staticTitle"] != "External Title" {
		t.Errorf("Expected staticTitle to be 'External Title', got %v", model.State["staticTitle"])
	}
	if model.State["externalField"] != "External Value" {
		t.Errorf("Expected externalField to be 'External Value', got %v", model.State["externalField"])
	}
	if model.State["staticCount"] != 10 {
		t.Errorf("Expected staticCount to be 10, got %v", model.State["staticCount"])
	}

	items, ok := model.State["items"].([]interface{})
	if !ok {
		t.Errorf("Expected items to be []interface{}, got %T", model.State["items"])
	} else if len(items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(items))
	}
}

// mapsEqual compares two maps for equality (supports nested maps and slices)
func mapsEqual(a, b map[string]interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if len(a) != len(b) {
		return false
	}

	for k, v := range a {
		bv, exists := b[k]
		if !exists {
			return false
		}
		if !valuesEqual(v, bv) {
			return false
		}
	}
	return true
}

// valuesEqual compares two values for equality
func valuesEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	switch av := a.(type) {
	case string:
		bv, ok := b.(string)
		return ok && av == bv
	case float64:
		bv, ok := b.(float64)
		return ok && av == bv
	case int:
		bv, ok := b.(int)
		return ok && av == bv
	case bool:
		bv, ok := b.(bool)
		return ok && av == bv
	case []interface{}:
		bv, ok := b.([]interface{})
		if !ok || len(av) != len(bv) {
			return false
		}
		for i := range av {
			if !valuesEqual(av[i], bv[i]) {
				return false
			}
		}
		return true
	case map[string]interface{}:
		bv, ok := b.(map[string]interface{})
		if !ok {
			return false
		}
		return mapsEqual(av, bv)
	default:
		return false
	}
}

// JSONUnmarshal is a helper function to unmarshal JSON
func JSONUnmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

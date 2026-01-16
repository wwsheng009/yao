package tui

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigJSONSerialization(t *testing.T) {
	// Test JSON unmarshaling
	jsonData := `{
		"name": "Test TUI",
		"data": {
			"title": "Hello",
			"count": 0
		},
		"layout": {
			"direction": "vertical",
			"children": [
				{
					"type": "header",
					"props": {
						"title": "{{title}}"
					}
				}
			]
		},
		"bindings": {
			"q": {
				"process": "tui.Quit"
			}
		}
	}`

	var cfg Config
	err := json.Unmarshal([]byte(jsonData), &cfg)
	assert.NoError(t, err)
	assert.Equal(t, "Test TUI", cfg.Name)
	assert.Equal(t, "Hello", cfg.Data["title"])
	assert.Equal(t, "vertical", cfg.Layout.Direction)
	assert.Len(t, cfg.Layout.Children, 1)
	assert.Equal(t, "header", cfg.Layout.Children[0].Type)
	assert.NotNil(t, cfg.Bindings["q"])
	assert.Equal(t, "tui.Quit", cfg.Bindings["q"].Process)
}

func TestConfigMarshaling(t *testing.T) {
	// Test JSON marshaling
	cfg := Config{
		Name: "Test TUI",
		Data: map[string]interface{}{
			"title": "Hello",
		},
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					Type: "header",
					Props: map[string]interface{}{
						"title": "Test",
					},
				},
			},
		},
	}

	data, err := json.Marshal(cfg)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "Test TUI")
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: Config{
				Name: "Valid TUI",
				Layout: Layout{
					Direction: "vertical",
					Children: []Component{
						{Type: "header"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			config: Config{
				Layout: Layout{
					Direction: "vertical",
				},
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "invalid direction",
			config: Config{
				Name: "Test",
				Layout: Layout{
					Direction: "diagonal",
				},
			},
			wantErr: true,
			errMsg:  "invalid layout.direction",
		},
		{
			name: "missing component type",
			config: Config{
				Name: "Test",
				Layout: Layout{
					Direction: "vertical",
					Children: []Component{
						{Props: map[string]interface{}{}},
					},
				},
			},
			wantErr: true,
			errMsg:  "missing 'type' field",
		},
		{
			name: "default direction",
			config: Config{
				Name: "Test",
				Layout: Layout{
					Children: []Component{
						{Type: "header"},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestActionValidation(t *testing.T) {
	tests := []struct {
		name    string
		action  Action
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid process action",
			action: Action{
				Process: "models.user.Get",
				Args:    []interface{}{1},
			},
			wantErr: false,
		},
		{
			name: "valid script action",
			action: Action{
				Script: "scripts/tui/handler",
				Method: "onClick",
			},
			wantErr: false,
		},
		{
			name:    "missing process and script",
			action:  Action{Args: []interface{}{}},
			wantErr: true,
			errMsg:  "must specify either 'process' or 'script'",
		},
		{
			name: "script without method",
			action: Action{
				Script: "scripts/tui/handler",
			},
			wantErr: true,
			errMsg:  "must also specify 'method'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.action.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestErrorMsg(t *testing.T) {
	tests := []struct {
		name     string
		err      ErrorMsg
		expected string
	}{
		{
			name: "error with context",
			err: ErrorMsg{
				Err:     assert.AnError,
				Context: "script execution",
			},
			expected: "[TUI Error in script execution]",
		},
		{
			name: "error without context",
			err: ErrorMsg{
				Err: assert.AnError,
			},
			expected: "[TUI Error]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			assert.Contains(t, result, tt.expected)
		})
	}
}

func TestMessageTypes(t *testing.T) {
	// Test ProcessResultMsg
	prMsg := ProcessResultMsg{
		Target: "users",
		Data:   []interface{}{"user1", "user2"},
	}
	assert.Equal(t, "users", prMsg.Target)
	assert.Len(t, prMsg.Data, 2)

	// Test StateUpdateMsg
	suMsg := StateUpdateMsg{
		Key:   "count",
		Value: 10,
	}
	assert.Equal(t, "count", suMsg.Key)
	assert.Equal(t, 10, suMsg.Value)

	// Test StateBatchUpdateMsg
	sbMsg := StateBatchUpdateMsg{
		Updates: map[string]interface{}{
			"count":   10,
			"message": "updated",
		},
	}
	assert.Len(t, sbMsg.Updates, 2)

	// Test StreamChunkMsg
	scMsg := StreamChunkMsg{
		ID:      "stream1",
		Content: "chunk data",
	}
	assert.Equal(t, "stream1", scMsg.ID)
	assert.Equal(t, "chunk data", scMsg.Content)

	// Test StreamDoneMsg
	sdMsg := StreamDoneMsg{
		ID: "stream1",
	}
	assert.Equal(t, "stream1", sdMsg.ID)
}

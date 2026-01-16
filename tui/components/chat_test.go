package components

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseChatProps(t *testing.T) {
	now := time.Now()
	nowStr := now.Format(time.RFC3339)

	tests := []struct {
		name     string
		props    map[string]interface{}
		expected ChatProps
	}{
		{
			name:  "empty props",
			props: map[string]interface{}{},
			expected: ChatProps{
				ShowInput:      true,
				EnableMarkdown: true,
				GlamourStyle:   "dark",
				InputHeight:    3,
			},
		},
		{
			name: "full props",
			props: map[string]interface{}{
				"messages": []interface{}{
					map[string]interface{}{
						"id":        "1",
						"role":      "user",
						"content":   "Hello",
						"timestamp": nowStr,
					},
					map[string]interface{}{
						"id":        "2",
						"role":      "assistant",
						"content":   "Hi there!",
						"timestamp": nowStr,
					},
				},
				"inputPlaceholder": "Type a message...",
				"showInput":        true,
				"enableMarkdown":   false,
				"glamourStyle":     "light",
				"width":            80,
				"height":           30,
				"inputHeight":      5,
			},
			expected: ChatProps{
				Messages: []Message{
					{
						ID:        "1",
						Role:      "user",
						Content:   "Hello",
						Timestamp: now,
					},
					{
						ID:        "2",
						Role:      "assistant",
						Content:   "Hi there!",
						Timestamp: now,
					},
				},
				InputPlaceholder: "Type a message...",
				ShowInput:        true,
				EnableMarkdown:   false,
				GlamourStyle:     "light",
				Width:            80,
				Height:           30,
				InputHeight:      5,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseChatProps(tt.props)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRenderChat(t *testing.T) {
	props := ChatProps{
		Messages: []Message{
			{
				ID:        "1",
				Role:      "user",
				Content:   "Hello",
				Timestamp: time.Now(),
			},
			{
				ID:        "2",
				Role:      "assistant",
				Content:   "Hi there!",
				Timestamp: time.Now(),
			},
		},
		InputPlaceholder: "Type your message...",
		ShowInput:        true,
		EnableMarkdown:   false,
		Width:            80,
		Height:           30,
	}

	result := RenderChat(props, 80)
	assert.NotEmpty(t, result)
}
package components

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

// Message represents a chat message
type Message struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"` // "user" or "assistant"
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// ChatProps defines the properties for the Chat component
type ChatProps struct {
	// Messages contains the chat history (handled separately to preserve nanosecond precision)
	Messages []Message `json:"-"`

	// InputPlaceholder is the placeholder for the input field
	InputPlaceholder string `json:"inputPlaceholder"`

	// ShowInput determines if the input field is shown
	ShowInput bool `json:"showInput"`

	// EnableMarkdown enables Markdown rendering for messages
	EnableMarkdown bool `json:"enableMarkdown"`

	// GlamourStyle sets the Glamour style for Markdown rendering
	GlamourStyle string `json:"glamourStyle"`

	// Width specifies the chat width (0 for auto)
	Width int `json:"width"`

	// Height specifies the chat height (0 for auto)
	Height int `json:"height"`

	// InputHeight specifies the input field height
	InputHeight int `json:"inputHeight"`

	// Style is the general chat style
	Style lipglossStyleWrapper `json:"style"`

	// UserMessageStyle is the style for user messages
	UserMessageStyle lipglossStyleWrapper `json:"userMessageStyle"`

	// AssistantMessageStyle is the style for assistant messages
	AssistantMessageStyle lipglossStyleWrapper `json:"assistantMessageStyle"`

	// InputStyle is the style for the input field
	InputStyle lipglossStyleWrapper `json:"inputStyle"`

	// TimestampStyle is the style for timestamps
	TimestampStyle lipglossStyleWrapper `json:"timestampStyle"`
}

// ChatModel represents a chat model for interactive chats
type ChatModel struct {
	Viewport    viewport.Model
	TextInput   textarea.Model
	props       ChatProps
	messages    []Message
	historyText string
}

// updateHistoryText updates the history text based on current messages
func (cm *ChatModel) updateHistoryText() {
	var historyText strings.Builder
	for _, msg := range cm.messages {
		// Format message based on role
		var msgStyle lipgloss.Style
		var prefix string

		switch msg.Role {
		case "user":
			msgStyle = cm.props.UserMessageStyle.GetStyle()
			prefix = "👤 You: "
		case "assistant":
			msgStyle = cm.props.AssistantMessageStyle.GetStyle()
			prefix = "🤖 Assistant: "
		default:
			msgStyle = cm.props.UserMessageStyle.GetStyle()
			prefix = fmt.Sprintf("%s: ", msg.Role)
		}

		// Apply Markdown rendering if enabled
		content := msg.Content
		if cm.props.EnableMarkdown {
			renderer, err := glamour.NewTermRenderer(
				glamour.WithStandardStyle(cm.props.GlamourStyle),
				glamour.WithWordWrap(0),
			)
			if err == nil {
				rendered, err := renderer.Render(content)
				if err == nil {
					content = rendered
				}
			}
		}

		// Format message
		messageText := prefix + content
		tsStyle := cm.props.TimestampStyle.GetStyle()
		if tsStyle.GetBackground() != lipgloss.Color("") || tsStyle.GetForeground() != lipgloss.Color("") {
			ts := msg.Timestamp.Format("15:04")
			timestamp := tsStyle.Render(fmt.Sprintf("[%s]", ts))
			messageText = timestamp + " " + messageText
		}

		historyText.WriteString(msgStyle.Render(messageText))
		historyText.WriteString("\n\n")
	}

	cm.historyText = historyText.String()
	cm.Viewport.SetContent(cm.historyText)
}

// RenderChat renders a chat component
func RenderChat(props ChatProps, width int) string {
	var sb strings.Builder

	// Prepare chat history text
	var historyText strings.Builder
	for _, msg := range props.Messages {
		// Format message based on role
		var msgStyle lipgloss.Style
		var prefix string

		switch msg.Role {
		case "user":
			msgStyle = props.UserMessageStyle.GetStyle()
			prefix = "👤 You: "
		case "assistant":
			msgStyle = props.AssistantMessageStyle.GetStyle()
			prefix = "🤖 Assistant: "
		default:
			msgStyle = props.UserMessageStyle.GetStyle()
			prefix = fmt.Sprintf("%s: ", msg.Role)
		}

		// Apply Markdown rendering if enabled
		content := msg.Content
		if props.EnableMarkdown {
			renderer, err := glamour.NewTermRenderer(
				glamour.WithStandardStyle(props.GlamourStyle),
				glamour.WithWordWrap(0),
			)
			if err == nil {
				rendered, err := renderer.Render(content)
				if err == nil {
					content = rendered
				}
			}
		}

		// Format message
		messageText := prefix + content
		tsStyle := props.TimestampStyle.GetStyle()
		if tsStyle.GetBackground() != lipgloss.Color("") || tsStyle.GetForeground() != lipgloss.Color("") {
			ts := msg.Timestamp.Format("15:04")
			timestamp := tsStyle.Render(fmt.Sprintf("[%s]", ts))
			messageText = timestamp + " " + messageText
		}

		historyText.WriteString(msgStyle.Render(messageText))
		historyText.WriteString("\n\n")
	}

	// Create viewport for chat history
	chatWidth := props.Width
	if chatWidth <= 0 && width > 0 {
		chatWidth = width
	}

	chatHeight := props.Height
	if chatHeight <= 0 {
		chatHeight = 15 // Default height
	}

	// Reserve space for input if shown
	if props.ShowInput {
		inputHeight := props.InputHeight
		if inputHeight <= 0 {
			inputHeight = 3 // Default input height
		}
		chatHeight -= inputHeight + 1 // +1 for spacing
	}

	vp := viewport.New(chatWidth, chatHeight)
	vp.SetContent(historyText.String())

	// Add chat history to output
	sb.WriteString(vp.View())

	// Add input field if shown
	if props.ShowInput {
		sb.WriteString("\n")

		// Create a simple representation of the input field
		inputStyle := props.InputStyle.GetStyle()
		placeholder := props.InputPlaceholder
		if placeholder == "" {
			placeholder = "Type your message..."
		}

		inputText := fmt.Sprintf("> %s", placeholder)
		sb.WriteString(inputStyle.Render(inputText))
	}

	return sb.String()
}

// ParseChatProps converts a generic props map to ChatProps using JSON unmarshaling
func ParseChatProps(props map[string]interface{}) ChatProps {
	// Set defaults
	cp := ChatProps{
		ShowInput:      true,
		EnableMarkdown: true,
		GlamourStyle:   "dark",
		InputHeight:    3,
	}

	// Handle Messages separately as it needs special processing
	if msgs, ok := props["messages"].([]interface{}); ok {
		cp.Messages = make([]Message, 0, len(msgs))
		for _, msgIntf := range msgs {
			if msgMap, ok := msgIntf.(map[string]interface{}); ok {
				msg := Message{}

				if id, ok := msgMap["id"].(string); ok {
					msg.ID = id
				}

				if role, ok := msgMap["role"].(string); ok {
					msg.Role = role
				}

				if content, ok := msgMap["content"].(string); ok {
					msg.Content = content
				}

				if tsStr, ok := msgMap["timestamp"].(string); ok {
					if ts, err := time.Parse(time.RFC3339, tsStr); err == nil {
						msg.Timestamp = ts
					} else {
						msg.Timestamp = time.Now()
					}
				} else {
					msg.Timestamp = time.Now()
				}

				cp.Messages = append(cp.Messages, msg)
			}
		}
	}

	// Unmarshal remaining properties
	if dataBytes, err := json.Marshal(props); err == nil {
		_ = json.Unmarshal(dataBytes, &cp)
	}

	return cp
}

// HandleChatUpdate handles updates for chat components
// This is used when the chat is interactive
func HandleChatUpdate(msg tea.Msg, chatModel *ChatModel) (ChatModel, tea.Cmd) {
	if chatModel == nil {
		return *chatModel, nil
	}

	var cmd tea.Cmd
	var cmds []tea.Cmd

	// Update viewport
	chatModel.Viewport, cmd = chatModel.Viewport.Update(msg)
	cmds = append(cmds, cmd)

	// Update text input if it exists
	if chatModel.TextInput.Focused() {
		chatModel.TextInput, cmd = chatModel.TextInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return *chatModel, tea.Batch(cmds...)
}

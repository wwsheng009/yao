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
	"github.com/yaoapp/yao/tui/core"
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
	id          string // Unique identifier for this instance
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

// NewChatModel creates a new ChatModel from ChatProps
func NewChatModel(props ChatProps, id string) ChatModel {
	cm := ChatModel{
		props:    props,
		messages: props.Messages,
		id:       id,
	}

	// Initialize viewport
	viewWidth := props.Width
	if viewWidth <= 0 {
		viewWidth = 80
	}

	viewHeight := props.Height
	if viewHeight <= 0 {
		viewHeight = 20
	}

	// Reserve space for input if shown
	if props.ShowInput {
		inputHeight := props.InputHeight
		if inputHeight <= 0 {
			inputHeight = 3
		}
		viewHeight -= inputHeight + 1
	}

	cm.Viewport = viewport.New(viewWidth, viewHeight)

	// Initialize text input
	cm.TextInput = textarea.New()
	cm.TextInput.Placeholder = props.InputPlaceholder
	if props.InputPlaceholder == "" {
		cm.TextInput.Placeholder = "Type your message..."
	}
	cm.TextInput.ShowLineNumbers = false
	cm.TextInput.Focus()

	// Update history text
	cm.updateHistoryText()

	return cm
}

// Init initializes the chat model
func (m *ChatModel) Init() tea.Cmd {
	return nil
}

// View returns the string representation of the chat
func (m *ChatModel) View() string {
	var sb strings.Builder

	// Add viewport (chat history)
	sb.WriteString(m.Viewport.View())

	// Add input field if shown
	if m.props.ShowInput {
		sb.WriteString("\n")
		sb.WriteString(m.TextInput.View())
	}

	return sb.String()
}

// GetID returns the unique identifier for this component instance
func (m *ChatModel) GetID() string {
	return m.id
}

// GetComponentType returns the component type
func (m *ChatModel) GetComponentType() string {
	return "chat"
}

func (m *ChatModel) Render(config core.RenderConfig) (string, error) {
	// Parse configuration data
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("ChatModel: invalid data type")
	}

	// Parse chat properties
	props := ParseChatProps(propsMap)

	// Update component properties
	m.props = props

	// Return rendered view
	return m.View(), nil
}

// AddMessage adds a new message to the chat
func (m *ChatModel) AddMessage(role, content string) {
	msg := Message{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	}
	m.messages = append(m.messages, msg)
	m.updateHistoryText()
	// Scroll to bottom to show new message
	m.Viewport.GotoBottom()
}

// GetMessages returns all messages
func (m *ChatModel) GetMessages() []Message {
	return m.messages
}

// ClearMessages clears all messages
func (m *ChatModel) ClearMessages() {
	m.messages = []Message{}
	m.updateHistoryText()
}

// ChatComponentWrapper wraps ChatModel to implement ComponentInterface properly
type ChatComponentWrapper struct {
	model *ChatModel
}

// NewChatComponentWrapper creates a wrapper that implements ComponentInterface
func NewChatComponentWrapper(chatModel *ChatModel) *ChatComponentWrapper {
	return &ChatComponentWrapper{
		model: chatModel,
	}
}

func (w *ChatComponentWrapper) Init() tea.Cmd {
	return w.model.Init()
}

func (w *ChatComponentWrapper) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
	// Handle targeted messages first
	switch msg := msg.(type) {
	case core.TargetedMsg:
		// Check if this message is targeted to this component
		if msg.TargetID == w.model.id {
			return w.UpdateMsg(msg.InnerMsg)
		}
		return w, nil, core.Ignored

	case tea.KeyMsg:
		var cmds []tea.Cmd

		switch msg.Type {
		case tea.KeyEsc:
			// Blur the input when ESC is pressed
			w.model.TextInput.Blur()
			// Publish focus changed event
			cmds = append(cmds, core.PublishEvent(w.model.id, core.EventInputFocusChanged, map[string]interface{}{
				"focused": false,
			}))
			if len(cmds) > 0 {
				return w, tea.Batch(cmds...), core.Handled
			}
			return w, nil, core.Handled

		case tea.KeyEnter:
			// Check if shift is pressed (for multi-line input)
			if msg.String() == "shift+enter" || msg.Alt {
				// Allow multi-line input
				var cmd tea.Cmd
				w.model.TextInput, cmd = w.model.TextInput.Update(msg)
				return w, cmd, core.Handled
			}

			// Get input text
			inputText := w.model.TextInput.Value()
			if inputText == "" {
				// Empty input, ignore
				return w, nil, core.Handled
			}

			// Clear input
			w.model.TextInput.Reset()

			// Add user message
			w.model.AddMessage("user", inputText)

			// Publish message sent event
			cmds = append(cmds, core.PublishEvent(w.model.id, core.EventChatMessageSent, map[string]interface{}{
				"role":    "user",
				"content": inputText,
			}))

			// Publish input enter pressed event
			cmds = append(cmds, core.PublishEvent(w.model.id, core.EventInputEnterPressed, map[string]interface{}{
				"value": inputText,
			}))

			if len(cmds) > 0 {
				return w, tea.Batch(cmds...), core.Handled
			}
			return w, nil, core.Handled

		case tea.KeyCtrlC:
			// Let Ctrl+C pass through for quit handling
			return w, nil, core.Ignored
		}

		// For other key messages, update the text input model
		oldValue := w.model.TextInput.Value()
		var cmd tea.Cmd
		w.model.TextInput, cmd = w.model.TextInput.Update(msg)

		// Check if value changed
		newValue := w.model.TextInput.Value()
		if oldValue != newValue {
			cmds = append(cmds, cmd)
			// Publish value changed event
			cmds = append(cmds, core.PublishEvent(w.model.id, core.EventInputValueChanged, map[string]interface{}{
				"oldValue": oldValue,
				"newValue": newValue,
			}))
			return w, tea.Batch(cmds...), core.Handled
		}
		return w, cmd, core.Handled

	case core.ActionMsg:
		// Handle internal action messages
		switch msg.Action {
		case core.EventChatMessageReceived:
			// Add received message to chat
			if data, ok := msg.Data.(map[string]interface{}); ok {
				if role, ok := data["role"].(string); ok {
					if content, ok := data["content"].(string); ok {
						w.model.AddMessage(role, content)
						return w, nil, core.Handled
					}
				}
			}
		}

		// Default: ignore action message
		return w, nil, core.Ignored
	}

	// For other messages, update viewport and text input
	var cmd tea.Cmd
	var cmds []tea.Cmd
	oldValue := ""

	if w.model.TextInput.Focused() {
		oldValue = w.model.TextInput.Value()
	}

	updatedModel, chatCmd := HandleChatUpdate(msg, w.model)
	w.model = &updatedModel
	cmds = append(cmds, chatCmd)

	// Check if text input value changed
	if w.model.TextInput.Focused() {
		newValue := w.model.TextInput.Value()
		if oldValue != newValue {
			cmds = append(cmds, core.PublishEvent(w.model.id, core.EventInputValueChanged, map[string]interface{}{
				"oldValue": oldValue,
				"newValue": newValue,
			}))
		}
	}

	if len(cmds) > 0 {
		return w, tea.Batch(cmds...), core.Handled
	}
	return w, cmd, core.Handled
}

func (w *ChatComponentWrapper) View() string {
	return w.model.View()
}

func (w *ChatComponentWrapper) GetID() string {
	return w.model.id
}

// GetValue returns the current input value
func (w *ChatComponentWrapper) GetValue() string {
	return w.model.TextInput.Value()
}

// SetValue sets the input value
func (w *ChatComponentWrapper) SetValue(value string) {
	w.model.TextInput.SetValue(value)
}

// SetFocus sets or removes focus from the chat component
func (m *ChatModel) SetFocus(focus bool) {
	if focus {
		m.TextInput.Focus()
	} else {
		m.TextInput.Blur()
	}
}

func (w *ChatComponentWrapper) SetFocus(focus bool) {
	w.model.SetFocus(focus)
}

func (w *ChatComponentWrapper) GetComponentType() string {
	return "chat"
}

func (w *ChatComponentWrapper) Render(config core.RenderConfig) (string, error) {
	return w.model.Render(config)
}

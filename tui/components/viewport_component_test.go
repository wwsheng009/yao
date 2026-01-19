package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yaoapp/yao/tui/core"
)

func TestViewportComponentWrapper_UpdateMsg_TargetedMsg(t *testing.T) {
	props := ViewportProps{
		Content:      "Test content",
		ShowScrollbar: true,
		Width:        80,
		Height:       20,
	}
	wrapper := NewViewportComponentWrapper(props, "test-viewport")

	// Test targeted message to this component
	innerMsg := tea.KeyMsg{Type: tea.KeyUp}
	targetedMsg := core.TargetedMsg{
		TargetID: "test-viewport",
		InnerMsg: innerMsg,
	}
	comp, _, response := wrapper.UpdateMsg(targetedMsg)
	if response != core.Handled {
		t.Errorf("Expected Handled response, got %v", response)
	}
	if comp != wrapper {
		t.Error("Expected wrapper to be returned")
	}

	// Test targeted message to different component
	targetedMsg = core.TargetedMsg{
		TargetID: "other-component",
		InnerMsg: innerMsg,
	}
	comp, _, response = wrapper.UpdateMsg(targetedMsg)
	if response != core.Ignored {
		t.Errorf("Expected Ignored response, got %v", response)
	}
	if comp != wrapper {
		t.Error("Expected wrapper to be returned")
	}
}

func TestViewportComponentWrapper_UpdateMsg_ScrollKeys(t *testing.T) {
	props := ViewportProps{
		Content: "Line 1\nLine 2\nLine 3\nLine 4\nLine 5\nLine 6\nLine 7\nLine 8\nLine 9\nLine 10\n" +
			"Line 11\nLine 12\nLine 13\nLine 14\nLine 15\nLine 16\nLine 17\nLine 18\nLine 19\nLine 20\n",
		ShowScrollbar: true,
		Width:        50,
		Height:       10,
	}
	wrapper := NewViewportComponentWrapper(props, "test-viewport")

	testCases := []struct {
		name     string
		msg      tea.KeyMsg
		expected core.Response
	}{
		{"Scroll Up", tea.KeyMsg{Type: tea.KeyUp}, core.Handled},
		{"Scroll Down", tea.KeyMsg{Type: tea.KeyDown}, core.Handled},
		{"Page Up", tea.KeyMsg{Type: tea.KeyPgUp}, core.Handled},
		{"Page Down", tea.KeyMsg{Type: tea.KeyPgDown}, core.Handled},
		{"Home", tea.KeyMsg{Type: tea.KeyHome}, core.Handled},
		{"End", tea.KeyMsg{Type: tea.KeyEnd}, core.Handled},
		{"Escape", tea.KeyMsg{Type: tea.KeyEsc}, core.Ignored},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, response := wrapper.UpdateMsg(tc.msg)
			if response != tc.expected {
				t.Errorf("Expected %v response, got %v", tc.expected, response)
			}
		})
	}
}

func TestViewportComponentWrapper_UpdateMsg_ActionMsg(t *testing.T) {
	props := ViewportProps{
		Content:      "Original content",
		ShowScrollbar: true,
	}
	wrapper := NewViewportComponentWrapper(props, "test-viewport")

	// Test EventDataLoaded action
	actionMsg := core.ActionMsg{
		ID:     "test-viewport",
		Action: core.EventDataLoaded,
		Data: map[string]interface{}{
			"content": "New loaded content",
		},
	}
	_, _, response := wrapper.UpdateMsg(actionMsg)
	if response != core.Handled {
		t.Errorf("Expected Handled response, got %v", response)
	}

	// Verify content was updated
	if wrapper.props.Content != "Original content" {
		t.Error("Expected props.Content to remain unchanged")
	}
	// Note: The actual viewport content is updated via SetContent

	// Test EventDataRefreshed action
	actionMsg = core.ActionMsg{
		ID:     "test-viewport",
		Action: core.EventDataRefreshed,
		Data: map[string]interface{}{
			"content": "Refreshed content",
		},
	}
	_, _, response = wrapper.UpdateMsg(actionMsg)
	if response != core.Handled {
		t.Errorf("Expected Handled response, got %v", response)
	}
}

func TestViewportComponentWrapper_UpdateMsg_WindowSize(t *testing.T) {
	props := ViewportProps{
		Content:      "Test content",
		ShowScrollbar: true,
		Width:        0, // Auto
		Height:       0, // Auto
	}
	wrapper := NewViewportComponentWrapper(props, "test-viewport")

	// Test window resize
	windowSizeMsg := tea.WindowSizeMsg{
		Width:  100,
		Height: 30,
	}
	_, _, response := wrapper.UpdateMsg(windowSizeMsg)
	if response != core.Handled {
		t.Errorf("Expected Handled response, got %v", response)
	}

	// Verify viewport dimensions were updated
	if wrapper.model.Width != 100 {
		t.Errorf("Expected width 100, got %d", wrapper.model.Width)
	}
	if wrapper.model.Height != 30 {
		t.Errorf("Expected height 30, got %d", wrapper.model.Height)
	}
}

func TestViewportComponentWrapper_SetContent(t *testing.T) {
	props := ViewportProps{
		Content:      "Original content",
		ShowScrollbar: true,
		EnableGlamour: false,
	}
	wrapper := NewViewportComponentWrapper(props, "test-viewport")

	// Update content
	newContent := "Updated content line 1\nUpdated content line 2\nUpdated content line 3"
	wrapper.SetContent(newContent)

	// Content is updated internally but we can't directly access it
	// Just verify the method doesn't panic
	view := wrapper.View()
	if view == "" {
		t.Error("Expected non-empty view")
	}
}

func TestViewportComponentWrapper_GotoTop(t *testing.T) {
	props := ViewportProps{
		Content: "Line 1\nLine 2\nLine 3\nLine 4\nLine 5\nLine 6\nLine 7\nLine 8\nLine 9\nLine 10\n" +
			"Line 11\nLine 12\nLine 13\nLine 14\nLine 15\nLine 16\nLine 17\nLine 18\nLine 19\nLine 20\n",
		Width:        50,
		Height:       10,
		ShowScrollbar: true,
	}
	wrapper := NewViewportComponentWrapper(props, "test-viewport")

	// Scroll to bottom first
	wrapper.GotoBottom()

	// Scroll to top
	wrapper.GotoTop()

	// Verify viewport is at top
	if wrapper.model.YOffset != 0 {
		t.Errorf("Expected YOffset 0 after GotoTop, got %d", wrapper.model.YOffset)
	}
}

func TestViewportComponentWrapper_GotoBottom(t *testing.T) {
	props := ViewportProps{
		Content: "Line 1\nLine 2\nLine 3\nLine 4\nLine 5\nLine 6\nLine 7\nLine 8\nLine 9\nLine 10\n" +
			"Line 11\nLine 12\nLine 13\nLine 14\nLine 15\nLine 16\nLine 17\nLine 18\nLine 19\nLine 20\n",
		Width:        50,
		Height:       10,
		ShowScrollbar: true,
	}
	wrapper := NewViewportComponentWrapper(props, "test-viewport")

	// Scroll to bottom
	wrapper.GotoBottom()

	// Verify viewport is at or near bottom (should be > 0)
	if wrapper.model.YOffset == 0 {
		t.Error("Expected YOffset > 0 after GotoBottom")
	}
}

func TestViewportComponentWrapper_SetFocus(t *testing.T) {
	props := ViewportProps{
		Content:      "Test content",
		ShowScrollbar: true,
	}
	wrapper := NewViewportComponentWrapper(props, "test-viewport")

	// SetFocus should not panic for viewport
	wrapper.SetFocus(true)
	wrapper.SetFocus(false)

	// Viewport doesn't have visual focus state, just verify no panic
}

func TestNewViewportModel(t *testing.T) {
	props := ViewportProps{
		Content:      "Test content",
		ShowScrollbar: true,
		EnableGlamour: false,
		AutoScroll:    true,
		Width:        80,
		Height:       20,
	}
	viewportModel := NewViewportModel(props, "test-viewport")

	// Verify properties
	if viewportModel.props.Content != "Test content" {
		t.Errorf("Expected content 'Test content', got '%s'", viewportModel.props.Content)
	}
	if viewportModel.props.Width != 80 {
		t.Errorf("Expected width 80, got %d", viewportModel.props.Width)
	}
	if viewportModel.props.Height != 20 {
		t.Errorf("Expected height 20, got %d", viewportModel.props.Height)
	}
	if viewportModel.props.ShowScrollbar != true {
		t.Error("Expected ShowScrollbar to be true")
	}
	if viewportModel.props.AutoScroll != true {
		t.Error("Expected AutoScroll to be true")
	}
}

func TestNewViewportModel_AutoDimensions(t *testing.T) {
	props := ViewportProps{
		Content:      "Line 1\nLine 2\nLine 3\nLine 4\nLine 5",
		ShowScrollbar: true,
		Width:        0, // Auto
		Height:       0, // Auto
	}
	viewportModel := NewViewportModel(props, "test-viewport")

	// Verify auto-calculated dimensions
	if viewportModel.Model.Width == 0 {
		t.Error("Expected width to be calculated")
	}
	// Height should be line count + 2 = 5 + 2 = 7
	if viewportModel.Model.Height != 7 {
		t.Errorf("Expected height 7, got %d", viewportModel.Model.Height)
	}
}

func TestViewportModel_Init(t *testing.T) {
	props := ViewportProps{Content: "Test content"}
	viewportModel := NewViewportModel(props, "test-viewport")

	cmd := viewportModel.Init()
	if cmd != nil {
		t.Error("Expected nil command from Init")
	}
}

func TestViewportModel_View(t *testing.T) {
	props := ViewportProps{Content: "Test content"}
	viewportModel := NewViewportModel(props, "test-viewport")

	view := viewportModel.View()
	if view == "" {
		t.Error("Expected non-empty view")
	}
}

func TestViewportModel_SetContent(t *testing.T) {
	props := ViewportProps{
		Content:      "Original content",
		EnableGlamour: false,
		AutoScroll:    true,
	}
	viewportModel := NewViewportModel(props, "test-viewport")

	// Update content
	newContent := "New content line 1\nNew content line 2\nNew content line 3\n" +
		"New content line 4\nNew content line 5\nNew content line 6\n" +
		"New content line 7\nNew content line 8\nNew content line 9\n" +
		"New content line 10\nNew content line 11\nNew content line 12"
	viewportModel.SetContent(newContent)

	// Verify auto-scroll
	if viewportModel.Model.YOffset == 0 {
		t.Error("Expected YOffset > 0 after SetContent with AutoScroll enabled")
	}
}

func TestViewportModel_GotoTop(t *testing.T) {
	props := ViewportProps{
		Content:      "Line 1\nLine 2\nLine 3\nLine 4\nLine 5\nLine 6\nLine 7\nLine 8\nLine 9\nLine 10\n" +
			"Line 11\nLine 12\nLine 13\nLine 14\nLine 15\nLine 16\nLine 17\nLine 18\nLine 19\nLine 20\n",
		Height:       10,
	}
	viewportModel := NewViewportModel(props, "test-viewport")

	// Scroll to bottom first
	viewportModel.GotoBottom()

	// Scroll to top
	viewportModel.GotoTop()

	// Verify viewport is at top
	if viewportModel.Model.YOffset != 0 {
		t.Errorf("Expected YOffset 0 after GotoTop, got %d", viewportModel.Model.YOffset)
	}
}

func TestViewportModel_GotoBottom(t *testing.T) {
	props := ViewportProps{
		Content:      "Line 1\nLine 2\nLine 3\nLine 4\nLine 5\nLine 6\nLine 7\nLine 8\nLine 9\nLine 10\n" +
			"Line 11\nLine 12\nLine 13\nLine 14\nLine 15\nLine 16\nLine 17\nLine 18\nLine 19\nLine 20\n",
		Height:       10,
	}
	viewportModel := NewViewportModel(props, "test-viewport")

	// Scroll to bottom
	viewportModel.GotoBottom()

	// Verify viewport is at bottom
	if viewportModel.Model.YOffset == 0 {
		t.Error("Expected YOffset > 0 after GotoBottom")
	}
}

func TestViewportComponentWrapper_GetID(t *testing.T) {
	props := ViewportProps{Content: "Test content"}
	wrapper := NewViewportComponentWrapper(props, "test-id-456")

	if wrapper.GetID() != "test-id-456" {
		t.Errorf("Expected id 'test-id-456', got '%s'", wrapper.GetID())
	}
}

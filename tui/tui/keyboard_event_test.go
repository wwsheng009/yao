package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

// TestKeyboardEventsInRuntimeMode 测试 Runtime 模式下的键盘事件
func TestKeyboardEventsInRuntimeMode(t *testing.T) {
	useRuntime := true
	cfg := &Config{
		Name:      "Keyboard Test",
		UseRuntime: &useRuntime,
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					ID:   "input1",
					Type: "input",
					Props: map[string]interface{}{
						"placeholder": "First input",
					},
				},
				{
					ID:   "input2",
					Type: "input",
					Props: map[string]interface{}{
						"placeholder": "Second input",
					},
				},
			},
		},
	}

	model := NewModel(cfg, nil)

	// 初始化模型
	cmd := model.Init()
	assert.NotNil(t, cmd, "Init should return commands")

	// 处理初始化命令
	if cmd != nil {
		msg := cmd()
		var m tea.Model = model
		m, _ = m.Update(msg)
		model = m.(*Model)
	}

	t.Logf("After Init - CurrentFocus: %q", model.CurrentFocus)
	t.Logf("Components in map: %d", len(model.Components))
	t.Logf("UseRuntime: %v", model.UseRuntime)

	// 检查组件是否正确注册
	for id, comp := range model.Components {
		t.Logf("Component: %s, Type: %s", id, comp.Type)
	}

	// 模拟 Tab 键
	tabMsg := tea.KeyMsg{Type: tea.KeyTab}
	var m1 tea.Model = model
	m1, tabCmd := m1.Update(tabMsg)
	model = m1.(*Model)

	t.Logf("After Tab - CurrentFocus: %q, Cmd: %v", model.CurrentFocus, tabCmd != nil)

	// 处理 Tab 命令
	if tabCmd != nil {
		msg := tabCmd()
		var m2 tea.Model = model
		m2, _ = m2.Update(msg)
		model = m2.(*Model)
		t.Logf("After Tab command - CurrentFocus: %q", model.CurrentFocus)
	}

	// 模拟字符输入
	charMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	var m3 tea.Model = model
	m3, charCmd := m3.Update(charMsg)
	model = m3.(*Model)

	t.Logf("After 'a' key - CurrentFocus: %q, Cmd: %v", model.CurrentFocus, charCmd != nil)

	// 处理字符命令
	if charCmd != nil {
		msg := charCmd()
		var m4 tea.Model = model
		m4, _ = m4.Update(msg)
		model = m4.(*Model)
	}

	// 验证输入值
	if model.CurrentFocus != "" {
		if comp, exists := model.Components[model.CurrentFocus]; exists {
			if wrapper, ok := comp.Instance.(interface{ GetValue() string }); ok {
				value := wrapper.GetValue()
				t.Logf("Input value: %q", value)
			}
		}
	}
}

// TestKeyboardEventsLegacyMode 测试 Legacy 模式下的键盘事件
func TestKeyboardEventsLegacyMode(t *testing.T) {
	useRuntime := false
	cfg := &Config{
		Name:      "Keyboard Test (Legacy)",
		UseRuntime: &useRuntime,
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					ID:   "input1",
					Type: "input",
					Props: map[string]interface{}{
						"placeholder": "First input",
					},
				},
				{
					ID:   "input2",
					Type: "input",
					Props: map[string]interface{}{
						"placeholder": "Second input",
					},
				},
			},
		},
	}

	model := NewModel(cfg, nil)

	// 初始化模型
	cmd := model.Init()
	assert.NotNil(t, cmd, "Init should return commands")

	// 处理初始化命令
	if cmd != nil {
		msg := cmd()
		var m tea.Model = model
		m, _ = m.Update(msg)
		model = m.(*Model)
	}

	t.Logf("After Init - CurrentFocus: %q", model.CurrentFocus)
	t.Logf("Components in map: %d", len(model.Components))
	t.Logf("UseRuntime: %v", model.UseRuntime)

	// 模拟 Tab 键
	tabMsg := tea.KeyMsg{Type: tea.KeyTab}
	var m1 tea.Model = model
	m1, tabCmd := m1.Update(tabMsg)
	model = m1.(*Model)

	t.Logf("After Tab - CurrentFocus: %q, Cmd: %v", model.CurrentFocus, tabCmd != nil)

	// 处理 Tab 命令
	if tabCmd != nil {
		msg := tabCmd()
		var m2 tea.Model = model
		m2, _ = m2.Update(msg)
		model = m2.(*Model)
		t.Logf("After Tab command - CurrentFocus: %q", model.CurrentFocus)
	}
}

// TestFocusableComponentsDetection 测试可聚焦组件检测
func TestFocusableComponentsDetection(t *testing.T) {
	useRuntime := true
	cfg := &Config{
		Name:      "Focusable Detection Test",
		UseRuntime: &useRuntime,
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					ID:   "input1",
					Type: "input",
					Props: map[string]interface{}{
						"placeholder": "Input 1",
					},
				},
				{
					ID:   "text1",
					Type: "text",
					Props: map[string]interface{}{
						"content": "Some text",
					},
				},
				{
					ID:   "input2",
					Type: "input",
					Props: map[string]interface{}{
						"placeholder": "Input 2",
					},
				},
			},
		},
	}

	model := NewModel(cfg, nil)

	// 初始化
	cmd := model.Init()
	if cmd != nil {
		msg := cmd()
		var m tea.Model = model
		m, _ = m.Update(msg)
		model = m.(*Model)
	}

	// 检查可聚焦组件
	focusableIDs := model.getFocusableComponentIDs()
	t.Logf("Focusable components: %v", focusableIDs)
	t.Logf("Components in map: %d", len(model.Components))

	// 应该只包含 input 组件，不包含 text 组件
	assert.NotEmpty(t, focusableIDs, "Should have focusable components")
}

// TestMessageHandlerRegistration 测试消息处理器注册
func TestMessageHandlerRegistration(t *testing.T) {
	cfg := &Config{
		Name: "Message Handler Test",
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					ID:   "input1",
					Type: "input",
				},
			},
		},
	}

	model := NewModel(cfg, nil)

	// 检查消息处理器是否正确注册
	t.Logf("Registered message handlers:")
	for msgType := range model.MessageHandlers {
		t.Logf("  - %s", msgType)
	}

	// 验证 tea.KeyMsg 处理器存在
	_, exists := model.MessageHandlers["tea.KeyMsg"]
	assert.True(t, exists, "tea.KeyMsg handler should be registered")
}

// TestKeyPressFlow 测试按键处理的完整流程
func TestKeyPressFlow(t *testing.T) {
	cfg := &Config{
		Name: "Key Press Flow Test",
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					ID:   "input1",
					Type: "input",
					Props: map[string]interface{}{
						"placeholder": "Type here",
					},
				},
			},
		},
	}

	model := NewModel(cfg, nil)

	// 初始化
	cmd := model.Init()
	if cmd != nil {
		msg := cmd()
		var m tea.Model = model
		m, _ = m.Update(msg)
		model = m.(*Model)
	}

	t.Logf("=== Testing Key Press Flow ===")
	t.Logf("1. Initial state - CurrentFocus: %q", model.CurrentFocus)
	t.Logf("2. Components: %d", len(model.Components))

	// 测试字符输入
	testKeys := []tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune{'H'}},
		{Type: tea.KeyRunes, Runes: []rune{'e'}},
		{Type: tea.KeyRunes, Runes: []rune{'l'}},
		{Type: tea.KeyRunes, Runes: []rune{'l'}},
		{Type: tea.KeyRunes, Runes: []rune{'o'}},
	}

	for i, keyMsg := range testKeys {
		var m tea.Model = model
		m, keyCmd := m.Update(keyMsg)

		// 处理命令
		if keyCmd != nil {
			msg := keyCmd()
			var m2 tea.Model = m
			var cmd tea.Cmd
		m2, cmd = m2.Update(msg)
		_ = cmd
			model = m2.(*Model)
		} else {
			model = m.(*Model)
		}

		t.Logf("3. After key %d (%c) - CurrentFocus: %q", i+1, keyMsg.Runes[0], model.CurrentFocus)
	}

	// 获取最终输入值
	if model.CurrentFocus != "" {
		if comp, exists := model.Components[model.CurrentFocus]; exists {
			if wrapper, ok := comp.Instance.(interface{ GetValue() string }); ok {
				value := wrapper.GetValue()
				t.Logf("4. Final input value: %q", value)
				assert.Equal(t, "Hello", value, "Input should contain 'Hello'")
			}
		}
	}
}

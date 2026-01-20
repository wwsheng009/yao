package core

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

// MockInput 模拟输入组件
type MockInput struct {
	value   string
	focused bool
}

func (m *MockInput) GetValue() string {
	return m.value
}

func (m *MockInput) Focused() bool {
	return m.focused
}

// MockList 模拟列表组件
type MockList struct {
	index    int
	selected interface{}
}

func (m *MockList) Index() int {
	return m.index
}

func (m *MockList) SelectedItem() interface{} {
	return m.selected
}

// MockComponent 模拟交互组件行为
type MockComponent struct {
	id         string
	mockInput  *MockInput
	mockList   *MockList
	hasFocus   bool
	specialKey func(keyMsg tea.KeyMsg) (tea.Cmd, Response, bool)
}

// GetFocus implements [ComponentInterface].
func (m *MockComponent) GetFocus() bool {
	return m.hasFocus
}

func (m *MockComponent) GetID() string {
	return m.id
}

func (m *MockComponent) CaptureState() map[string]interface{} {
	state := map[string]interface{}{}
	if m.mockInput != nil {
		state["value"] = m.mockInput.GetValue()
		state["focused"] = m.mockInput.Focused()
	}
	if m.mockList != nil {
		state["index"] = m.mockList.Index()
		state["selected"] = m.mockList.SelectedItem()
	}
	return state
}

func (m *MockComponent) DetectStateChanges(old, new map[string]interface{}) []tea.Cmd {
	return []tea.Cmd{}
}

func (m *MockComponent) HandleSpecialKey(keyMsg tea.KeyMsg) (tea.Cmd, Response, bool) {
	if m.specialKey != nil {
		return m.specialKey(keyMsg)
	}
	return nil, Ignored, false
}

// 实现 ComponentInterface 的方法
func (m *MockComponent) View() string {
	return ""
}

func (m *MockComponent) Init() tea.Cmd {
	return nil
}

func (m *MockComponent) UpdateMsg(msg tea.Msg) (ComponentInterface, tea.Cmd, Response) {
	return m, nil, Handled
}

func (m *MockComponent) SetFocus(focus bool) {
	// 空实现
}

func (m *MockComponent) GetComponentType() string {
	return "mock"
}

func (m *MockComponent) Render(config RenderConfig) (string, error) {
	return "", nil
}

func (m *MockComponent) UpdateRenderConfig(config RenderConfig) error {
	return nil
}

func (m *MockComponent) Cleanup() {
	// 空实现
}

func (m *MockComponent) GetStateChanges() (map[string]interface{}, bool) {
	return nil, false
}

func (m *MockComponent) GetSubscribedMessageTypes() []string {
	return []string{}
}

func TestHandleTargetedMsg(t *testing.T) {
	// 测试定向消息匹配
	msg := TargetedMsg{TargetID: "test-id", InnerMsg: tea.KeyMsg{Type: tea.KeyEnter}}
	shouldRecurse, _, response := HandleTargetedMsg(msg, "test-id")
	assert.True(t, shouldRecurse)
	assert.Equal(t, Handled, response)

	// 测试不匹配
	shouldRecurse, _, response = HandleTargetedMsg(msg, "other-id")
	assert.False(t, shouldRecurse)
	assert.Equal(t, Ignored, response)

	// 测试非定向消息
	nonTargetedMsg := tea.KeyMsg{Type: tea.KeyEnter}
	shouldRecurse, _, response = HandleTargetedMsg(nonTargetedMsg, "test-id")
	assert.False(t, shouldRecurse)
	assert.Equal(t, Handled, response)
}

func TestInputStateHelper_DetectValuesChange(t *testing.T) {
	// 创建模拟输入
	mockInput := &MockInput{value: "old", focused: false}
	helper := &InputStateHelper{Valuer: mockInput, Focuser: mockInput, ComponentID: "test"}

	oldState := helper.CaptureState()
	mockInput.value = "new"
	newState := helper.CaptureState()

	eventCmds := helper.DetectStateChanges(oldState, newState)
	// 这里应该生成一个事件命令
	assert.Greater(t, len(eventCmds), 0)
}

func TestListStateHelper_DetectSelectionChange(t *testing.T) {
	// 创建模拟列表
	mockList := &MockList{index: 0, selected: "item1"}
	helper := &ListStateHelper{Indexer: mockList, Selector: mockList, ComponentID: "test"}

	oldState := helper.CaptureState()
	mockList.index = 1
	newState := helper.CaptureState()

	eventCmds := helper.DetectStateChanges(oldState, newState)
	// 这里应该生成一个事件命令
	assert.Greater(t, len(eventCmds), 0)
}

func TestHandleStateChanges(t *testing.T) {
	mockInput := &MockInput{value: "initial", focused: true}
	helper := &InputStateHelper{Valuer: mockInput, Focuser: mockInput, ComponentID: "test-state-change"}

	// 测试无更新命令的情况
	cmd := HandleStateChanges(helper, nil)
	assert.NotNil(t, cmd)

	// 测试有更新命令的情况
	updateCmd := func() tea.Msg {
		mockInput.value = "updated"
		return nil
	}
	cmd = HandleStateChanges(helper, updateCmd)
	assert.NotNil(t, cmd)
}

func TestDefaultInteractiveUpdateMsg(t *testing.T) {
	component := &MockComponent{
		id:       "test-component",
		hasFocus: true,
	}

	// 测试基本的消息处理
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}

	// 模拟获取绑定的函数
	getBindings := func() []ComponentBinding {
		return []ComponentBinding{}
	}

	// 模拟处理绑定的函数
	handleBinding := func(keyMsg tea.KeyMsg, binding ComponentBinding) (tea.Cmd, Response, bool) {
		return nil, Handled, false
	}

	// 模拟委托更新的函数
	delegateUpdate := func(msg tea.Msg) tea.Cmd {
		return nil
	}

	cmd, response := DefaultInteractiveUpdateMsg(component, msg, getBindings, handleBinding, delegateUpdate)
	assert.NotNil(t, cmd)
	assert.Equal(t, Handled, response)
}

func BenchmarkHandleTargetedMsg(b *testing.B) {
	msg := TargetedMsg{TargetID: "test-id", InnerMsg: tea.KeyMsg{Type: tea.KeyEnter}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		HandleTargetedMsg(msg, "test-id")
	}
}

func BenchmarkHandleStateChanges(b *testing.B) {
	mockInput := &MockInput{value: "initial", focused: true}
	helper := &InputStateHelper{Valuer: mockInput, Focuser: mockInput, ComponentID: "test-state-change"}

	updateCmd := func() tea.Msg {
		mockInput.value = "updated"
		return nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		HandleStateChanges(helper, updateCmd)
	}
}

func BenchmarkDefaultInteractiveUpdateMsg(b *testing.B) {
	component := &MockComponent{
		id:       "test-component",
		hasFocus: true,
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}

	getBindings := func() []ComponentBinding {
		return []ComponentBinding{}
	}

	handleBinding := func(keyMsg tea.KeyMsg, binding ComponentBinding) (tea.Cmd, Response, bool) {
		return nil, Handled, false
	}

	delegateUpdate := func(msg tea.Msg) tea.Cmd {
		return nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DefaultInteractiveUpdateMsg(component, msg, getBindings, handleBinding, delegateUpdate)
	}
}

// TestHandleStateChanges_WithUpdateMsgAndNoEventCmds 测试当有 updateMsg 但没有 eventCmds 时的情况
// 修复之前的问题：eventCmds 为空但 updateMsg 不为空时，应该返回 updateMsg
func TestHandleStateChanges_WithUpdateMsgAndNoEventCmds(t *testing.T) {
	mockInput := &MockInput{value: "initial", focused: true}
	helper := &InputStateHelper{Valuer: mockInput, Focuser: mockInput, ComponentID: "test"}

	// 创建一个会返回非 nil 消息的 updateCmd，但不会改变组件状态
	updateMsg := tea.KeyMsg{Type: tea.KeyEnter}
	updateCmd := func() tea.Msg {
		return updateMsg
	}

	cmd := HandleStateChanges(helper, updateCmd)
	assert.NotNil(t, cmd)

	// 执行命令获取消息
	msg := cmd()
	assert.Equal(t, updateMsg, msg)
}

// TestHandleStateChanges_WithUpdateMsgAndEventCmds 测试当有 updateMsg 和 eventCmds 时的情况
// 修复之前的问题：两者都应该被正确处理
func TestHandleStateChanges_WithUpdateMsgAndEventCmds(t *testing.T) {
	mockInput := &MockInput{value: "initial", focused: true}
	component := &MockComponent{
		id:        "test-component",
		mockInput: mockInput,
	}

	// 创建一个会改变组件状态并返回消息的 updateCmd
	updateMsg := tea.KeyMsg{Type: tea.KeyEnter}
	updateCmd := func() tea.Msg {
		mockInput.value = "updated"
		return updateMsg
	}

	cmd := HandleStateChanges(component, updateCmd)
	assert.NotNil(t, cmd)

	// 执行命令获取消息
	msg := cmd()
	assert.Equal(t, updateMsg, msg)
}

// TestHandleStateChanges_WithoutUpdateMsgAndWithEventCmds 测试没有 updateMsg 但有 eventCmds 时的情况
func TestHandleStateChanges_WithoutUpdateMsgAndWithEventCmds(t *testing.T) {
	mockInput := &MockInput{value: "initial", focused: true}
	helper := &InputStateHelper{Valuer: mockInput, Focuser: mockInput, ComponentID: "test"}

	// 创建一个会改变组件状态但返回 nil 的 updateCmd
	updateCmd := func() tea.Msg {
		mockInput.value = "updated"
		return nil
	}

	cmd := HandleStateChanges(helper, updateCmd)
	assert.NotNil(t, cmd)

	// 执行命令，应该处理状态变化事件
	msg := cmd()
	// updateMsg 是 nil，所以最终消息应该是 nil（因为 eventCmds 也是返回 nil 的 PublishEvent）
	assert.Nil(t, msg)
}

// TestHandleDefaultEscape 测试 handleDefaultEscape 函数
func TestHandleDefaultEscape(t *testing.T) {
	// 创建一个模拟组件
	mockComponent := &MockComponent{
		id:       "test-component",
		hasFocus: true,
	}

	// 调用 handleDefaultEscape
	cmd, response := handleDefaultEscape(mockComponent)

	// 验证返回值
	assert.NotNil(t, cmd)
	assert.Equal(t, Ignored, response)

	// 验证事件被发布（通过执行命令）
	msg := cmd()
	assert.NotNil(t, msg)

	// 检查消息类型
	eventMsg, ok := msg.(ActionMsg)
	assert.True(t, ok, "应该返回 ActionMsg 类型")
	assert.Equal(t, EventEscapePressed, eventMsg.Action)
	assert.Equal(t, "test-component", eventMsg.ID)
}

// TestHandleDefaultEscape_WithoutBlur 测试没有 Blur 方法的组件
func TestHandleDefaultEscape_WithoutBlur(t *testing.T) {
	// 创建一个没有实现 Blur 方法的模拟组件
	mockComponent := &MockComponent{
		id:       "test-component-no-blur",
		hasFocus: true,
	}

	// 调用 handleDefaultEscape
	cmd, response := handleDefaultEscape(mockComponent)

	// 验证返回值
	assert.NotNil(t, cmd)
	assert.Equal(t, Ignored, response)

	// 验证事件被发布
	msg := cmd()
	assert.NotNil(t, msg)

	eventMsg, ok := msg.(ActionMsg)
	assert.True(t, ok, "应该返回 ActionMsg 类型")
	assert.Equal(t, EventEscapePressed, eventMsg.Action)
	assert.Equal(t, "test-component-no-blur", eventMsg.ID)
}

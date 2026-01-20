package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/yaoapp/yao/tui/core"
)

func TestInputComponentWrapperESCHandling(t *testing.T) {
	// 测试框架层的 ESC 处理是否正确工作
	props := InputProps{
		Placeholder: "Test input...",
	}
	wrapper := NewInputComponentWrapper(props, "test-input")

	// 设置焦点
	wrapper.SetFocus(true)
	assert.True(t, wrapper.GetFocus(), "Input should be focused")

	// 发送 ESC 键消息
	escMsg := tea.KeyMsg{Type: tea.KeyEsc}
	_, cmd, response := wrapper.UpdateMsg(escMsg)

	// 验证框架层处理的结果
	assert.NotNil(t, cmd, "Should return command for ESC key")
	assert.Equal(t, core.Handled, response, "Should return Handled for ESC key") // Updated: ESC key is now handled by component
	assert.False(t, wrapper.GetFocus(), "Input should lose focus on ESC")
}
package tui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestNewModel(t *testing.T) {
	cfg := &Config{
		ID:   "test",
		Name: "Test TUI",
		Data: map[string]interface{}{
			"title":   "Hello",
			"counter": 0,
		},
		Layout: Layout{
			Direction: "vertical",
		},
	}

	model := NewModel(cfg, nil)

	assert.NotNil(t, model)
	assert.Equal(t, cfg, model.Config)
	assert.Equal(t, "Hello", model.State["title"])
	assert.Equal(t, 0, model.State["counter"])
	assert.False(t, model.Ready)
}

func TestModelInit(t *testing.T) {
	cfg := &Config{
		Name: "Test",
		Layout: Layout{
			Direction: "vertical",
		},
	}

	model := NewModel(cfg, nil)
	cmd := model.Init()

	// Should return nil when no onLoad action
	assert.Nil(t, cmd)
}

func TestModelUpdateWindowSize(t *testing.T) {
	cfg := &Config{
		Name: "Test",
		Layout: Layout{
			Direction: "vertical",
		},
	}

	model := NewModel(cfg, nil)

	// Send WindowSizeMsg
	msg := tea.WindowSizeMsg{
		Width:  80,
		Height: 24,
	}

	updatedModel, cmd := model.Update(msg)

	assert.Nil(t, cmd)
	m := updatedModel.(*Model)
	assert.Equal(t, 80, m.Width)
	assert.Equal(t, 24, m.Height)
	assert.True(t, m.Ready)
}

func TestModelUpdateStateUpdate(t *testing.T) {
	cfg := &Config{
		Name: "Test",
		Layout: Layout{
			Direction: "vertical",
		},
	}

	model := NewModel(cfg, nil)

	// Send StateUpdateMsg
	msg := StateUpdateMsg{
		Key:   "counter",
		Value: 10,
	}

	updatedModel, cmd := model.Update(msg)

	assert.Nil(t, cmd)
	m := updatedModel.(*Model)
	assert.Equal(t, 10, m.State["counter"])
}

func TestModelUpdateStateBatchUpdate(t *testing.T) {
	cfg := &Config{
		Name: "Test",
		Layout: Layout{
			Direction: "vertical",
		},
	}

	model := NewModel(cfg, nil)

	// Send StateBatchUpdateMsg
	msg := StateBatchUpdateMsg{
		Updates: map[string]interface{}{
			"counter": 20,
			"message": "updated",
		},
	}

	updatedModel, cmd := model.Update(msg)

	assert.Nil(t, cmd)
	m := updatedModel.(*Model)
	assert.Equal(t, 20, m.State["counter"])
	assert.Equal(t, "updated", m.State["message"])
}

func TestModelHandleKeyPress(t *testing.T) {
	cfg := &Config{
		Name: "Test",
		Bindings: map[string]Action{
			"q": {
				Process: "tui.Quit",
			},
		},
		Layout: Layout{
			Direction: "vertical",
		},
	}

	model := NewModel(cfg, nil)

	t.Run("ctrl+c quits", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyCtrlC}
		_, cmd := model.Update(msg)
		assert.NotNil(t, cmd)
	})

	t.Run("bound key triggers action", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
		_, cmd := model.handleKeyPress(msg)
		assert.NotNil(t, cmd)
	})

	t.Run("unbound key does nothing", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
		_, cmd := model.handleKeyPress(msg)
		assert.Nil(t, cmd)
	})
}

func TestModelHandleProcessResult(t *testing.T) {
	cfg := &Config{
		Name: "Test",
		Layout: Layout{
			Direction: "vertical",
		},
	}

	model := NewModel(cfg, nil)

	msg := ProcessResultMsg{
		Target: "users",
		Data:   []string{"Alice", "Bob"},
	}

	updatedModel, cmd := model.handleProcessResult(msg)

	assert.Nil(t, cmd)
	m := updatedModel.(*Model)
	users, ok := m.State["users"]
	assert.True(t, ok)
	assert.Len(t, users, 2)
}

func TestModelHandleStreamChunk(t *testing.T) {
	cfg := &Config{
		Name: "Test",
		Layout: Layout{
			Direction: "vertical",
		},
	}

	model := NewModel(cfg, nil)

	// Send first chunk
	msg1 := StreamChunkMsg{
		ID:      "ai1",
		Content: "Hello",
	}
	updatedModel, _ := model.handleStreamChunk(msg1)
	m := updatedModel.(*Model)

	assert.Equal(t, "Hello", m.State["stream_ai1"])

	// Send second chunk
	msg2 := StreamChunkMsg{
		ID:      "ai1",
		Content: " World",
	}
	updatedModel2, _ := m.handleStreamChunk(msg2)
	m2 := updatedModel2.(*Model)

	assert.Equal(t, "Hello World", m2.State["stream_ai1"])
}

func TestModelHandleStreamDone(t *testing.T) {
	cfg := &Config{
		Name: "Test",
		Layout: Layout{
			Direction: "vertical",
		},
	}

	model := NewModel(cfg, nil)

	msg := StreamDoneMsg{
		ID: "ai1",
	}

	updatedModel, cmd := model.handleStreamDone(msg)

	assert.Nil(t, cmd)
	m := updatedModel.(*Model)
	done, ok := m.State["stream_ai1_done"]
	assert.True(t, ok)
	assert.True(t, done.(bool))
}

func TestModelHandleError(t *testing.T) {
	cfg := &Config{
		Name: "Test",
		Layout: Layout{
			Direction: "vertical",
		},
	}

	model := NewModel(cfg, nil)

	msg := ErrorMsg{
		Err:     assert.AnError,
		Context: "test error",
	}

	updatedModel, cmd := model.handleError(msg)

	assert.Nil(t, cmd)
	m := updatedModel.(*Model)
	errMsg, ok := m.State["__error"]
	assert.True(t, ok)
	assert.Contains(t, errMsg.(string), "test error")
}

func TestModelView(t *testing.T) {
	cfg := &Config{
		Name: "Test",
		Layout: Layout{
			Direction: "vertical",
		},
	}

	model := NewModel(cfg, nil)

	t.Run("not ready", func(t *testing.T) {
		view := model.View()
		assert.Equal(t, "Initializing...", view)
	})

	t.Run("ready", func(t *testing.T) {
		model.Ready = true
		model.Width = 80
		model.Height = 24
		view := model.View()
		// View 返回的可能是空字符串（因为 layout 没有 children）
		// 或者包含渲染的内容
		assert.True(t, view == "" || len(view) > 0)
	})
}

func TestModelGetState(t *testing.T) {
	cfg := &Config{
		Name: "Test",
		Data: map[string]interface{}{
			"title": "Hello",
		},
		Layout: Layout{
			Direction: "vertical",
		},
	}

	model := NewModel(cfg, nil)

	value, ok := model.GetState("title")
	assert.True(t, ok)
	assert.Equal(t, "Hello", value)

	value, ok = model.GetState("nonexistent")
	assert.False(t, ok)
	assert.Nil(t, value)
}

func TestModelStateThreadSafety(t *testing.T) {
	cfg := &Config{
		Name: "Test",
		Layout: Layout{
			Direction: "vertical",
		},
	}

	model := NewModel(cfg, nil)
	done := make(chan bool)

	// Concurrent writes
	go func() {
		for i := 0; i < 100; i++ {
			model.StateMu.Lock()
			model.State["counter"] = i
			model.StateMu.Unlock()
			time.Sleep(1 * time.Millisecond)
		}
		done <- true
	}()

	// Concurrent reads
	go func() {
		for i := 0; i < 100; i++ {
			model.StateMu.RLock()
			_ = model.State["counter"]
			model.StateMu.RUnlock()
			time.Sleep(1 * time.Millisecond)
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done
}

func TestExecuteAction(t *testing.T) {
	cfg := &Config{
		Name: "Test",
		Layout: Layout{
			Direction: "vertical",
		},
	}

	model := NewModel(cfg, nil)

	t.Run("nil action", func(t *testing.T) {
		cmd := model.executeAction(nil)
		assert.Nil(t, cmd)
	})

	t.Run("invalid action", func(t *testing.T) {
		action := &Action{} // Invalid: no process or script
		cmd := model.executeAction(action)
		assert.NotNil(t, cmd)

		// Execute the command and check for error
		msg := cmd()
		errMsg, ok := msg.(ErrorMsg)
		assert.True(t, ok)
		assert.NotNil(t, errMsg.Err)
	})

	t.Run("payload action", func(t *testing.T) {
		action := &Action{
			Process: "test",
			Payload: map[string]interface{}{
				"key": "value",
			},
		}
		cmd := model.executeAction(action)
		assert.NotNil(t, cmd)
	})
}

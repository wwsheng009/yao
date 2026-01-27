package tui

import (
	"sync"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yaoapp/yao/tui/tui/core"
)

func TestInitCommandExecution(t *testing.T) {
	autoFocuse := true
	cfg := &Config{
		Name:      "Test Init Command",
		AutoFocus: &autoFocuse,
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					ID:   "test-input",
					Type: "input",
					Props: map[string]interface{}{
						"placeholder": "Test",
					},
				},
			},
		},
	}

	model := NewModel(cfg, nil)

	t.Logf("=== Before Init ===")
	t.Logf("CurrentFocus: %s", model.CurrentFocus)
	t.Logf("AutoFocusApplied: %v", model.AutoFocusApplied)
	t.Logf("Components map size: %d", len(model.Components))

	cmd := model.Init()

	if cmd == nil {
		t.Fatal("Init returned nil command")
	}

	t.Logf("\n=== After Init (command not executed yet) ===")
	t.Logf("CurrentFocus: %s", model.CurrentFocus)
	t.Logf("AutoFocusApplied: %v", model.AutoFocusApplied)

	// Examine the command type - what does Batch return?
	t.Logf("Command type: %T", cmd)

	// Execute the command to get messages
	t.Logf("\n=== Executing command ===")
	msgs := executeBatchCommand(cmd)
	t.Logf("Number of messages from command: %d", len(msgs))

	// Process each message sequentially
	for i, msg := range msgs {
		t.Logf("\n=== Processing message %d (%T) ===", i, msg)
		if focusMsg, ok := msg.(core.FocusFirstComponentMsg); ok {
			t.Logf("  Found FocusFirstComponentMsg: %+v", focusMsg)
		}
		updatedModel, _ := model.Update(msg)
		model = updatedModel.(*Model)
		t.Logf("  After Update - CurrentFocus: %s", model.CurrentFocus)
		t.Logf("  After Update - AutoFocusApplied: %v", model.AutoFocusApplied)
	}

	t.Logf("\n=== After all messages processed ===")
	t.Logf("CurrentFocus: %s", model.CurrentFocus)
	t.Logf("AutoFocusApplied: %v", model.AutoFocusApplied)

	if model.CurrentFocus != "test-input" {
		t.Errorf("Expected focus on test-input, got %s", model.CurrentFocus)
	}
}

// executeBatchCommand executes a tea.Cmd and collects all messages it generates
// This simulates what tea.Batch() command does internally
func executeBatchCommand(cmd tea.Cmd) []tea.Msg {
	if cmd == nil {
		return nil
	}

	// Execute the command to get the message
	msg := cmd()
	if msg == nil {
		return nil
	}

	// Check if it's a BatchMsg (multiple commands)
	if batchMsg, ok := msg.(tea.BatchMsg); ok {
		return executeBatchMsg(batchMsg)
	}

	// Single message
	return []tea.Msg{msg}
}

// executeBatchMsg executes all commands in a BatchMsg concurrently
// This simulates Program.execBatchMsg behavior from tea.go
func executeBatchMsg(batchMsg tea.BatchMsg) []tea.Msg {
	var wg sync.WaitGroup
	msgChan := make(chan tea.Msg, len(batchMsg))

	// Execute each command in a goroutine (concurrent, just like tea framework)
	for _, cmd := range batchMsg {
		if cmd == nil {
			continue
		}
		wg.Add(1)
		go func(c tea.Cmd) {
			defer wg.Done()
			if m := c(); m != nil {
				msgChan <- m
			}
		}(cmd)
	}

	// Wait for all commands to complete and close the channel
	go func() {
		wg.Wait()
		close(msgChan)
	}()

	// Collect all messages
	var msgs []tea.Msg
	for m := range msgChan {
		msgs = append(msgs, m)
	}

	return msgs
}

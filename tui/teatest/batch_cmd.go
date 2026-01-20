package teatest

import (
	"sync"

	tea "github.com/charmbracelet/bubbletea"
)

// BatchCommandTest is a test utility for working with tea.Batch() commands
// It helps execute and collect messages from Batch commands in tests

// ExecuteBatchCommand executes a tea.Cmd and collects all messages it generates
// This simulates what tea.Batch() command does internally
// Returns a slice of messages that should be processed sequentially
func ExecuteBatchCommand(cmd tea.Cmd) []tea.Msg {
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
		return executeBatch(batchMsg)
	}

	// Single message
	return []tea.Msg{msg}
}

// executeBatch executes all commands in a BatchMsg concurrently
// This simulates Program.execBatchMsg behavior from tea.go
// WaitGroup ensures all commands complete before returning
func executeBatch(batchMsg tea.BatchMsg) []tea.Msg {
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

// ProcessSequentialCmd executes a tea.Cmd and processes all resulting messages sequentially
// This is the most common pattern in TUI tests for initializing components
// It returns the final updated model
// Simplified version without recursion to avoid stack overflow
func ProcessSequentialCmd(model tea.Model, cmd tea.Cmd) tea.Model {
	if cmd == nil {
		return model
	}

	// Execute command and get all messages
	msgs := ExecuteBatchCommand(cmd)

	// Process each message sequentially through Update()
	// Only one level deep to avoid infinite recursion
	for _, msg := range msgs {
		updatedModel, _ := model.Update(msg)
		model = updatedModel
	}

	return model
}

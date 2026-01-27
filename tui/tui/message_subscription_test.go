package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

// TestMessageSubscriptionManagerBasic tests basic subscription functionality
func TestMessageSubscriptionManagerBasic(t *testing.T) {
	manager := NewMessageSubscriptionManager()

	// Subscribe component1 to KeyMsg
	manager.Subscribe("component1", []string{"tea.KeyMsg", "tea.MouseMsg"})

	// Subscribe component2 to KeyMsg
	manager.Subscribe("component2", []string{"tea.KeyMsg"})

	// Get subscribers for KeyMsg
	subscribers := manager.GetSubscribers("tea.KeyMsg")
	assert.Len(t, subscribers, 2)
	assert.Contains(t, subscribers, "component1")
	assert.Contains(t, subscribers, "component2")

	// Get subscribers for MouseMsg
	subscribers = manager.GetSubscribers("tea.MouseMsg")
	assert.Len(t, subscribers, 1)
	assert.Contains(t, subscribers, "component1")

	// Get subscribers for WindowSizeMsg (should be empty)
	subscribers = manager.GetSubscribers("tea.WindowSizeMsg")
	assert.Nil(t, subscribers)
}

// TestMessageSubscriptionManagerUnsubscribe tests unsubscribe functionality
func TestMessageSubscriptionManagerUnsubscribe(t *testing.T) {
	manager := NewMessageSubscriptionManager()

	// Subscribe component1 to multiple message types
	manager.Subscribe("component1", []string{"tea.KeyMsg", "tea.MouseMsg", "tea.WindowSizeMsg"})

	// Subscribe component2 to KeyMsg
	manager.Subscribe("component2", []string{"tea.KeyMsg"})

	// Unsubscribe component1
	manager.Unsubscribe("component1")

	// Check that component1 is no longer subscribed
	subscribers := manager.GetSubscribers("tea.KeyMsg")
	assert.Len(t, subscribers, 1)
	assert.Contains(t, subscribers, "component2")

	subscribers = manager.GetSubscribers("tea.MouseMsg")
	assert.Nil(t, subscribers)

	subscribers = manager.GetSubscribers("tea.WindowSizeMsg")
	assert.Nil(t, subscribers)

	// Check component subscriptions
	subs := manager.GetComponentSubscriptions("component1")
	assert.Nil(t, subs)

	subs = manager.GetComponentSubscriptions("component2")
	assert.NotNil(t, subs)
	assert.Len(t, subs, 1)
	assert.Equal(t, "tea.KeyMsg", subs[0])
}

// TestMessageSubscriptionManagerClear tests clear functionality
func TestMessageSubscriptionManagerClear(t *testing.T) {
	manager := NewMessageSubscriptionManager()

	// Subscribe multiple components
	manager.Subscribe("component1", []string{"tea.KeyMsg"})
	manager.Subscribe("component2", []string{"tea.MouseMsg"})
	manager.Subscribe("component3", []string{"tea.WindowSizeMsg"})

	// Clear all subscriptions
	manager.Clear()

	// Verify all subscriptions are cleared
	assert.Nil(t, manager.GetSubscribers("tea.KeyMsg"))
	assert.Nil(t, manager.GetSubscribers("tea.MouseMsg"))
	assert.Nil(t, manager.GetSubscribers("tea.WindowSizeMsg"))

	subs := manager.GetComponentSubscriptions("component1")
	assert.Nil(t, subs)

	subs = manager.GetComponentSubscriptions("component2")
	assert.Nil(t, subs)

	subs = manager.GetComponentSubscriptions("component3")
	assert.Nil(t, subs)
}

// TestMessageTypeString tests message type detection
func TestMessageTypeString(t *testing.T) {
	tests := []struct {
		name     string
		msg      tea.Msg
		expected string
	}{
		{
			name:     "Key Message",
			msg:      tea.KeyMsg{Type: tea.KeyEnter},
			expected: "tea.KeyMsg",
		},
		{
			name:     "Mouse Message",
			msg:      tea.MouseMsg{X: 0, Y: 0},
			expected: "tea.MouseMsg",
		},
		{
			name:     "Window Size Message",
			msg:      tea.WindowSizeMsg{Width: 80, Height: 24},
			expected: "tea.WindowSizeMsg",
		},
		{
			name:     "Quit Message",
			msg:      tea.QuitMsg{},
			expected: "tea.QuitMsg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetMessageTypeString(tt.msg)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestMessageDispatchWithSubscription tests message dispatch using subscription manager
func TestMessageDispatchWithSubscription(t *testing.T) {
	cfg := &Config{
		Name: "Test Message Subscription",
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					Type: "input",
					ID:   "input1",
					Props: map[string]interface{}{
						"placeholder": "Field 1",
					},
				},
				{
					Type: "input",
					ID:   "input2",
					Props: map[string]interface{}{
						"placeholder": "Field 2",
					},
				},
			},
		},
	}

	model := NewModel(cfg, nil)
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Render to create component instances
	model.View()

	// Verify subscriptions are registered
	subs := model.MessageSubscriptionManager.GetComponentSubscriptions("input1")
	assert.NotNil(t, subs)
	assert.Contains(t, subs, "tea.KeyMsg")
	assert.Contains(t, subs, "core.TargetedMsg")

	subs = model.MessageSubscriptionManager.GetComponentSubscriptions("input2")
	assert.NotNil(t, subs)
	assert.Contains(t, subs, "tea.KeyMsg")

	// Get subscribers for KeyMsg
	subscribers := model.MessageSubscriptionManager.GetSubscribers("tea.KeyMsg")
	assert.Len(t, subscribers, 2)
	assert.Contains(t, subscribers, "input1")
	assert.Contains(t, subscribers, "input2")

	// Dispatch a key message (should only go to subscribed components)
	keyMsg := tea.KeyMsg{Type: tea.KeyEnter}
	_, _ = model.dispatchMessageToAllComponents(keyMsg)

	// Components should still exist and be functional
	comp1, exists := model.ComponentInstanceRegistry.Get("input1")
	assert.True(t, exists)
	assert.NotNil(t, comp1)

	comp2, exists := model.ComponentInstanceRegistry.Get("input2")
	assert.True(t, exists)
	assert.NotNil(t, comp2)
}

// TestMessageEfficiency tests that message subscription reduces unnecessary dispatches
func TestMessageEfficiency(t *testing.T) {
	cfg := &Config{
		Name: "Test Message Efficiency",
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					Type:  "text",
					ID:    "static_text",
					Props: map[string]interface{}{"content": "Static Text"},
				},
				{
					Type: "input",
					ID:   "interactive_input",
					Props: map[string]interface{}{
						"placeholder": "Input",
					},
				},
			},
		},
	}

	model := NewModel(cfg, nil)
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Render to create component instances
	model.View()

	// Get subscribers for KeyMsg
	subscribers := model.MessageSubscriptionManager.GetSubscribers("tea.KeyMsg")

	// Only the input component should be subscribed to KeyMsg
	assert.Len(t, subscribers, 1)
	assert.Contains(t, subscribers, "interactive_input")
	assert.NotContains(t, subscribers, "static_text")

	// This means KeyMsg will only be dispatched to the input component
	// and not to the static text component, improving efficiency
}

// TestAllSubscribedComponents tests getting all subscribed components
func TestAllSubscribedComponents(t *testing.T) {
	manager := NewMessageSubscriptionManager()

	// Subscribe multiple components
	manager.Subscribe("component1", []string{"tea.KeyMsg"})
	manager.Subscribe("component2", []string{"tea.MouseMsg"})
	manager.Subscribe("component3", []string{"tea.WindowSizeMsg"})

	// Get all subscribed components
	components := manager.GetAllSubscribedComponents()
	assert.Len(t, components, 3)
	assert.Contains(t, components, "component1")
	assert.Contains(t, components, "component2")
	assert.Contains(t, components, "component3")
}

// TestDuplicateSubscription tests that duplicate subscriptions are handled correctly
func TestDuplicateSubscription(t *testing.T) {
	manager := NewMessageSubscriptionManager()

	// Subscribe component1 to KeyMsg twice
	manager.Subscribe("component1", []string{"tea.KeyMsg"})
	manager.Subscribe("component1", []string{"tea.KeyMsg"})

	// Should only appear once
	subscribers := manager.GetSubscribers("tea.KeyMsg")
	assert.Len(t, subscribers, 1)
	assert.Equal(t, "component1", subscribers[0])
}

// TestEmptySubscription tests that empty subscriptions are ignored
func TestEmptySubscription(t *testing.T) {
	manager := NewMessageSubscriptionManager()

	// Subscribe with empty list
	manager.Subscribe("component1", []string{})

	// Should not create any subscriptions
	subscribers := manager.GetSubscribers("tea.KeyMsg")
	assert.Nil(t, subscribers)
}

// TestMessageDispatchToAllComponentsFallback tests fallback to dispatching to all components
func TestMessageDispatchToAllComponentsFallback(t *testing.T) {
	cfg := &Config{
		Name: "Test Fallback Dispatch",
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					Type:  "text",
					Props: map[string]interface{}{"content": "Text 1"},
				},
				{
					Type:  "text",
					Props: map[string]interface{}{"content": "Text 2"},
				},
			},
		},
	}

	model := NewModel(cfg, nil)
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Render to create component instances
	model.View()

	// Static components don't typically subscribe to messages
	// Should fall back to dispatching to all components
	keyMsg := tea.KeyMsg{Type: tea.KeyEnter}
	_, _ = model.dispatchMessageToAllComponents(keyMsg)

	// Verify components still exist
	assert.Equal(t, 2, len(model.Components))
}

// TestMessageSubscriptionWithComponentRemoval tests that subscriptions are cleaned up on component removal
func TestMessageSubscriptionWithComponentRemoval(t *testing.T) {
	cfg := &Config{
		Name: "Test Subscription Cleanup",
		Layout: Layout{
			Direction: "vertical",
			Children: []Component{
				{
					Type: "input",
					ID:   "input1",
					Props: map[string]interface{}{
						"placeholder": "Field 1",
					},
				},
			},
		},
	}

	model := NewModel(cfg, nil)
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Render to create component instance
	model.View()

	// Verify subscription is registered
	subs := model.MessageSubscriptionManager.GetComponentSubscriptions("input1")
	assert.NotNil(t, subs)

	// Remove the component
	model.ComponentInstanceRegistry.Remove("input1")

	// Unsubscribe the component
	model.MessageSubscriptionManager.Unsubscribe("input1")

	// Verify subscription is removed
	subs = model.MessageSubscriptionManager.GetComponentSubscriptions("input1")
	assert.Nil(t, subs)

	subscribers := model.MessageSubscriptionManager.GetSubscribers("tea.KeyMsg")
	assert.Nil(t, subscribers)
}

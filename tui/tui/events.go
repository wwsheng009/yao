package tui

import (
	"github.com/yaoapp/yao/tui/tui/core"
)

// NewEventBus creates a new EventBus instance
func NewEventBus() *core.EventBus {
	return core.NewEventBus()
}

// SubscribeToEventBus registers a callback for a specific action
// Returns an unsubscribe function that should be called to clean up
func SubscribeToEventBus(bus *core.EventBus, action string, callback func(core.ActionMsg)) func() {
	return bus.Subscribe(action, callback)
}

// PublishToEventBus sends an action message to all subscribers
func PublishToEventBus(bus *core.EventBus, msg core.ActionMsg) {
	bus.Publish(msg)
}

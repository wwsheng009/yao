package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yaoapp/kun/log"
	"github.com/yaoapp/yao/tui/core"
)

// InitializeComponents initializes all component instances before rendering.
// This should be called once during model initialization.
// Component instances are created and registered, ready for rendering.
func (m *Model) InitializeComponents() []tea.Cmd {
	log.Trace("InitializeComponents: Starting component initialization")

	// Ensure component registry is initialized
	if m.ComponentInstanceRegistry == nil {
		m.ComponentInstanceRegistry = NewComponentInstanceRegistry()
	}

	// Get component factory from global registry
	registry := GetGlobalRegistry()

	var allCmds []tea.Cmd
	// Recursively initialize all components in the layout
	if err := m.initializeLayoutNode(&m.Config.Layout, m.Width, m.Height, registry, 0, &allCmds); err != nil {
		log.Error("InitializeComponents error: %v", err)
		// Continue initialization even if some components fail
	}

	return allCmds
}

// initializeLayoutNode recursively initializes components in a layout node
func (m *Model) initializeLayoutNode(layout *Layout, width, height int, registry *ComponentRegistry, depth int, cmds *[]tea.Cmd) error {
	// Check maximum layout depth to prevent stack overflow
	if depth > maxLayoutDepth {
		return fmt.Errorf("layout depth exceeded maximum limit: %d (max: %d)", depth, maxLayoutDepth)
	}

	if len(layout.Children) == 0 {
		return nil
	}

	// Initialize each child component
	for _, child := range layout.Children {
		// If child is a layout component, initialize it recursively
		if child.Type == "layout" {
			if nestedLayout, ok := child.Props["layout"].(*Layout); ok {
				if err := m.initializeLayoutNode(nestedLayout, width, height, registry, depth+1, cmds); err != nil {
					return err
				}
				continue
			}
		}

		// Initialize individual component
		if err := m.initializeComponent(&child, registry, cmds); err != nil {
			return err
		}
	}

	return nil
}

// initializeComponent creates and registers a component instance
func (m *Model) initializeComponent(comp *Component, registry *ComponentRegistry, cmds *[]tea.Cmd) error {
	if comp == nil || comp.Type == "" {
		return fmt.Errorf("component is nil or has empty type")
	}

	// Get component factory
	factory, exists := registry.GetComponentFactory(ComponentType(comp.Type))
	if !exists || factory == nil {
		return fmt.Errorf("unknown component type: %s", comp.Type)
	}

	// Apply state binding to props for initialization
	props := m.resolveProps(comp)

	// Create render config for initialization
	// 初始化时的渲染配置，高宽是0，没有意义，屏幕还没显示。
	renderConfig := core.RenderConfig{
		Data:  props,
		Width: m.Width,
		Height: m.Height,
	}

	// Get or create component instance
	componentInstance, isNew := m.ComponentInstanceRegistry.GetOrCreate(
		comp.ID,
		comp.Type,
		factory,
		renderConfig,
	)

	// For interactive components with ID, register in Components map and set up subscriptions
	if comp.ID != "" && isInteractiveComponent(comp.Type) {
		if m.Components == nil {
			m.Components = make(map[string]*core.ComponentInstance)
		}

		// Always register in Components map (even if existing instance)
		m.Components[comp.ID] = componentInstance

		if isNew {
			log.Trace("InitializeComponents: Created new component instance %s (type: %s)", comp.ID, comp.Type)

			// Register message subscriptions if component implements GetSubscribedMessageTypes
			if subscriber, ok := componentInstance.Instance.(interface{ GetSubscribedMessageTypes() []string }); ok {
				messageTypes := subscriber.GetSubscribedMessageTypes()
				if len(messageTypes) > 0 {
					m.MessageSubscriptionManager.Subscribe(comp.ID, messageTypes)
					log.Trace("InitializeComponents: Registered message subscriptions for %s: %v", comp.ID, messageTypes)
				}
			}
		}

		// Call Init() method on the component instance
		if initCmd := componentInstance.Instance.Init(); initCmd != nil {
			// Collect command instead of sending it directly
			*cmds = append(*cmds, initCmd)
		}
	} else {
		log.Trace("InitializeComponents: Reusing existing component instance %s (type: %s)", comp.ID, comp.Type)
	}

	return nil
}

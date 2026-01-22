package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yaoapp/kun/log"
	"github.com/yaoapp/yao/tui/core"
	"github.com/yaoapp/yao/tui/layout"
)

const maxLayoutDepth = 50

// InitializeComponents initializes all component instances and creates the layout tree.
func (m *Model) InitializeComponents() []tea.Cmd {
	log.Trace("InitializeComponents: Starting component initialization")

	if m.ComponentInstanceRegistry == nil {
		m.ComponentInstanceRegistry = NewComponentInstanceRegistry()
	}

	registry := GetGlobalRegistry()

	// Create layout tree from Config.Layout
	m.LayoutRoot = m.createLayoutTree(&m.Config.Layout, registry, 0)
	log.Trace("InitializeComponents: Created layout tree with root ID: %s", m.LayoutRoot.ID)

	// Initialize layout engine with the created tree
	m.LayoutEngine = layout.NewEngine(&layout.LayoutConfig{
		Root: m.LayoutRoot,
		WindowSize: &layout.WindowSize{
			Width:  m.Width,
			Height: m.Height,
		},
		PropsResolver: func(node *layout.LayoutNode) map[string]interface{} {
			// Find component config by ID
			if node.ID == "" {
				return node.Props
			}
			comp := m.findComponentConfig(node.ID)
			if comp != nil {
				// Use Model's props resolution logic which handles {{}} expressions
				return m.resolveProps(comp)
			}
			return node.Props
		},
	})
	log.Trace("InitializeComponents: Created layout engine, window: %dx%d", m.Width, m.Height)

	var allCmds []tea.Cmd

	// Initialize all components in the layout tree
	m.initializeLayoutComponents(m.LayoutRoot, registry, &allCmds)
	log.Trace("InitializeComponents: Initialized components, got %d commands", len(allCmds))

	// Initialize the Renderer with the LayoutEngine and this Model as context
	m.Renderer = layout.NewRenderer(m.LayoutEngine, m)
	log.Trace("InitializeComponents: Initialized renderer")

	// Don't trigger layout calculation yet - wait for window size message
	// Layout will be calculated on first render when window size is known

	// Handle autofocus
	if m.Config.AutoFocus != nil && *m.Config.AutoFocus {
		focusableIDs := m.getFocusableComponentIDs()
		if len(focusableIDs) > 0 {
			log.Trace("InitializeComponents: AutoFocus enabled, routing focus to first focusable component: %s", focusableIDs[0])
			cmd := m.setFocus(focusableIDs[0])
			if cmd != nil {
				allCmds = append(allCmds, cmd)
			}
		} else {
			log.Trace("InitializeComponents: AutoFocus enabled but no focusable components found")
		}
	}

	return allCmds
}

// createLayoutTree creates a LayoutNode tree from the Config.Layout structure.
func (m *Model) createLayoutTree(layoutCfg *Layout, registry *ComponentRegistry, depth int) *layout.LayoutNode {
	if layoutCfg == nil || depth > maxLayoutDepth {
		return nil
	}

	builder := layout.NewBuilder()

	direction := layout.DirectionColumn
	if layoutCfg.Direction == "horizontal" {
		direction = layout.DirectionRow
	}

	builder.PushContainer(&layout.ContainerConfig{
		ID:   fmt.Sprintf("layout_depth_%d", depth),
		Type: layout.LayoutFlex,
	})

	layout.ApplyStyle(builder.Current(),
		layout.WithFlexDirection(direction),
	)

	if len(layoutCfg.Padding) > 0 {
		padding := layoutCfg.Padding
		for len(padding) < 4 {
			padding = append(padding, padding[len(padding)-1])
		}
		layout.ApplyStyle(builder.Current(),
			layout.WithPadding(padding[0], padding[1], padding[2], padding[3]),
		)
	}

	for _, child := range layoutCfg.Children {
		if child.Type == "layout" {
			// Support both old format (children + direction) and new format (props.layout)
			if nestedLayout, ok := child.Props["layout"].(*Layout); ok {
				// New format: nested layout in props.layout
				nestedNode := m.createLayoutTree(nestedLayout, registry, depth+1)
				if nestedNode != nil {
					builder.AddNode(nestedNode)
				}
			} else if len(child.Children) > 0 {
				// Old format: type="layout" has its own direction and children
				// Create a temporary Layout structure from child
				oldFormatLayout := &Layout{
					Direction: child.Direction,
					Children:  child.Children,
					Padding:   layoutCfg.Padding, // Inherit padding from parent
				}
				nestedNode := m.createLayoutTree(oldFormatLayout, registry, depth+1)
				if nestedNode != nil {
					builder.AddNode(nestedNode)
				}
			}
		} else {
			// Regular component - create placeholder node
			componentNode := m.createComponentNode(child)
			componentNode.Parent = builder.Current()
			builder.Current().Children = append(builder.Current().Children, componentNode)
		}
	}

	builder.Pop()
	return builder.Root()
}

// createComponentNode creates a layout node for a regular component
func (m *Model) createComponentNode(child Component) *layout.LayoutNode {
	// Create size objects from component properties
	var width, height *layout.Size

	if child.Width != nil {
		width = layout.NewSize(child.Width)
	} else {
		width = layout.NewSize(nil)
	}

	if child.Height != nil {
		height = layout.NewSize(child.Height)
	} else {
		height = layout.NewSize(nil)
	}

	node := &layout.LayoutNode{
		ID:            child.ID,
		Type:          layout.LayoutFlex, // Container type
		ComponentType: child.Type,        // Store original component type
		Props:         child.Props,       // Store component properties
		Style: &layout.LayoutStyle{
			Direction: layout.DirectionColumn,
			Width:     width,
			Height:    height,
			Position:  layout.PositionRelative,
		},
	}
	return node
}

// initializeLayoutComponents walks the layout tree and initializes component instances.
func (m *Model) initializeLayoutComponents(node *layout.LayoutNode, registry *ComponentRegistry, cmds *[]tea.Cmd) {
	if node == nil {
		return
	}

	log.Trace("initializeLayoutComponents: Processing node %s (children: %d, component: %v)",
		node.ID, len(node.Children), node.Component != nil)

	// Initialize component if this node has an ID and no children (leaf node)
	// This is a regular component, not a layout container
	if node.ID != "" && node.Component == nil && len(node.Children) == 0 {
		compConfig := m.findComponentConfig(node.ID)
		log.Trace("initializeLayoutComponents: Found config for %s: %v", node.ID, compConfig != nil)
		if compConfig != nil {
			if err := m.initializeComponent(compConfig, registry, cmds); err != nil {
				log.Error("Failed to initialize component %s: %v", node.ID, err)
			} else {
				// Bind component instance to layout node
				if instance, exists := m.ComponentInstanceRegistry.Get(node.ID); exists {
					node.Component = instance
					log.Trace("Initialized and bound component %s (type: %s)", node.ID, compConfig.Type)
				} else {
					log.Warn("Component instance not found in registry for %s", node.ID)
				}
			}
		} else {
			log.Warn("Could not find config for component %s", node.ID)
		}
	}

	// Recursively initialize children
	for _, child := range node.Children {
		m.initializeLayoutComponents(child, registry, cmds)
	}
}

// findComponentConfig finds the component configuration by ID in the layout tree.
func (m *Model) findComponentConfig(id string) *Component {
	return findComponentInLayout(&m.Config.Layout, id)
}

// findComponentInLayout recursively searches for a component by ID.
func findComponentInLayout(l *Layout, id string) *Component {
	if l == nil {
		return nil
	}

	for _, child := range l.Children {
		if child.ID == id {
			return &child
		}

		if child.Type == "layout" {
			if nestedLayout, ok := child.Props["layout"].(*Layout); ok {
				if found := findComponentInLayout(nestedLayout, id); found != nil {
					return found
				}
			}
		}
	}

	return nil
}

// initializeComponent creates and registers a component instance.
func (m *Model) initializeComponent(comp *Component, registry *ComponentRegistry, cmds *[]tea.Cmd) error {
	if comp == nil || comp.Type == "" {
		return fmt.Errorf("component is nil or has empty type")
	}

	factory, exists := registry.GetComponentFactory(ComponentType(comp.Type))
	if !exists || factory == nil {
		return fmt.Errorf("unknown component type: %s", comp.Type)
	}

	props := m.resolveProps(comp)

	renderConfig := core.RenderConfig{
		Data:   props,
		Width:  m.Width,
		Height: m.Height,
	}

	componentInstance, isNew := m.ComponentInstanceRegistry.GetOrCreate(
		comp.ID,
		comp.Type,
		factory,
		renderConfig,
	)

	if comp.ID != "" && isInteractiveComponent(comp.Type) {
		if m.Components == nil {
			m.Components = make(map[string]*core.ComponentInstance)
		}

		m.Components[comp.ID] = componentInstance

		if isNew {
			log.Trace("InitializeComponents: Created new component instance %s (type: %s)", comp.ID, comp.Type)

			if subscriber, ok := componentInstance.Instance.(interface{ GetSubscribedMessageTypes() []string }); ok {
				messageTypes := subscriber.GetSubscribedMessageTypes()
				if len(messageTypes) > 0 {
					m.MessageSubscriptionManager.Subscribe(comp.ID, messageTypes)
					log.Trace("InitializeComponents: Registered message subscriptions for %s: %v", comp.ID, messageTypes)
				}
			}
		}

		if initCmd := componentInstance.Instance.Init(); initCmd != nil {
			*cmds = append(*cmds, initCmd)
		}
	} else {
		log.Trace("InitializeComponents: Reusing existing component instance %s (type: %s)", comp.ID, comp.Type)
	}

	return nil
}

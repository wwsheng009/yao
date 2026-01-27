package component

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yaoapp/kun/log"
	"github.com/yaoapp/yao/tui/tui/core"
	"github.com/yaoapp/yao/tui/tui/legacy/layout"
)

// LayoutConfig represents the layout configuration
type LayoutConfig struct {
	Direction string
	Children  []Component
	Padding   []int
}

// Model is the interface for the TUI model (to avoid import cycle)
type Model interface {
	GetComponentRegistry() *InstanceRegistry
	GetState() map[string]interface{}
	GetComponents() map[string]*core.ComponentInstance
}

// Initializer handles component initialization
type Initializer struct {
	model        Model
	registry     *InstanceRegistry
	layoutEngine *layout.Engine
	layoutRoot   *layout.LayoutNode
	renderer     *layout.Renderer
}

// NewInitializer creates a new component initializer
func NewInitializer(model Model) *Initializer {
	return &Initializer{
		model:    model,
		registry: model.GetComponentRegistry(),
	}
}

// Initialize initializes all components from the layout
func (i *Initializer) Initialize(layoutCfg *LayoutConfig, width, height int) []tea.Cmd {
	log.Trace("InitializeComponents: Starting component initialization")

	// Create layout tree from Config.Layout
	i.layoutRoot = i.createLayoutTree(layoutCfg, 0, make(map[string]int))
	log.Trace("InitializeComponents: Created layout tree with root ID: %s", i.layoutRoot.ID)

	// Initialize layout engine with the created tree
	i.layoutEngine = layout.NewEngine(&layout.LayoutConfig{
		Root: i.layoutRoot,
		WindowSize: &layout.WindowSize{
			Width:  width,
			Height: height,
		},
		PropsResolver: func(node *layout.LayoutNode) map[string]interface{} {
			if node.ID == "" {
				return node.Props
			}
			comp := i.findComponentConfig(node.ID)
			if comp != nil {
				return i.resolveProps(comp)
			}
			return node.Props
		},
	})
	log.Trace("InitializeComponents: Created layout engine, window: %dx%d", width, height)

	var allCmds []tea.Cmd

	// Initialize all components in the layout tree
	i.initializeLayoutComponents(i.layoutRoot, &allCmds)
	log.Trace("InitializeComponents: Initialized components, got %d commands", len(allCmds))

	// Initialize the Renderer with the LayoutEngine
	// Note: We can't use layout.NewRenderer directly since it needs the full Model
	// This will be set by the caller
	log.Trace("InitializeComponents: Component initialization complete")

	return allCmds
}

// SetRenderer sets the renderer (called by tea package to avoid import cycle)
func (i *Initializer) SetRenderer(renderer interface{}) {
	// The renderer would be set here
	// For now, we just log
	log.Trace("Initializer: Renderer set")
}

// GetLayoutEngine returns the layout engine
func (i *Initializer) GetLayoutEngine() *layout.Engine {
	return i.layoutEngine
}

// GetLayoutRoot returns the layout root
func (i *Initializer) GetLayoutRoot() *layout.LayoutNode {
	return i.layoutRoot
}

// GetRenderer returns the renderer
func (i *Initializer) GetRenderer() interface{} {
	return i.renderer
}

// createLayoutTree creates a LayoutNode tree from the LayoutConfig structure
func (i *Initializer) createLayoutTree(layoutCfg *LayoutConfig, depth int, typeCounters map[string]int) *layout.LayoutNode {
	const maxLayoutDepth = 50
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

	if depth == 0 {
		layout.ApplyStyle(builder.Current(),
			layout.WithWidth("flex"),
			layout.WithHeight("flex"),
		)
	}

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

	for idx := range layoutCfg.Children {
		child := &layoutCfg.Children[idx]

		// Check if this is a container component with children
		if child.Type == "layout" || len(child.Children) > 0 {
			nestedCounters := make(map[string]int)
			for k, v := range typeCounters {
				nestedCounters[k] = v
			}

			nestedDirection := "column"
			if child.Direction != "" {
				nestedDirection = child.Direction
			} else if child.Type == "row" || child.Type == "flex" {
				nestedDirection = "row"
			}

			oldFormatLayout := &LayoutConfig{
				Direction: nestedDirection,
				Children:  child.Children,
				Padding:   layoutCfg.Padding,
			}
			nestedNode := i.createLayoutTree(oldFormatLayout, depth+1, nestedCounters)
			if nestedNode != nil {
				if child.Width != nil {
					nestedNode.Style.Width = layout.NewSize(child.Width)
				}
				if child.Height != nil {
					nestedNode.Style.Height = layout.NewSize(child.Height)
				}
				if child.ID != "" {
					nestedNode.ID = child.ID
				}
				nestedNode.ComponentType = child.Type

				builder.AddNode(nestedNode)
			}
		} else {
			// Generate ID for components without one
			if child.ID == "" {
				count := typeCounters[child.Type]
				typeCounters[child.Type] = count + 1
				child.ID = fmt.Sprintf("comp_%s_%d", child.Type, count)
				log.Trace("createLayoutTree: Generated ID %s for component type %s", child.ID, child.Type)
			}

			// Regular component - create placeholder node
			componentNode := i.createComponentNode(*child)
			componentNode.Parent = builder.Current()
			builder.Current().Children = append(builder.Current().Children, componentNode)
		}
	}

	builder.Pop()
	return builder.Root()
}

// createComponentNode creates a layout node for a regular component
func (i *Initializer) createComponentNode(child Component) *layout.LayoutNode {
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
		Type:          layout.LayoutFlex,
		ComponentType: child.Type,
		Props:         child.Props,
		Style: &layout.LayoutStyle{
			Direction: layout.DirectionColumn,
			Width:     width,
			Height:    height,
			Position:  layout.PositionRelative,
		},
	}
	return node
}

// initializeLayoutComponents walks the layout tree and initializes component instances
func (i *Initializer) initializeLayoutComponents(node *layout.LayoutNode, cmds *[]tea.Cmd) {
	if node == nil {
		return
	}

	log.Trace("initializeLayoutComponents: Processing node %s (children: %d, component: %v)",
		node.ID, len(node.Children), node.Component != nil)

	// Initialize component if this node has no children (leaf node)
	if node.Component == nil && len(node.Children) == 0 {
		// Component initialization would happen here
		if node.ComponentType != "" && node.Props != nil {
			log.Trace("initializeLayoutComponents: Would initialize component type %s", node.ComponentType)
		}
	}

	// Recursively initialize children
	for _, child := range node.Children {
		i.initializeLayoutComponents(child, cmds)
	}
}

// findComponentConfig finds the component configuration by ID
func (i *Initializer) findComponentConfig(id string) *Component {
	// This would search the layout tree for the component
	return nil
}

// resolveProps resolves component properties
func (i *Initializer) resolveProps(comp *Component) map[string]interface{} {
	if comp.Props == nil {
		return make(map[string]interface{})
	}
	result := make(map[string]interface{})
	for k, v := range comp.Props {
		result[k] = v
	}
	return result
}

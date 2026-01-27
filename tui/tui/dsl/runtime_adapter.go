// Package dsl provides the DSL (Domain Specific Language) parser for TUI layouts.
//
// This file implements the adapter that converts DSL Nodes to runtime LayoutNodes.
package dsl

import (
	"fmt"

	"github.com/yaoapp/kun/log"
	tuiruntime "github.com/yaoapp/yao/tui/tui/runtime"
)

// ToLayoutNode converts a DSL Node tree to a runtime LayoutNode tree.
// This is the main entry point for converting parsed DSL to the runtime format.
func (n *Node) ToLayoutNode() *tuiruntime.LayoutNode {
	if n == nil {
		return nil
	}

	// Determine node type
	nodeType := mapDSLTypeToRuntime(n.Type)

	// Convert style
	style := n.convertStyle()

	// Create layout node
	layoutNode := tuiruntime.NewLayoutNode(n.ID, nodeType, style)

	// Copy props
	if n.Props != nil {
		layoutNode.Props = make(map[string]interface{})
		for k, v := range n.Props {
			layoutNode.Props[k] = v
		}
	}

	// Recursively convert children
	for _, child := range n.Children {
		childLayoutNode := child.ToLayoutNode()
		if childLayoutNode != nil {
			layoutNode.AddChild(childLayoutNode)
		}
	}

	return layoutNode
}

// ToLayoutTree converts the entire DSL config to a runtime LayoutNode tree.
func (c *Config) ToLayoutTree() *tuiruntime.LayoutNode {
	if c.Layout == nil {
		return nil
	}
	return c.Layout.ToLayoutNode()
}

// convertStyle converts DSL style properties to runtime Style.
func (n *Node) convertStyle() tuiruntime.Style {
	style := tuiruntime.NewStyle()

	// Apply style from NodeStyle if present
	if n.Style != nil {
		style = applyStyleFromNodeStyle(style, n.Style)
	}

	// Apply direct style properties (shorthand) only if no NodeStyle
	if n.Style == nil {
		style = applyDirectStyleProps(style, n)
	}

	return style
}

// applyStyleFromNodeStyle applies style from a NodeStyle object.
func applyStyleFromNodeStyle(style tuiruntime.Style, nodeStyle *NodeStyle) tuiruntime.Style {
	// Width
	if nodeStyle.Width != nil {
		width, _, _, _ := ParseSize(nodeStyle.Width)
		style.Width = &width
	}

	// Height
	if nodeStyle.Height != nil {
		height, _, _, _ := ParseSize(nodeStyle.Height)
		style.Height = &height
	}

	// FlexGrow
	if nodeStyle.FlexGrow > 0 {
		style = style.WithFlexGrow(nodeStyle.FlexGrow)
	}

	// Direction
	if nodeStyle.Direction != "" {
		style.Direction = mapDirection(nodeStyle.Direction)
	}

	// Padding
	if len(nodeStyle.Padding) > 0 {
		padding := toRuntimeInsets(nodeStyle.Padding)
		style = style.WithPadding(padding)
	}

	// Border
	if nodeStyle.Border != nil {
		border := toRuntimeInsets([]int{
			nodeStyle.Border.Top,
			nodeStyle.Border.Right,
			nodeStyle.Border.Bottom,
			nodeStyle.Border.Left,
		})
		style = style.WithBorder(border)
	}

	// Gap
	if nodeStyle.Gap > 0 {
		style = style.WithGap(nodeStyle.Gap)
	}

	// ZIndex
	if nodeStyle.ZIndex > 0 {
		style = style.WithZIndex(nodeStyle.ZIndex)
	}

	// Overflow
	if nodeStyle.Overflow != "" {
		style = style.WithOverflow(mapOverflow(nodeStyle.Overflow))
	}

	// AlignItems
	if nodeStyle.AlignItems != "" {
		style = style.WithAlignItems(mapAlign(nodeStyle.AlignItems))
	}

	// Justify
	if nodeStyle.Justify != "" {
		style = style.WithJustify(mapJustify(nodeStyle.Justify))
	}

	return style
}

// applyDirectStyleProps applies direct style properties from the node.
func applyDirectStyleProps(style tuiruntime.Style, node *Node) tuiruntime.Style {
	// Width
	if node.Width != nil {
		width, isPercent, _, _ := ParseSize(node.Width)
		style.Width = &width
		_ = isPercent // Already encoded in width value
	}

	// Height
	if node.Height != nil {
		height, isPercent, isFlex, _ := ParseSize(node.Height)
		if isFlex {
			style.FlexGrow = 1
		} else {
			style.Height = &height
		}
		_ = isPercent // Already encoded in height value
	}

	// FlexGrow
	if node.FlexGrow > 0 {
		style = style.WithFlexGrow(node.FlexGrow)
	}

	// Direction
	if node.Direction != "" {
		style.Direction = mapDirection(node.Direction)
	}

	// Padding
	if len(node.Padding) > 0 {
		padding := toRuntimeInsets(node.Padding)
		style = style.WithPadding(padding)
	}

	// Border
	borderSpec := node.Border
	if borderSpec == nil && node.BorderWidth != nil {
		// Handle different types for BorderWidth
		borderSpec = parseBorderFromWidth(node.BorderWidth)
	}
	if borderSpec != nil {
		border := toRuntimeInsets([]int{
			borderSpec.Top,
			borderSpec.Right,
			borderSpec.Bottom,
			borderSpec.Left,
		})
		style = style.WithBorder(border)
	}

	// Gap
	if node.Gap > 0 {
		style = style.WithGap(node.Gap)
	}

	// ZIndex
	if node.ZIndex > 0 {
		style = style.WithZIndex(node.ZIndex)
	}

	// Overflow
	if node.Overflow != "" {
		style = style.WithOverflow(mapOverflow(node.Overflow))
	}

	// AlignItems
	if node.AlignItems != "" {
		style = style.WithAlignItems(mapAlign(node.AlignItems))
	}

	// Justify
	if node.Justify != "" {
		style = style.WithJustify(mapJustify(node.Justify))
	}

	return style
}

// mapDSLTypeToRuntime maps DSL component types to runtime NodeTypes.
func mapDSLTypeToRuntime(dslType string) tuiruntime.NodeType {
	switch dslType {
	case "layout", "vertical", "horizontal":
		// For layout nodes, determine direction from Direction property
		return tuiruntime.NodeTypeFlex
	case "row":
		return tuiruntime.NodeTypeRow
	case "column":
		return tuiruntime.NodeTypeColumn
	default:
		// All component types are treated as custom (leaf nodes)
		return tuiruntime.NodeTypeCustom
	}
}

// mapDirection maps direction strings to runtime Direction.
func mapDirection(dir string) tuiruntime.Direction {
	switch dir {
	case "row", "horizontal":
		return tuiruntime.DirectionRow
	case "column", "vertical":
		return tuiruntime.DirectionColumn
	default:
		return tuiruntime.DirectionRow // Default
	}
}

// mapOverflow maps overflow strings to runtime Overflow.
func mapOverflow(overflow string) tuiruntime.Overflow {
	switch overflow {
	case "visible":
		return tuiruntime.OverflowVisible
	case "hidden":
		return tuiruntime.OverflowHidden
	case "scroll":
		return tuiruntime.OverflowScroll
	default:
		return tuiruntime.OverflowVisible // Default
	}
}

// mapAlign maps align strings to runtime Align.
func mapAlign(align string) tuiruntime.Align {
	switch align {
	case "start":
		return tuiruntime.AlignStart
	case "center":
		return tuiruntime.AlignCenter
	case "end":
		return tuiruntime.AlignEnd
	case "stretch":
		return tuiruntime.AlignStretch
	default:
		return tuiruntime.AlignStart // Default
	}
}

// mapJustify maps justify strings to runtime Justify.
func mapJustify(justify string) tuiruntime.Justify {
	switch justify {
	case "start":
		return tuiruntime.JustifyStart
	case "center":
		return tuiruntime.JustifyCenter
	case "end":
		return tuiruntime.JustifyEnd
	case "space-between":
		return tuiruntime.JustifySpaceBetween
	case "space-around":
		return tuiruntime.JustifySpaceAround
	case "space-evenly":
		return tuiruntime.JustifySpaceEvenly
	default:
		return tuiruntime.JustifyStart // Default
	}
}

// parseBorderFromWidth parses BorderWidth value to BorderSpec.
// Handles int, float64, []interface{}, and map[string]interface{}.
func parseBorderFromWidth(value interface{}) *BorderSpec {
	switch v := value.(type) {
	case int:
		return &BorderSpec{Top: v, Right: v, Bottom: v, Left: v}
	case float64:
		width := int(v)
		return &BorderSpec{Top: width, Right: width, Bottom: width, Left: width}
	case []interface{}:
		// Array format: [all] or [top,right,bottom,left]
		spec := BorderSpec{}
		switch len(v) {
		case 1:
			spec.Top = toInt(v[0])
			spec.Right = spec.Top
			spec.Bottom = spec.Top
			spec.Left = spec.Top
		case 2:
			spec.Top = toInt(v[0])
			spec.Bottom = spec.Top
			spec.Right = toInt(v[1])
			spec.Left = spec.Right
		case 3:
			spec.Top = toInt(v[0])
			spec.Right = toInt(v[1])
			spec.Bottom = toInt(v[2])
			spec.Left = spec.Right
		case 4:
			spec.Top = toInt(v[0])
			spec.Right = toInt(v[1])
			spec.Bottom = toInt(v[2])
			spec.Left = toInt(v[3])
		}
		return &spec
	case map[string]interface{}:
		return &BorderSpec{
			Top:    intFromMap(v, "top"),
			Right:  intFromMap(v, "right"),
			Bottom: intFromMap(v, "bottom"),
			Left:   intFromMap(v, "left"),
		}
	}
	return nil
}

// toRuntimeInsets converts a slice to tuiruntime.Insets.
func toRuntimeInsets(padding []int) tuiruntime.Insets {
	if len(padding) == 0 {
		return tuiruntime.Insets{}
	}

	var top, right, bottom, left int
	switch len(padding) {
	case 1:
		top, right, bottom, left = padding[0], padding[0], padding[0], padding[0]
	case 2:
		top, right, bottom, left = padding[0], padding[1], padding[0], padding[1]
	case 3:
		top, right, bottom, left = padding[0], padding[1], padding[2], padding[1]
	case 4:
		top, right, bottom, left = padding[0], padding[1], padding[2], padding[3]
	default:
		top, right, bottom, left = padding[0], padding[1], padding[2], padding[3]
	}

	return tuiruntime.Insets{
		Top:    top,
		Right:  right,
		Bottom: bottom,
		Left:   left,
	}
}

// BindComponents binds component instances to layout nodes.
// This is called after component instances are created.
func (c *Config) BindComponents(componentMap map[string]*tuiruntime.LayoutNode) {
	if c.Layout == nil {
		return
	}
	c.Layout.bindComponents(componentMap)
}

// bindComponents recursively binds component instances.
func (n *Node) bindComponents(componentMap map[string]*tuiruntime.LayoutNode) {
	if _, ok := componentMap[n.ID]; ok {
		// Component found, it will be linked elsewhere
		log.Trace("DSL: Component %s found in component map", n.ID)
	}

	for _, child := range n.Children {
		child.bindComponents(componentMap)
	}
}

// ValidateAndConvert validates the DSL config and converts to runtime LayoutNode.
// This is a convenience function that combines validation and conversion.
func ValidateAndConvert(data []byte, filename string) (*Config, *tuiruntime.LayoutNode, error) {
	// Parse
	cfg, err := ParseFile(data, filename)
	if err != nil {
		return nil, nil, fmt.Errorf("parse error: %w", err)
	}

	// Validate
	if err := cfg.Validate(); err != nil {
		return nil, nil, fmt.Errorf("validation error: %w", err)
	}

	// Assign IDs to nodes without IDs
	counter := 0
	cfg.Layout.AssignIDs("node", &counter)

	// Convert to runtime LayoutNode
	root := cfg.ToLayoutTree()

	return cfg, root, nil
}

// GetLayoutStats returns statistics about the layout tree.
func (c *Config) GetLayoutStats() LayoutStats {
	if c.Layout == nil {
		return LayoutStats{}
	}
	return c.Layout.collectStats()
}

// LayoutStats holds statistics about a layout tree.
type LayoutStats struct {
	TotalNodes     int
	ContainerNodes int
	LeafNodes      int
	MaxDepth       int
}

// collectStats recursively collects statistics about the node tree.
func (n *Node) collectStats() LayoutStats {
	stats := LayoutStats{
		TotalNodes: 1,
	}

	if isLayoutType(n.Type) {
		stats.ContainerNodes = 1
	} else {
		stats.LeafNodes = 1
	}

	maxChildDepth := 0
	for _, child := range n.Children {
		childStats := child.collectStats()
		stats.TotalNodes += childStats.TotalNodes
		stats.ContainerNodes += childStats.ContainerNodes
		stats.LeafNodes += childStats.LeafNodes
		if childStats.MaxDepth > maxChildDepth {
			maxChildDepth = childStats.MaxDepth
		}
	}
	stats.MaxDepth = maxChildDepth + 1

	return stats
}

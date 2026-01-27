// Package dsl provides the DSL (Domain Specific Language) parser for TUI layouts.
//
// It supports both JSON and YAML configuration formats and converts them
// into the runtime LayoutNode tree used by the new TUI runtime engine.
package dsl

import (
	"fmt"

	"github.com/yaoapp/yao/tui/tui/core"
)

// Config represents the parsed TUI configuration file (.tui.yao or .tui.json).
// This is the DSL-specific configuration that will be converted to runtime LayoutNode.
type Config struct {
	// ID is the unique identifier for this TUI (derived from file path)
	ID string `json:"id,omitempty" yaml:"id,omitempty"`

	// Name is the human-readable name of the TUI
	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	// Data contains the initial state data (supports mustache template binding)
	Data map[string]interface{} `json:"data,omitempty" yaml:"data,omitempty"`

	// OnLoad is the action to execute when TUI loads
	OnLoad *core.Action `json:"onLoad,omitempty" yaml:"onLoad,omitempty"`

	// Layout defines the UI structure (root layout node)
	Layout *Node `json:"layout,omitempty" yaml:"layout,omitempty"`

	// Bindings maps keyboard shortcuts to actions
	Bindings map[string]core.Action `json:"bindings,omitempty" yaml:"bindings,omitempty"`

	// LogLevel controls the verbosity of logging
	LogLevel string `json:"logLevel,omitempty" yaml:"logLevel,omitempty"`

	// AutoFocus enables automatic focus to the first focusable component
	AutoFocus *bool `json:"autoFocus,omitempty" yaml:"autoFocus,omitempty"`

	// NavigationMode defines how Tab/ShiftTab keys are handled
	NavigationMode string `json:"navigationMode,omitempty" yaml:"navigationMode,omitempty"`

	// TabCycles defines whether Tab navigation cycles through components
	TabCycles bool `json:"tabCycles,omitempty" yaml:"tabCycles,omitempty"`
}

// Node represents a UI node in the DSL.
// It can be a layout container (with children) or a leaf component.
type Node struct {
	// ID is the unique identifier for this node
	ID string `json:"id,omitempty" yaml:"id,omitempty"`

	// Type specifies the node type:
	// - Container types: "layout", "row", "column", "vertical", "horizontal"
	// - Component types: "header", "text", "menu", "progress", "input", etc.
	Type string `json:"type,omitempty" yaml:"type,omitempty"`

	// Props contains component-specific properties
	Props map[string]interface{} `json:"props,omitempty" yaml:"props,omitempty"`

	// Actions maps event names to actions for this component
	Actions map[string]core.Action `json:"actions,omitempty" yaml:"actions,omitempty"`

	// Children contains child nodes (for container nodes)
	Children []*Node `json:"children,omitempty" yaml:"children,omitempty"`

	// Style properties for layout
	Style *NodeStyle `json:"style,omitempty" yaml:"style,omitempty"`

	// Layout shortcut properties (when Style is not specified)
	Direction string      `json:"direction,omitempty" yaml:"direction,omitempty"` // "row", "column", "vertical", "horizontal"
	Width     interface{} `json:"width,omitempty" yaml:"width,omitempty"`       // number or "50%"
	Height    interface{} `json:"height,omitempty" yaml:"height,omitempty"`     // number or "100%" or "flex"
	Padding   []int       `json:"padding,omitempty" yaml:"padding,omitempty"`   // [top, right, bottom, left]

	// Border properties
	Border    *BorderSpec `json:"border,omitempty" yaml:"border,omitempty"`
	BorderWidth interface{} `json:"borderWidth,omitempty" yaml:"borderWidth,omitempty"` // number or [top, right, bottom, left]

	// Flex properties
	FlexGrow float64 `json:"flexGrow,omitempty" yaml:"flexGrow,omitempty"`

	// Alignment properties
	AlignItems string `json:"alignItems,omitempty" yaml:"alignItems,omitempty"` // "start", "center", "end", "stretch"
	Justify    string `json:"justify,omitempty" yaml:"justify,omitempty"`       // "start", "center", "end", "space-between", "space-around", "space-evenly"

	// Gap between children
	Gap int `json:"gap,omitempty" yaml:"gap,omitempty"`

	// Z-Index for layering
	ZIndex int `json:"zIndex,omitempty" yaml:"zIndex,omitempty"`

	// Overflow behavior
	Overflow string `json:"overflow,omitempty" yaml:"overflow,omitempty"` // "visible", "hidden", "scroll"

	// Bind specifies the state key to bind data from
	Bind string `json:"bind,omitempty" yaml:"bind,omitempty"`
}

// NodeStyle represents style properties for a node.
type NodeStyle struct {
	Width     interface{} `json:"width,omitempty" yaml:"width,omitempty"`
	Height    interface{} `json:"height,omitempty" yaml:"height,omitempty"`
	FlexGrow  float64     `json:"flexGrow,omitempty" yaml:"flexGrow,omitempty"`
	Direction string      `json:"direction,omitempty" yaml:"direction,omitempty"`
	Padding   []int       `json:"padding,omitempty" yaml:"padding,omitempty"`
	Border    *BorderSpec `json:"border,omitempty" yaml:"border,omitempty"`
	Gap       int         `json:"gap,omitempty" yaml:"gap,omitempty"`
	ZIndex    int         `json:"zIndex,omitempty" yaml:"zIndex,omitempty"`
	Overflow  string      `json:"overflow,omitempty" yaml:"overflow,omitempty"`
	AlignItems string     `json:"alignItems,omitempty" yaml:"alignItems,omitempty"`
	Justify    string     `json:"justify,omitempty" yaml:"justify,omitempty"`
}

// BorderSpec represents border specification.
type BorderSpec struct {
	Top    int `json:"top,omitempty" yaml:"top,omitempty"`
	Right  int `json:"right,omitempty" yaml:"right,omitempty"`
	Bottom int `json:"bottom,omitempty" yaml:"bottom,omitempty"`
	Left   int `json:"left,omitempty" yaml:"left,omitempty"`
}

// StringSize represents a size that can be a number, percentage, or "flex".
type StringSize string

const (
	// SizeFlex indicates the node should grow to fill available space
	SizeFlex StringSize = "flex"
	// SizeAuto indicates the size should be determined by content
	SizeAuto StringSize = "auto"
)

// ParseSize parses a size value from interface{}.
// Returns (size, isPercent, isFlex, isAuto).
func ParseSize(value interface{}) (int, bool, bool, bool) {
	if value == nil {
		return 0, false, false, true // Default to auto
	}

	switch v := value.(type) {
	case float64:
		return int(v), false, false, false
	case int:
		return v, false, false, false
	case string:
		if v == "flex" {
			return 0, false, true, false
		}
		if v == "auto" {
			return 0, false, false, true
		}
		// Parse percentage (e.g., "50%", "100%")
		if len(v) > 0 && v[len(v)-1] == '%' {
			percent := 0
			if _, err := fmt.Sscanf(v, "%d%%", &percent); err == nil {
				// Encode percentage as negative value for runtime
				// -50 means 50%
				return -percent, true, false, false
			}
		}
	}

	return 0, false, false, true // Default to auto
}

// ParseBorder parses a border specification.
// Can be a single number, array of numbers, or BorderSpec.
func ParseBorder(value interface{}) BorderSpec {
	if value == nil {
		return BorderSpec{}
	}

	switch v := value.(type) {
	case int:
		return BorderSpec{Top: v, Right: v, Bottom: v, Left: v}
	case float64:
		width := int(v)
		return BorderSpec{Top: width, Right: width, Bottom: width, Left: width}
	case []interface{}:
		result := BorderSpec{}
		switch len(v) {
		case 1:
			result.Top = toInt(v[0])
			result.Right = result.Top
			result.Bottom = result.Top
			result.Left = result.Top
		case 2:
			result.Top = toInt(v[0])
			result.Bottom = result.Top
			result.Right = toInt(v[1])
			result.Left = result.Right
		case 3:
			result.Top = toInt(v[0])
			result.Right = toInt(v[1])
			result.Bottom = toInt(v[2])
			result.Left = result.Right
		case 4:
			result.Top = toInt(v[0])
			result.Right = toInt(v[1])
			result.Bottom = toInt(v[2])
			result.Left = toInt(v[3])
		}
		return result
	case map[string]interface{}:
		return BorderSpec{
			Top:    intFromMap(v, "top"),
			Right:  intFromMap(v, "right"),
			Bottom: intFromMap(v, "bottom"),
			Left:   intFromMap(v, "left"),
		}
	}

	return BorderSpec{}
}

// toInt converts an interface{} to int.
func toInt(value interface{}) int {
	switch v := value.(type) {
	case int:
		return v
	case float64:
		return int(v)
	case string:
		// Try to parse string as int
		var i int
		if _, err := fmt.Sscanf(v, "%d", &i); err == nil {
			return i
		}
	}
	return 0
}

// intFromMap gets an int value from a map with a default of 0.
func intFromMap(m map[string]interface{}, key string) int {
	if val, ok := m[key]; ok {
		return toInt(val)
	}
	return 0
}

// ToBorderStyle converts BorderSpec to runtime Insets.
func (b BorderSpec) ToBorderStyle() [4]int {
	return [4]int{b.Top, b.Right, b.Bottom, b.Left}
}

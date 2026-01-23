package dsl

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/yaoapp/yao/tui/runtime"
	"github.com/yaoapp/yao/tui/runtime/registry"
	"github.com/yaoapp/yao/tui/ui/components"
)

// Factory creates native Runtime components from DSL configuration.
// This bridges the gap between DSL YAML/JSON configs and native components.
type Factory struct {
	registry *registry.Registry
}

// NewFactory creates a new DSL component factory.
func NewFactory() *Factory {
	return &Factory{
		registry: registry.DefaultRegistry,
	}
}

// NewFactoryWithRegistry creates a factory with a custom registry.
func NewFactoryWithRegistry(reg *registry.Registry) *Factory {
	return &Factory{
		registry: reg,
	}
}

// ComponentConfig represents a component definition from DSL.
type ComponentConfig struct {
	ID     string                 // Component ID
	Type   string                 // Component type (e.g., "input", "button")
	Props  map[string]interface{} // Component properties
	Width  interface{}            // Width specification (number or "flex")
	Height interface{}            // Height specification (number or "flex")
}

// Create creates a native Runtime component from DSL config.
func (f *Factory) Create(config *ComponentConfig) (runtime.Component, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}

	// Create component instance via registry
	component, err := f.registry.Create(config.Type)
	if err != nil {
		return nil, fmt.Errorf("failed to create component type %s: %w", config.Type, err)
	}

	// Apply props to component
	if err := f.ApplyProps(component, config); err != nil {
		return nil, fmt.Errorf("failed to apply props for %s: %w", config.ID, err)
	}

	return component, nil
}

// CreateInput creates an Input component with props.
func (f *Factory) CreateInput(config *ComponentConfig) (*components.InputComponent, error) {
	comp, err := f.Create(config)
	if err != nil {
		return nil, err
	}
	input, ok := comp.(*components.InputComponent)
	if !ok {
		return nil, fmt.Errorf("created component is not an InputComponent")
	}
	return input, nil
}

// CreateButton creates a Button component with props.
func (f *Factory) CreateButton(config *ComponentConfig) (*components.ButtonComponent, error) {
	comp, err := f.Create(config)
	if err != nil {
		return nil, err
	}
	button, ok := comp.(*components.ButtonComponent)
	if !ok {
		return nil, fmt.Errorf("created component is not a ButtonComponent")
	}
	return button, nil
}

// CreateRow creates a Row component with props.
func (f *Factory) CreateRow(config *ComponentConfig) (*components.RowComponent, error) {
	comp, err := f.Create(config)
	if err != nil {
		return nil, err
	}
	row, ok := comp.(*components.RowComponent)
	if !ok {
		return nil, fmt.Errorf("created component is not a RowComponent")
	}
	return row, nil
}

// CreateColumn creates a Column component with props.
func (f *Factory) CreateColumn(config *ComponentConfig) (*components.ColumnComponent, error) {
	comp, err := f.Create(config)
	if err != nil {
		return nil, err
	}
	column, ok := comp.(*components.ColumnComponent)
	if !ok {
		return nil, fmt.Errorf("created component is not a ColumnComponent")
	}
	return column, nil
}

// CreateFlex creates a Flex component with props.
func (f *Factory) CreateFlex(config *ComponentConfig) (*components.FlexComponent, error) {
	comp, err := f.Create(config)
	if err != nil {
		return nil, err
	}
	flex, ok := comp.(*components.FlexComponent)
	if !ok {
		return nil, fmt.Errorf("created component is not a FlexComponent")
	}
	return flex, nil
}

// CreateText creates a Text component with props.
func (f *Factory) CreateText(config *ComponentConfig) (*components.TextComponent, error) {
	comp, err := f.Create(config)
	if err != nil {
		return nil, err
	}
	text, ok := comp.(*components.TextComponent)
	if !ok {
		return nil, fmt.Errorf("created component is not a TextComponent")
	}
	return text, nil
}

// CreateHeader creates a Header component with props.
func (f *Factory) CreateHeader(config *ComponentConfig) (*components.HeaderComponent, error) {
	comp, err := f.Create(config)
	if err != nil {
		return nil, err
	}
	header, ok := comp.(*components.HeaderComponent)
	if !ok {
		return nil, fmt.Errorf("created component is not a HeaderComponent")
	}
	return header, nil
}

// CreateFooter creates a Footer component with props.
func (f *Factory) CreateFooter(config *ComponentConfig) (*components.FooterComponent, error) {
	comp, err := f.Create(config)
	if err != nil {
		return nil, err
	}
	footer, ok := comp.(*components.FooterComponent)
	if !ok {
		return nil, fmt.Errorf("created component is not a FooterComponent")
	}
	return footer, nil
}

// ApplyProps applies DSL props to a native component.
func (f *Factory) ApplyProps(component runtime.Component, config *ComponentConfig) error {
	props := config.Props
	if props == nil {
		props = make(map[string]interface{})
	}

	// Set ID if provided
	if config.ID != "" {
		if idSetter, ok := component.(interface{ WithID(string) }); ok {
			idSetter.WithID(config.ID)
		}
	}

	// Apply props based on component type
	switch comp := component.(type) {
	case *components.InputComponent:
		f.applyInputProps(comp, props)
	case *components.ButtonComponent:
		f.applyButtonProps(comp, props)
	case *components.RowComponent:
		f.applyRowProps(comp, props)
	case *components.ColumnComponent:
		f.applyColumnProps(comp, props)
	case *components.FlexComponent:
		f.applyFlexProps(comp, props)
	case *components.TextComponent:
		f.applyTextProps(comp, props)
	case *components.HeaderComponent:
		f.applyHeaderProps(comp, props)
	case *components.FooterComponent:
		f.applyFooterProps(comp, props)
	}

	return nil
}

// applyInputProps applies props to Input components.
func (f *Factory) applyInputProps(comp *components.InputComponent, props map[string]interface{}) {
	if placeholder, ok := props["placeholder"].(string); ok {
		comp.WithPlaceholder(placeholder)
	}
	if value, ok := props["value"].(string); ok {
		comp.WithValue(value)
	}
	if width, ok := props["width"].(int); ok {
		comp.WithWidth(width)
	}
	if maxLength, ok := props["maxLength"].(int); ok {
		comp.WithMaxLength(maxLength)
	}
	// Handle onChange callback
	if onChange, ok := props["onChange"]; ok {
		if fn, ok := onChange.(func(string)); ok {
			comp.WithOnChange(fn)
		}
	}
	if onEnter, ok := props["onEnter"]; ok {
		if fn, ok := onEnter.(func(string)); ok {
			comp.WithOnEnter(fn)
		}
	}
}

// applyButtonProps applies props to Button components.
func (f *Factory) applyButtonProps(comp *components.ButtonComponent, props map[string]interface{}) {
	if label, ok := props["label"].(string); ok {
		comp.WithLabel(label)
	}
	if disabled, ok := props["disabled"].(bool); ok {
		comp.WithDisabled(disabled)
	}
	// Handle onClick callback
	if onClick, ok := props["onClick"]; ok {
		if fn, ok := onClick.(func()); ok {
			comp.WithOnClick(fn)
		}
	}
	// Handle style props
	if style, ok := props["style"].(map[string]interface{}); ok {
		if normalStyle, ok := style["normal"]; ok {
			if s, ok := normalStyle.(lipgloss.Style); ok {
				comp.WithNormalStyle(s)
			}
		}
		if focusedStyle, ok := style["focused"]; ok {
			if s, ok := focusedStyle.(lipgloss.Style); ok {
				comp.WithFocusedStyle(s)
			}
		}
		if disabledStyle, ok := style["disabled"]; ok {
			if s, ok := disabledStyle.(lipgloss.Style); ok {
				comp.WithDisabledStyle(s)
			}
		}
	}
}

// applyRowProps applies props to Row components.
func (f *Factory) applyRowProps(comp *components.RowComponent, props map[string]interface{}) {
	if gap, ok := props["gap"].(int); ok {
		comp.WithGap(gap)
	}
	if spacing, ok := props["spacing"].(int); ok {
		comp.WithSpacing(spacing)
	}
	// Handle padding: [top, right, bottom, left] or single value
	if padding, ok := props["padding"]; ok {
		if p, ok := parsePadding(padding); ok {
			comp.WithPadding(p[0], p[1], p[2], p[3])
		}
	}
	// Handle align
	if align, ok := props["align"].(string); ok {
		if a, ok := parseAlign(align); ok {
			comp.WithAlign(a)
		}
	}
	// Handle justify
	if justify, ok := props["justify"].(string); ok {
		if j, ok := parseJustify(justify); ok {
			comp.WithJustify(j)
		}
	}
}

// applyColumnProps applies props to Column components.
func (f *Factory) applyColumnProps(comp *components.ColumnComponent, props map[string]interface{}) {
	if gap, ok := props["gap"].(int); ok {
		comp.WithGap(gap)
	}
	if spacing, ok := props["spacing"].(int); ok {
		comp.WithSpacing(spacing)
	}
	// Handle padding: [top, right, bottom, left] or single value
	if padding, ok := props["padding"]; ok {
		if p, ok := parsePadding(padding); ok {
			comp.WithPadding(p[0], p[1], p[2], p[3])
		}
	}
	// Handle align
	if align, ok := props["align"].(string); ok {
		if a, ok := parseAlign(align); ok {
			comp.WithAlign(a)
		}
	}
	// Handle justify
	if justify, ok := props["justify"].(string); ok {
		if j, ok := parseJustify(justify); ok {
			comp.WithJustify(j)
		}
	}
}

// applyFlexProps applies props to Flex components.
func (f *Factory) applyFlexProps(comp *components.FlexComponent, props map[string]interface{}) {
	// Apply gap/spacing
	if gap, ok := props["gap"].(int); ok {
		comp.WithGap(gap)
	}
	if spacing, ok := props["spacing"].(int); ok {
		comp.WithSpacing(spacing)
	}
	// Handle padding: [top, right, bottom, left] or single value
	if padding, ok := props["padding"]; ok {
		if p, ok := parsePadding(padding); ok {
			comp.WithPadding(p[0], p[1], p[2], p[3])
		}
	}
	// Handle align
	if align, ok := props["align"].(string); ok {
		if a, ok := parseAlign(align); ok {
			comp.WithAlign(a)
		}
	}
	// Handle justify
	if justify, ok := props["justify"].(string); ok {
		if j, ok := parseJustify(justify); ok {
			comp.WithJustify(j)
		}
	}

	// Flex-specific props
	if direction, ok := props["direction"].(string); ok {
		switch direction {
		case "row", "horizontal":
			comp.WithRow()
		case "column", "vertical":
			comp.WithColumn()
		}
	}
	if wrap, ok := props["wrap"].(bool); ok {
		comp.WithWrap(wrap)
	}
}

// applyTextProps applies props to Text components.
func (f *Factory) applyTextProps(comp *components.TextComponent, props map[string]interface{}) {
	if content, ok := props["content"].(string); ok {
		comp.WithContent(content)
	}
	if text, ok := props["text"].(string); ok {
		comp.WithContent(text)
	}
}

// applyHeaderProps applies props to Header components.
func (f *Factory) applyHeaderProps(comp *components.HeaderComponent, props map[string]interface{}) {
	// Content/Title
	if title, ok := props["title"].(string); ok {
		comp.WithContent(title)
	}
	if content, ok := props["content"].(string); ok {
		comp.WithContent(content)
	}

	// Alignment
	if align, ok := props["align"].(string); ok {
		comp.WithAlign(align)
	}

	// Colors
	if color, ok := props["color"].(string); ok {
		comp.WithColor(color)
	}
	if background, ok := props["background"].(string); ok {
		comp.WithBackground(background)
	}

	// Text decorations
	if bold, ok := props["bold"].(bool); ok {
		comp.WithBold(bold)
	}
	if italic, ok := props["italic"].(bool); ok {
		comp.WithItalic(italic)
	}
	if underline, ok := props["underline"].(bool); ok {
		comp.WithUnderline(underline)
	}

	// Padding
	if padding, ok := props["padding"]; ok {
		if p, ok := parsePadding(padding); ok {
			comp.WithPadding(p[0], p[1], p[2], p[3])
		}
	}
}

// applyFooterProps applies props to Footer components.
func (f *Factory) applyFooterProps(comp *components.FooterComponent, props map[string]interface{}) {
	// Content/Text
	if text, ok := props["text"].(string); ok {
		comp.WithContent(text)
	}
	if content, ok := props["content"].(string); ok {
		comp.WithContent(content)
	}

	// Alignment
	if align, ok := props["align"].(string); ok {
		comp.WithAlign(align)
	}

	// Colors
	if color, ok := props["color"].(string); ok {
		comp.WithColor(color)
	}
	if background, ok := props["background"].(string); ok {
		comp.WithBackground(background)
	}

	// Text decorations
	if bold, ok := props["bold"].(bool); ok {
		comp.WithBold(bold)
	}
	if italic, ok := props["italic"].(bool); ok {
		comp.WithItalic(italic)
	}
	if underline, ok := props["underline"].(bool); ok {
		comp.WithUnderline(underline)
	}

	// Padding
	if padding, ok := props["padding"]; ok {
		if p, ok := parsePadding(padding); ok {
			comp.WithPadding(p[0], p[1], p[2], p[3])
		}
	}

	// Margin
	if marginTop, ok := props["marginTop"].(int); ok {
		if marginBottom, ok2 := props["marginBottom"].(int); ok2 {
			comp.WithMargin(marginTop, marginBottom)
		} else {
			comp.WithMargin(marginTop, 0)
		}
	}
	if marginBottom, ok := props["marginBottom"].(int); ok {
		comp.WithMargin(0, marginBottom)
	}
}

// parseAlign parses align string to runtime.Align.
func parseAlign(s string) (runtime.Align, bool) {
	switch s {
	case "start", "left", "top":
		return runtime.AlignStart, true
	case "center", "middle":
		return runtime.AlignCenter, true
	case "end", "right", "bottom":
		return runtime.AlignEnd, true
	case "stretch":
		return runtime.AlignStretch, true
	}
	return "", false
}

// parseJustify parses justify string to runtime.Justify.
func parseJustify(s string) (runtime.Justify, bool) {
	switch s {
	case "start", "left", "top":
		return runtime.JustifyStart, true
	case "center", "middle":
		return runtime.JustifyCenter, true
	case "end", "right", "bottom":
		return runtime.JustifyEnd, true
	case "space-between":
		return runtime.JustifySpaceBetween, true
	case "space-around":
		return runtime.JustifySpaceAround, true
	case "space-evenly":
		return runtime.JustifySpaceEvenly, true
	}
	return "", false
}

// parsePadding parses padding from various formats to [top, right, bottom, left].
// Supports: int, []int, []interface{}
func parsePadding(padding interface{}) ([4]int, bool) {
	var result [4]int

	switch p := padding.(type) {
	case int:
		result = [4]int{p, p, p, p}
		return result, true
	case []int:
		if len(p) == 1 {
			result = [4]int{p[0], p[0], p[0], p[0]}
		} else if len(p) >= 4 {
			result = [4]int{p[0], p[1], p[2], p[3]}
		} else {
			return result, false
		}
		return result, true
	case []interface{}:
		if len(p) == 1 {
			if v, ok := p[0].(int); ok {
				result = [4]int{v, v, v, v}
				return result, true
			}
		} else if len(p) >= 4 {
			v0, ok0 := p[0].(int)
			v1, ok1 := p[1].(int)
			v2, ok2 := p[2].(int)
			v3, ok3 := p[3].(int)
			if ok0 && ok1 && ok2 && ok3 {
				result = [4]int{v0, v1, v2, v3}
				return result, true
			}
		}
		return result, false
	}

	return result, false
}

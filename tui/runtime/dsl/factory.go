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
	case *components.ListComponent:
		f.applyListProps(comp, props)
	case *components.TableComponent:
		f.applyTableProps(comp, props)
	case *components.FormComponent:
		f.applyFormProps(comp, props)
	case *components.TextareaComponent:
		f.applyTextareaProps(comp, props)
	case *components.ProgressComponent:
		f.applyProgressProps(comp, props)
	case *components.SpinnerComponent:
		f.applySpinnerProps(comp, props)
	case *components.ModalComponent:
		f.applyModalProps(comp, props)
	case *components.TabsComponent:
		f.applyTabsProps(comp, props)
	case *components.TreeComponent:
		f.applyTreeProps(comp, props)
	case *components.SplitPaneComponent:
		f.applySplitPaneProps(comp, props)
	case *components.ContextMenuComponent:
		f.applyContextMenuProps(comp, props)
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
	if width, ok := props["width"].(float64); ok {
		comp.WithWidth(int(width))
	}
	if maxLength, ok := props["maxLength"].(int); ok {
		comp.WithMaxLength(maxLength)
	}
	if prompt, ok := props["prompt"].(string); ok {
		comp.WithPrompt(prompt)
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

	// Colors - convert color names to ANSI codes
	if color, ok := props["color"].(string); ok {
		comp.WithColor(ColorNameToANSI(color))
	}
	if background, ok := props["background"].(string); ok {
		comp.WithBackground(ColorNameToANSI(background))
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

	// Colors - convert color names to ANSI codes
	if color, ok := props["color"].(string); ok {
		comp.WithColor(ColorNameToANSI(color))
	}
	if background, ok := props["background"].(string); ok {
		comp.WithBackground(ColorNameToANSI(background))
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

// applyListProps applies props to List components.
func (f *Factory) applyListProps(comp *components.ListComponent, props map[string]interface{}) {
	// Title
	if title, ok := props["title"].(string); ok {
		comp.WithTitle(title)
	}

	// Dimensions
	if width, ok := props["width"].(int); ok {
		comp.WithWidth(width)
	}
	if width, ok := props["width"].(float64); ok {
		comp.WithWidth(int(width))
	}
	if height, ok := props["height"].(int); ok {
		comp.WithHeight(height)
	}
	if height, ok := props["height"].(float64); ok {
		comp.WithHeight(int(height))
	}

	// Display options
	if showTitle, ok := props["showTitle"].(bool); ok {
		comp.WithShowTitle(showTitle)
	}
	if showStatusBar, ok := props["showStatusBar"].(bool); ok {
		comp.WithShowStatusBar(showStatusBar)
	}
	if showFilter, ok := props["showFilter"].(bool); ok {
		comp.WithShowFilter(showFilter)
	}
	if filteringEnabled, ok := props["filteringEnabled"].(bool); ok {
		comp.WithFilteringEnabled(filteringEnabled)
	}

	// Colors - convert color names to ANSI codes
	if color, ok := props["color"].(string); ok {
		comp.WithColor(ColorNameToANSI(color))
	}
	if background, ok := props["background"].(string); ok {
		comp.WithBackground(ColorNameToANSI(background))
	}

	// Padding
	if padding, ok := props["padding"]; ok {
		if p, ok := parsePadding(padding); ok {
			comp.WithPadding(p[0], p[1], p[2], p[3])
		}
	}

	// Items - convert from various formats
	if itemsData, ok := props["items"]; ok {
		items := f.ParseListItems(itemsData)
		comp.WithItems(items)
	}

	// Bind data support
	if bindData, ok := props["__bind_data"]; ok {
		items := f.ParseListItems(bindData)
		comp.WithItems(items)
	}
}

// ParseListItems parses items from various formats.
func (f *Factory) ParseListItems(itemsData interface{}) []components.RuntimeListItem {
	var items []components.RuntimeListItem

	switch data := itemsData.(type) {
	case []interface{}:
		items = make([]components.RuntimeListItem, 0, len(data))
		for _, itemData := range data {
			if itemMap, ok := itemData.(map[string]interface{}); ok {
				item := components.RuntimeListItem{}

				// Extract title
				var title string
				if titleStr, ok := itemMap["title"].(string); ok {
					title = titleStr
				} else if titleInt, ok := itemMap["title"].(int); ok {
					title = fmt.Sprintf("%d", titleInt)
				} else if titleFloat, ok := itemMap["title"].(float64); ok {
					title = fmt.Sprintf("%d", int(titleFloat))
				}
				item.SetTitle(title)

				// Extract description
				if desc, ok := itemMap["description"].(string); ok {
					item.SetDescription(desc)
				} else if desc, ok := itemMap["desc"].(string); ok {
					item.SetDescription(desc)
				}

				// Extract value
				if value, ok := itemMap["value"]; ok {
					item.Value = value
				} else {
					item.Value = title
				}

				// Extract disabled
				if disabled, ok := itemMap["disabled"].(bool); ok {
					item.Disabled = disabled
				}

				items = append(items, item)
			} else {
				// Simple string or primitive value
				text := fmt.Sprintf("%v", itemData)
				item := components.NewRuntimeListItem(text, "")
				item.Value = itemData
				items = append(items, item)
			}
		}
	case []string:
		items = make([]components.RuntimeListItem, 0, len(data))
		for _, str := range data {
			item := components.NewRuntimeListItem(str, "")
			items = append(items, item)
		}
	}

	return items
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
	return runtime.JustifyStart, false
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

// applyTableProps applies props to Table components.
func (f *Factory) applyTableProps(comp *components.TableComponent, props map[string]interface{}) {
	// Dimensions
	if width, ok := props["width"].(int); ok {
		comp.WithWidth(width)
	}
	if width, ok := props["width"].(float64); ok {
		comp.WithWidth(int(width))
	}
	if height, ok := props["height"].(int); ok {
		comp.WithHeight(height)
	}
	if height, ok := props["height"].(float64); ok {
		comp.WithHeight(int(height))
	}

	// Display options
	if focused, ok := props["focused"].(bool); ok {
		comp.WithFocused(focused)
	}
	if showBorder, ok := props["showBorder"].(bool); ok {
		comp.WithShowBorder(showBorder)
	}

	// Header styles - convert color names to ANSI codes
	if headerColor, ok := props["headerColor"].(string); ok {
		comp.WithHeaderColor(ColorNameToANSI(headerColor))
	}
	if headerBackground, ok := props["headerBackground"].(string); ok {
		comp.WithHeaderBackground(ColorNameToANSI(headerBackground))
	}
	if headerBold, ok := props["headerBold"].(bool); ok {
		comp.WithHeaderBold(headerBold)
	}

	// Cell styles - convert color names to ANSI codes
	if cellColor, ok := props["cellColor"].(string); ok {
		comp.WithCellColor(ColorNameToANSI(cellColor))
	}
	if cellBackground, ok := props["cellBackground"].(string); ok {
		comp.WithCellBackground(ColorNameToANSI(cellBackground))
	}

	// Selected row styles - convert color names to ANSI codes
	if selectedColor, ok := props["selectedColor"].(string); ok {
		comp.WithSelectedColor(ColorNameToANSI(selectedColor))
	}
	if selectedBackground, ok := props["selectedBackground"].(string); ok {
		comp.WithSelectedBackground(ColorNameToANSI(selectedBackground))
	}
	if selectedBold, ok := props["selectedBold"].(bool); ok {
		comp.WithSelectedBold(selectedBold)
	}

	// Border style - convert color name to ANSI code
	if borderColor, ok := props["borderColor"].(string); ok {
		comp.WithBorderColor(ColorNameToANSI(borderColor))
	}

	// Store bind data for later processing (after columns are set)
	var bindData interface{}
	if bd, ok := props["__bind_data"]; ok {
		bindData = bd
	}

	// Columns
	if columnsData, ok := props["columns"]; ok {
		columns := f.ParseTableColumns(columnsData)
		comp.WithColumns(columns)
	}

	// Rows/Data
	if rowsData, ok := props["rows"]; ok {
		rows := f.ParseTableRows(rowsData)
		comp.WithRows(rows)
	}
	if dataData, ok := props["data"]; ok {
		if dataArray, ok := dataData.([][]interface{}); ok {
			comp.WithData(dataArray)
		} else if dataArray, ok := dataData.([]interface{}); ok {
			// Convert []interface{} to [][]interface{}
			converted := make([][]interface{}, 0, len(dataArray))
			for _, rowIntf := range dataArray {
				if rowSlice, ok := rowIntf.([]interface{}); ok {
					converted = append(converted, rowSlice)
				}
			}
			comp.WithData(converted)
		}
	}

	// Bind data support - handle both 2D arrays and object arrays
	// This is processed AFTER columns are set, so we can use column keys for object arrays
	if bindData != nil {
		if objArray, ok := bindData.([]interface{}); ok {
			// Handle []map[string]interface{} format (object array)
			// Convert to [][]interface{} using column keys
			converted := f.ConvertObjectArrayToTableData(objArray, comp)
			if len(converted) > 0 {
				comp.WithData(converted)
			}
		} else if dataArray, ok := bindData.([][]interface{}); ok {
			comp.WithData(dataArray)
		}
	}

	// Padding
	if padding, ok := props["padding"]; ok {
		if p, ok := parsePadding(padding); ok {
			comp.WithPadding(p[0], p[1], p[2], p[3])
		}
	}
}

// ParseTableColumns parses columns from DSL.
func (f *Factory) ParseTableColumns(columnsData interface{}) []components.RuntimeColumn {
	var columns []components.RuntimeColumn

	switch data := columnsData.(type) {
	case []interface{}:
		columns = make([]components.RuntimeColumn, 0, len(data))
		for _, colData := range data {
			if colMap, ok := colData.(map[string]interface{}); ok {
				col := components.RuntimeColumn{}

				// Extract key
				if key, ok := colMap["key"].(string); ok {
					col.Key = key
				}
				// Extract title
				if title, ok := colMap["title"].(string); ok {
					col.Title = title
				}
				// Extract width
				if width, ok := colMap["width"].(int); ok {
					col.Width = width
				} else if width, ok := colMap["width"].(float64); ok {
					col.Width = int(width)
				}

				columns = append(columns, col)
			}
		}
	}

	return columns
}

// ParseTableRows parses rows from DSL.
func (f *Factory) ParseTableRows(rowsData interface{}) []components.RuntimeTableRow {
	var rows []components.RuntimeTableRow

	switch data := rowsData.(type) {
	case []interface{}:
		rows = make([]components.RuntimeTableRow, 0, len(data))
		for _, rowData := range data {
			if rowSlice, ok := rowData.([]interface{}); ok {
				row := components.RuntimeTableRow{}
				cells := make([]components.RuntimeTableCell, 0, len(rowSlice))
				for _, cellValue := range rowSlice {
					cells = append(cells, components.RuntimeTableCell{Value: cellValue})
				}
				row.Cells = cells
				rows = append(rows, row)
			}
		}
	}

	return rows
}

// applyFormProps applies props to Form components.
func (f *Factory) applyFormProps(comp *components.FormComponent, props map[string]interface{}) {
	// Dimensions
	if width, ok := props["width"].(int); ok {
		comp.WithWidth(width)
	}
	if width, ok := props["width"].(float64); ok {
		comp.WithWidth(int(width))
	}
	if height, ok := props["height"].(int); ok {
		comp.WithHeight(height)
	}
	if height, ok := props["height"].(float64); ok {
		comp.WithHeight(int(height))
	}

	// Title and description
	if title, ok := props["title"].(string); ok {
		comp.WithTitle(title)
	}
	if description, ok := props["description"].(string); ok {
		comp.WithDescription(description)
	}

	// Button labels
	if submitLabel, ok := props["submitLabel"].(string); ok {
		comp.WithSubmitLabel(submitLabel)
	}
	if cancelLabel, ok := props["cancelLabel"].(string); ok {
		comp.WithCancelLabel(cancelLabel)
	}

	// Fields
	if fieldsData, ok := props["fields"]; ok {
		fields := f.ParseFormFields(fieldsData)
		comp.WithFields(fields)
	}

	// Initial values
	if valuesData, ok := props["values"]; ok {
		if valuesMap, ok := valuesData.(map[string]interface{}); ok {
			for key, value := range valuesMap {
				if valueStr, ok := value.(string); ok {
					comp.SetFieldValue(key, valueStr)
				}
			}
		}
	}
}

// ParseFormFields parses form fields from DSL data.
func (f *Factory) ParseFormFields(fieldsData interface{}) []components.FormField {
	var fields []components.FormField

	switch data := fieldsData.(type) {
	case []interface{}:
		fields = make([]components.FormField, 0, len(data))
		for _, fieldData := range data {
			if fieldMap, ok := fieldData.(map[string]interface{}); ok {
				field := components.FormField{}

				// Extract type
				if fieldType, ok := fieldMap["type"].(string); ok {
					field.Type = fieldType
				} else {
					field.Type = "input" // Default
				}

				// Extract name
				if name, ok := fieldMap["name"].(string); ok {
					field.Name = name
				}

				// Extract label
				if label, ok := fieldMap["label"].(string); ok {
					field.Label = label
				}

				// Extract placeholder
				if placeholder, ok := fieldMap["placeholder"].(string); ok {
					field.Placeholder = placeholder
				}

				// Extract value
				if value, ok := fieldMap["value"].(string); ok {
					field.Value = value
				}

				// Extract required
				if required, ok := fieldMap["required"].(bool); ok {
					field.Required = required
				}

				// Extract validation
				if validation, ok := fieldMap["validation"].(string); ok {
					field.Validation = validation
				}

				// Extract options
				if options, ok := fieldMap["options"].([]interface{}); ok {
					field.Options = make([]string, len(options))
					for i, opt := range options {
						if optStr, ok := opt.(string); ok {
							field.Options[i] = optStr
						}
					}
				}

				// Extract width
				if width, ok := fieldMap["width"].(int); ok {
					field.Width = width
				} else if width, ok := fieldMap["width"].(float64); ok {
					field.Width = int(width)
				}

				fields = append(fields, field)
			}
		}
	}

	return fields
}

// applyTextareaProps applies props to Textarea components.
func (f *Factory) applyTextareaProps(comp *components.TextareaComponent, props map[string]interface{}) {
	// Dimensions
	if width, ok := props["width"].(int); ok {
		comp.WithWidth(width)
	}
	if width, ok := props["width"].(float64); ok {
		comp.WithWidth(int(width))
	}
	if height, ok := props["height"].(int); ok {
		comp.WithHeight(height)
	}
	if height, ok := props["height"].(float64); ok {
		comp.WithHeight(int(height))
	}
	if maxHeight, ok := props["maxHeight"].(int); ok {
		comp.WithMaxHeight(maxHeight)
	}
	if maxHeight, ok := props["maxHeight"].(float64); ok {
		comp.WithMaxHeight(int(maxHeight))
	}

	// Text properties
	if placeholder, ok := props["placeholder"].(string); ok {
		comp.WithPlaceholder(placeholder)
	}
	if value, ok := props["value"].(string); ok {
		comp.WithValue(value)
	}
	if prompt, ok := props["prompt"].(string); ok {
		comp.WithPrompt(prompt)
	}

	// Behavior
	if charLimit, ok := props["charLimit"].(int); ok {
		comp.WithCharLimit(charLimit)
	}
	if charLimit, ok := props["charLimit"].(float64); ok {
		comp.WithCharLimit(int(charLimit))
	}
	if showLineNumbers, ok := props["showLineNumbers"].(bool); ok {
		comp.WithShowLineNumbers(showLineNumbers)
	}
	if enterSubmits, ok := props["enterSubmits"].(bool); ok {
		comp.WithEnterSubmits(enterSubmits)
	}
}

// applyProgressProps applies props to Progress components.
func (f *Factory) applyProgressProps(comp *components.ProgressComponent, props map[string]interface{}) {
	// Dimensions
	if width, ok := props["width"].(int); ok {
		comp.WithWidth(width)
	}
	if width, ok := props["width"].(float64); ok {
		comp.WithWidth(int(width))
	}
	if height, ok := props["height"].(int); ok {
		comp.WithHeight(height)
	}
	if height, ok := props["height"].(float64); ok {
		comp.WithHeight(int(height))
	}

	// Progress properties
	if percent, ok := props["percent"].(float64); ok {
		comp.WithPercent(percent)
	}
	if percent, ok := props["percent"].(int); ok {
		comp.WithPercent(float64(percent))
	}
	if label, ok := props["label"].(string); ok {
		comp.WithLabel(label)
	}
	if showPercentage, ok := props["showPercentage"].(bool); ok {
		comp.WithShowPercentage(showPercentage)
	}

	// Characters
	if filledChar, ok := props["filledChar"].(string); ok {
		comp.WithFilledChar(filledChar)
	}
	if emptyChar, ok := props["emptyChar"].(string); ok {
		comp.WithEmptyChar(emptyChar)
	}

	// Colors
	if fullColor, ok := props["fullColor"].(string); ok {
		comp.WithFullColor(fullColor)
	} else if color, ok := props["color"].(string); ok {
		comp.WithFullColor(color)
	}
	if emptyColor, ok := props["emptyColor"].(string); ok {
		comp.WithEmptyColor(emptyColor)
	}

	// Behavior
	if animated, ok := props["animated"].(bool); ok {
		comp.WithAnimated(animated)
	}
}

// applySpinnerProps applies props to Spinner components.
func (f *Factory) applySpinnerProps(comp *components.SpinnerComponent, props map[string]interface{}) {
	// Dimensions
	if width, ok := props["width"].(int); ok {
		comp.WithWidth(width)
	}
	if width, ok := props["width"].(float64); ok {
		comp.WithWidth(int(width))
	}
	if height, ok := props["height"].(int); ok {
		comp.WithHeight(height)
	}
	if height, ok := props["height"].(float64); ok {
		comp.WithHeight(int(height))
	}

	// Spinner properties
	if style, ok := props["style"].(string); ok {
		comp.WithStyle(style)
	}
	if label, ok := props["label"].(string); ok {
		comp.WithLabel(label)
	}
	if labelPosition, ok := props["labelPosition"].(string); ok {
		comp.WithLabelPosition(labelPosition)
	}
	if running, ok := props["running"].(bool); ok {
		comp.WithRunning(running)
	}

	// Colors - convert color names to ANSI codes
	if color, ok := props["color"].(string); ok {
		comp.WithColor(ColorNameToANSI(color))
	}
	if background, ok := props["background"].(string); ok {
		comp.WithBackground(ColorNameToANSI(background))
	}

	// Speed
	if speed, ok := props["speed"].(int); ok {
		comp.WithSpeed(speed)
	} else if speed, ok := props["speed"].(float64); ok {
		comp.WithSpeed(int(speed))
	}

	// Custom frames
	if framesData, ok := props["frames"].([]interface{}); ok {
		frames := make([]string, 0, len(framesData))
		for _, f := range framesData {
			if frameStr, ok := f.(string); ok {
				frames = append(frames, frameStr)
			}
		}
		if len(frames) > 0 {
			comp.WithFrames(frames)
		}
	}
}

// applyModalProps applies props to Modal components.
func (f *Factory) applyModalProps(comp *components.ModalComponent, props map[string]interface{}) {
	if title, ok := props["title"].(string); ok {
		comp.WithTitle(title)
	}
	if width, ok := props["width"].(int); ok {
		comp.WithSize(width, 0)
	} else if width, ok := props["width"].(float64); ok {
		comp.WithSize(int(width), 0)
	}
	if height, ok := props["height"].(int); ok {
		comp.WithSize(0, height)
	} else if height, ok := props["height"].(float64); ok {
		comp.WithSize(0, int(height))
	}
	if centered, ok := props["centered"].(bool); ok {
		comp.WithCentered(centered)
	}
	if closeOnEsc, ok := props["closeOnEsc"].(bool); ok {
		comp.WithCloseOnEsc(closeOnEsc)
	}
	if closeOnBackdrop, ok := props["closeOnBackdrop"].(bool); ok {
		comp.WithCloseOnBackdrop(closeOnBackdrop)
	}
	if focusable, ok := props["focusable"].(bool); ok {
		comp.WithFocusable(focusable)
	}
	if zIndex, ok := props["zIndex"].(int); ok {
		comp.WithZIndex(zIndex)
	} else if zIndex, ok := props["zIndex"].(float64); ok {
		comp.WithZIndex(int(zIndex))
	}
	// Handle visibility
	if visible, ok := props["visible"].(bool); ok {
		if visible {
			comp.Show()
		} else {
			comp.Hide()
		}
	}
	// Handle onClose callback
	if onClose, ok := props["onClose"]; ok {
		if fn, ok := onClose.(func()); ok {
			comp.WithOnClose(fn)
		}
	}
}

// applyTabsProps applies props to Tabs components.
func (f *Factory) applyTabsProps(comp *components.TabsComponent, props map[string]interface{}) {
	if width, ok := props["width"].(int); ok {
		comp.WithSize(width, 0)
	} else if width, ok := props["width"].(float64); ok {
		comp.WithSize(int(width), 0)
	}
	if height, ok := props["height"].(int); ok {
		comp.WithSize(0, height)
	} else if height, ok := props["height"].(float64); ok {
		comp.WithSize(0, int(height))
	}
	if activeIndex, ok := props["activeIndex"].(int); ok {
		comp.WithActiveIndex(activeIndex)
	} else if activeIndex, ok := props["activeIndex"].(float64); ok {
		comp.WithActiveIndex(int(activeIndex))
	}
	if focusable, ok := props["focusable"].(bool); ok {
		comp.WithFocusable(focusable)
	}
	// Handle tab items from children
	if tabsData, ok := props["tabs"].([]interface{}); ok {
		tabs := make([]*components.TabItem, 0, len(tabsData))
		for _, tabData := range tabsData {
			if tabMap, ok := tabData.(map[string]interface{}); ok {
				tab := &components.TabItem{}
				if id, ok := tabMap["id"].(string); ok {
					tab.ID = id
				}
				if label, ok := tabMap["label"].(string); ok {
					tab.Label = label
				}
				if icon, ok := tabMap["icon"].(string); ok {
					tab.Icon = icon
				}
				if closable, ok := tabMap["closable"].(bool); ok {
					tab.Closable = closable
				}
				tabs = append(tabs, tab)
			}
		}
		if len(tabs) > 0 {
			comp.WithTabs(tabs)
		}
	}
	// Handle onTabChange callback
	if onTabChange, ok := props["onTabChange"]; ok {
		if fn, ok := onTabChange.(func(int)); ok {
			comp.WithOnTabChange(fn)
		}
	}
}

// applyTreeProps applies props to Tree components.
func (f *Factory) applyTreeProps(comp *components.TreeComponent, props map[string]interface{}) {
	if width, ok := props["width"].(int); ok {
		comp.WithSize(width, 0)
	} else if width, ok := props["width"].(float64); ok {
		comp.WithSize(int(width), 0)
	}
	if height, ok := props["height"].(int); ok {
		comp.WithSize(0, height)
	} else if height, ok := props["height"].(float64); ok {
		comp.WithSize(0, int(height))
	}
	if multiSelect, ok := props["multiSelect"].(bool); ok {
		comp.WithMultiSelect(multiSelect)
	}
	// Handle root node from children or props
	if rootNodeData, ok := props["root"].(map[string]interface{}); ok {
		rootNode := f.parseTreeNode(rootNodeData, nil)
		comp.WithRoot(rootNode)
	}
}

// parseTreeNode parses a tree node from props data.
func (f *Factory) parseTreeNode(data map[string]interface{}, parent *components.TreeNode) *components.TreeNode {
	if data == nil {
		return nil
	}

	node := &components.TreeNode{}

	if id, ok := data["id"].(string); ok {
		node.ID = id
	}
	if label, ok := data["label"].(string); ok {
		node.Label = label
	}
	if icon, ok := data["icon"].(string); ok {
		node.Icon = icon
	}
	if expanded, ok := data["expanded"].(bool); ok {
		node.Expanded = expanded
	}
	if dataNode, ok := data["data"]; ok {
		node.Data = dataNode
	}

	node.Parent = parent

	// Parse children
	if childrenData, ok := data["children"].([]interface{}); ok {
		node.Children = make([]*components.TreeNode, 0, len(childrenData))
		for _, childData := range childrenData {
			if childMap, ok := childData.(map[string]interface{}); ok {
				child := f.parseTreeNode(childMap, node)
				node.Children = append(node.Children, child)
			}
		}
	}

	return node
}

// applySplitPaneProps applies props to SplitPane components.
func (f *Factory) applySplitPaneProps(comp *components.SplitPaneComponent, props map[string]interface{}) {
	// Direction
	if direction, ok := props["direction"].(string); ok {
		switch direction {
		case "horizontal", "row", "h":
			comp.WithDirection(components.SplitHorizontal)
		case "vertical", "column", "v":
			comp.WithDirection(components.SplitVertical)
		}
	}

	// Split ratio
	if splitRatio, ok := props["splitRatio"].(float64); ok {
		comp.WithSplitRatio(splitRatio)
	} else if splitRatio, ok := props["splitRatio"].(int); ok {
		comp.WithSplitRatio(float64(splitRatio))
	}

	// Minimum split
	if minSplit, ok := props["minSplit"].(float64); ok {
		comp.WithMinSplit(minSplit)
	} else if minSplit, ok := props["minSplit"].(int); ok {
		comp.WithMinSplit(float64(minSplit))
	}

	// Handle size
	if handleSize, ok := props["handleSize"].(int); ok {
		comp.WithHandleSize(handleSize)
	} else if handleSize, ok := props["handleSize"].(float64); ok {
		comp.WithHandleSize(int(handleSize))
	}

	// Focusable
	if focusable, ok := props["focusable"].(bool); ok {
		comp.WithFocusable(focusable)
	}

	// Dimensions
	if width, ok := props["width"].(int); ok {
		comp.WithSize(width, 0)
	} else if width, ok := props["width"].(float64); ok {
		comp.WithSize(int(width), 0)
	}
	if height, ok := props["height"].(int); ok {
		comp.WithSize(0, height)
	} else if height, ok := props["height"].(float64); ok {
		comp.WithSize(0, int(height))
	}

	// Handle first pane component
	if firstConfig, ok := props["first"].(map[string]interface{}); ok {
		if componentType, ok := firstConfig["type"].(string); ok {
			config := &ComponentConfig{
				ID:     getStringOrDefault(firstConfig, "id", "first-pane"),
				Type:   componentType,
				Props:  getPropsMap(firstConfig),
				Width:  firstConfig["width"],
				Height: firstConfig["height"],
			}
			if firstComp, err := f.Create(config); err == nil {
				comp.WithFirst(firstComp)
			}
		}
	}

	// Handle second pane component
	if secondConfig, ok := props["second"].(map[string]interface{}); ok {
		if componentType, ok := secondConfig["type"].(string); ok {
			config := &ComponentConfig{
				ID:     getStringOrDefault(secondConfig, "id", "second-pane"),
				Type:   componentType,
				Props:  getPropsMap(secondConfig),
				Width:  secondConfig["width"],
				Height: secondConfig["height"],
			}
			if secondComp, err := f.Create(config); err == nil {
				comp.WithSecond(secondComp)
			}
		}
	}

	// Handle onResize callback
	if onResize, ok := props["onResize"]; ok {
		if fn, ok := onResize.(func(float64)); ok {
			comp.WithOnResize(fn)
		}
	}
}

// getStringOrDefault gets a string value from map or returns default.
func getStringOrDefault(m map[string]interface{}, key string, defaultVal string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return defaultVal
}

// getPropsMap extracts props from a component config map.
func getPropsMap(m map[string]interface{}) map[string]interface{} {
	props := make(map[string]interface{})
	for k, v := range m {
		// Skip structural keys
		if k != "type" && k != "id" && k != "width" && k != "height" {
			props[k] = v
		}
	}
	return props
}

// applyContextMenuProps applies props to ContextMenu components.
func (f *Factory) applyContextMenuProps(comp *components.ContextMenuComponent, props map[string]interface{}) {
	// Position
	if x, ok := props["x"].(int); ok {
		if y, ok2 := props["y"].(int); ok2 {
			comp.WithPosition(x, y)
		}
	} else if x, ok := props["x"].(float64); ok {
		if y, ok2 := props["y"].(float64); ok2 {
			comp.WithPosition(int(x), int(y))
		}
	}

	// Dimensions
	if width, ok := props["width"].(int); ok {
		comp.WithWidth(width)
	} else if width, ok := props["width"].(float64); ok {
		comp.WithWidth(int(width))
	}
	if maxWidth, ok := props["maxWidth"].(int); ok {
		comp.WithMaxWidth(maxWidth)
	} else if maxWidth, ok := props["maxWidth"].(float64); ok {
		comp.WithMaxWidth(int(maxWidth))
	}
	if maxHeight, ok := props["maxHeight"].(int); ok {
		comp.WithMaxHeight(maxHeight)
	} else if maxHeight, ok := props["maxHeight"].(float64); ok {
		comp.WithMaxHeight(int(maxHeight))
	}

	// Z-index
	if zIndex, ok := props["zIndex"].(int); ok {
		comp.WithZIndex(zIndex)
	} else if zIndex, ok := props["zIndex"].(float64); ok {
		comp.WithZIndex(int(zIndex))
	}

	// Focusable
	if focusable, ok := props["focusable"].(bool); ok {
		comp.WithFocusable(focusable)
	}

	// Visible
	if visible, ok := props["visible"].(bool); ok {
		if visible {
			comp.Show()
		} else {
			comp.Hide()
		}
	}

	// Handle menu items from children or props
	if itemsData, ok := props["items"].([]interface{}); ok {
		items := f.parseContextMenuItems(itemsData)
		if len(items) > 0 {
			comp.WithItems(items)
		}
	}

	// Handle callbacks
	if onDismiss, ok := props["onDismiss"]; ok {
		if fn, ok := onDismiss.(func()); ok {
			comp.WithOnDismiss(fn)
		}
	}
	if onShow, ok := props["onShow"]; ok {
		if fn, ok := onShow.(func()); ok {
			comp.WithOnShow(fn)
		}
	}
}

// parseContextMenuItems parses context menu items from DSL data.
func (f *Factory) parseContextMenuItems(itemsData []interface{}) []*components.ContextMenuItem {
	items := make([]*components.ContextMenuItem, 0, len(itemsData))

	for _, itemData := range itemsData {
		if itemMap, ok := itemData.(map[string]interface{}); ok {
			item := &components.ContextMenuItem{}

			// Check if this is a divider
			if divider, ok := itemMap["divider"].(bool); ok {
				if divider {
					item.Divider = true
					items = append(items, item)
					continue
				}
			}

			// Extract id
			if id, ok := itemMap["id"].(string); ok {
				item.ID = id
			}

			// Extract label
			if label, ok := itemMap["label"].(string); ok {
				item.Label = label
			}

			// Extract icon
			if icon, ok := itemMap["icon"].(string); ok {
				item.Icon = icon
			}

			// Extract shortcut
			if shortcut, ok := itemMap["shortcut"].(string); ok {
				item.Shortcut = shortcut
			}

			// Extract disabled
			if disabled, ok := itemMap["disabled"].(bool); ok {
				item.Disabled = disabled
			}

			// Handle action callback
			if action, ok := itemMap["action"]; ok {
				if fn, ok := action.(func()); ok {
					item.Action = fn
				}
			}

			// Handle submenu (recursive)
			if submenuData, ok := itemMap["submenu"].(map[string]interface{}); ok {
				if submenuItemsData, ok := submenuData["items"].([]interface{}); ok {
					submenuItems := f.parseContextMenuItems(submenuItemsData)
					submenu := components.NewContextMenu()
					submenu.WithItems(submenuItems)

					// Copy position and visibility settings
					if x, ok := submenuData["x"].(int); ok {
						if y, ok2 := submenuData["y"].(int); ok2 {
							submenu.WithPosition(x, y)
						}
					}
					if visible, ok := submenuData["visible"].(bool); ok {
						if visible {
							submenu.Show()
						}
					}

					item.Submenu = submenu
				}
			}

			items = append(items, item)
		} else if itemStr, ok := itemData.(string); ok {
			// Simple string item (for quick menus)
			item := &components.ContextMenuItem{
				ID:    itemStr,
				Label: itemStr,
			}
			items = append(items, item)
		}
	}

	return items
}

// ConvertObjectArrayToTableData converts an array of objects to a 2D data array
// using the column keys to map object fields to array positions
func (f *Factory) ConvertObjectArrayToTableData(objArray []interface{}, table *components.TableComponent) [][]interface{} {
	// Get columns from the table component
	columns := table.GetColumns()
	if len(columns) == 0 {
		// No columns set yet, can't convert
		return nil
	}

	result := make([][]interface{}, 0, len(objArray))
	for _, objIntf := range objArray {
		if objMap, ok := objIntf.(map[string]interface{}); ok {
			row := make([]interface{}, len(columns))
			for i, col := range columns {
				if val, ok := objMap[col.Key]; ok {
					row[i] = val
				} else {
					row[i] = ""
				}
			}
			result = append(result, row)
		}
	}
	return result
}


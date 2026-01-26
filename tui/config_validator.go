package tui

import (
	"fmt"
	"strings"

	"github.com/yaoapp/kun/log"
)

// ConfigValidator validates TUI configuration files.
// It works closely with the render engine to ensure consistency.
type ConfigValidator struct {
	config   *Config
	registry *ComponentRegistry // DEPRECATED: Will be nil
	errors   []ValidationError
	warnings []ValidationError
}

// ValidationError represents a validation error with context.
type ValidationError struct {
	Path    string // JSON path to the invalid field
	Message string // Error description
	Level   string // "error" or "warning"
	Index   int    // For array indices
}

// NewConfigValidator creates a new validator for the given config.
func NewConfigValidator(cfg *Config, registry *ComponentRegistry) *ConfigValidator {
	return &ConfigValidator{
		config:   cfg,
		registry: registry,
		errors:   []ValidationError{},
		warnings: []ValidationError{},
	}
}

// Validate performs all validations and logs detailed errors.
// Returns true if config is valid, false otherwise.
func (v *ConfigValidator) Validate() bool {
	log.Trace("ConfigValidator: Starting validation for TUI '%s'", v.config.Name)

	v.validateBasicStructure()
	v.validateLayoutStructure(&v.config.Layout, "layout")
	v.validateComponentsInLayout(&v.config.Layout, "layout")
	v.validateDataBindings()

	// Log all validation results
	v.logResults()

	return len(v.errors) == 0
}

// validateBasicStructure validates top-level config structure.
func (v *ConfigValidator) validateBasicStructure() {
	// Validate name
	if v.config.Name == "" {
		v.addError("name", "TUI name is required")
	} else if len(v.config.Name) > 100 {
		v.addWarning("name", "TUI name may be too long (> 100 characters)")
	}

	// Validate log level
	if v.config.LogLevel != "" {
		validLevels := map[string]bool{
			"trace": true,
			"debug": true,
			"info":  true,
			"warn":  true,
			"error": true,
			"none":  true,
		}
		if !validLevels[v.config.LogLevel] {
			v.addError("logLevel",
				fmt.Sprintf("invalid log level: '%s' (must be one of: trace, debug, info, warn, error, none)",
					v.config.LogLevel))
		}
	}

	// Note: autoFocus has a default value of true, so not setting it is fine
	// No warning needed as it's a reasonable default
}

// validateLayoutStructure validates layout structure recursively.
func (v *ConfigValidator) validateLayoutStructure(layout *Layout, path string) {
	if layout == nil {
		v.addError(path, "layout is nil")
		return
	}

	// Validate direction
	if layout.Direction == "" {
		layout.Direction = "vertical" // Set default
	}
	validDirections := map[string]bool{
		"vertical":   true,
		"horizontal": true,
		"column":     true,
		"row":        true,
	}
	if !validDirections[layout.Direction] {
		v.addError(path+".direction",
			fmt.Sprintf("invalid direction: '%s' (must be one of: vertical, horizontal, column, row)",
				layout.Direction))
	}

	// Validate padding
	if len(layout.Padding) > 0 && len(layout.Padding) < 4 {
		// Allow partial padding (will be normalized)
		v.addWarning(path+".padding",
			fmt.Sprintf("partial padding specified (%d values), will be normalized to 4", len(layout.Padding)))
	}

	// Validate children count
	if len(layout.Children) == 0 {
		v.addWarning(path+".children", "layout has no children")
	} else if len(layout.Children) > 100 {
		v.addWarning(path+".children", "layout has many children (> 100), may impact performance")
	}

	// Recursively validate nested layouts
	for i, child := range layout.Children {
		childPath := fmt.Sprintf("%s.children[%d]", path, i)
		v.validateChildStructure(&child, childPath, 0)
	}
}

// validateChildStructure validates a single child structure (nested).
func (v *ConfigValidator) validateChildStructure(child *Component, path string, depth int) {
	const maxNestingDepth = 50

	if depth > maxNestingDepth {
		v.addError(path, fmt.Sprintf("layout nesting depth exceeds maximum: %d", depth))
		return
	}

	// Validate nested layout - supports two formats:
	// Format 1 (old): child.type == "layout" with child.children and child.direction
	// Format 2 (new): child.type == "layout", nested layout in child.props.layout
	if child.Type == "layout" {
		// Check for nested layout in props (new format)
		if nestedLayout, ok := child.Props["layout"].(*Layout); ok {
			layoutPath := fmt.Sprintf("%s.props.layout", path)
			v.validateLayoutStructure(nestedLayout, layoutPath)
			v.validateComponentsInLayout(nestedLayout, layoutPath)
		}

		// Also validate if child has its own children (old format)
		// In old format, a type="layout" component can have direction and children
		if len(child.Children) > 0 {
			layoutPath := fmt.Sprintf("%s.children", path)
			// Validate as a regular layout
			v.validateLayoutStructure(&Layout{
				Direction: child.Direction,
				Children:  child.Children,
				Padding:   []int{},
			}, layoutPath)
			// And validate components in this nested layout
			v.validateComponentsInLayout(&Layout{
				Direction: child.Direction,
				Children:  child.Children,
				Padding:   []int{},
			}, layoutPath)
		}
	}
}

// validateComponentsInLayout validates components within the layout.
// This ensures components are compatible with the render engine.
func (v *ConfigValidator) validateComponentsInLayout(layout *Layout, path string) {
	if layout == nil {
		return
	}

	for i, child := range layout.Children {
		childPath := fmt.Sprintf("%s.children[%d]", path, i)
		v.validateComponent(&child, childPath, 0)
	}
}

// validateComponent validates a component with its configuration.
// Checks that component types exist and are compatible.
func (v *ConfigValidator) validateComponent(comp *Component, path string, depth int) {
	const maxNestingDepth = 50

	if depth > maxNestingDepth {
		return
	}

	// Skip nested layouts (validated elsewhere)
	if comp.Type == "layout" {
		if nestedLayout, ok := comp.Props["layout"].(*Layout); ok {
			layoutPath := fmt.Sprintf("%s.props.layout", path)
			v.validateLayoutStructure(nestedLayout, layoutPath)
			v.validateComponentsInLayout(nestedLayout, layoutPath)
		}
		return
	}

	// Validate type is present
	if comp.Type == "" {
		v.addError(path+".type", "component type is required")
		return
	}

	// Check if component type is registered
	if v.registry != nil {
		// Registry is deprecated, skip factory check
		_ = v.registry
	} else {
		// Check if it's a known built-in type
		builtIns := map[string]bool{
			"header": true, "footer": true, "text": true,
			"input": true, "textarea": true, "menu": true,
			"table": true, "list": true, "form": true,
			"chat": true, "crud": true, "viewport": true,
			"filepicker": true, "paginator": true,
			"progress": true, "timer": true, "stopwatch": true,
			"spinner": true, "cursor": true, "help": true,
			"row": true, "column": true, "flex": true,
			"box": true, "button": true, "calendar": true,
			"checkbox": true, "contextmenu": true,
			"modal": true, "radio": true, "splitpane": true,
			"tabs": true, "tree": true, "virtual_list": true,
		}

		if !builtIns[comp.Type] && comp.Type != "layout" {
			v.addError(path+".type",
				fmt.Sprintf("unknown component type: '%s'", comp.Type))
		}
	}

	// Validate ID
	if comp.ID == "" {
		v.addWarning(path+".id", "component has no ID - may cause issues with state binding and focus")
	}

	// Validate bindings
	if comp.Bind != "" {
		// Check if bound state key exists in data
		if v.config.Data != nil {
			if _, exists := v.getDataValue(comp.Bind); !exists {
				v.addWarning(path+".bind",
					fmt.Sprintf("bind references non-existent state key: '%s'", comp.Bind))
			}
		}
	}

	// Validate props for specific component types
	v.validateComponentProps(comp, path)
}

// validateComponentProps validates component-specific props.
// This ensures props are compatible with render expectations.
func (v *ConfigValidator) validateComponentProps(comp *Component, path string) {
	if comp.Props == nil {
		return
	}

	// Common validation for numeric sizes
	for key, value := range comp.Props {
		if strings.Contains(strings.ToLower(key), "width") ||
			strings.Contains(strings.ToLower(key), "height") ||
			strings.Contains(strings.ToLower(key), "size") {
			// Check if numeric values are reasonable
			if num, ok := value.(float64); ok && num < 0 {
				v.addWarning(path+".props."+key, "size value is negative, may cause rendering issues")
			}
		}
	}

	// Component-specific validation
	switch comp.Type {
	case "table":
		v.validateTableProps(comp, path)
	case "list":
		v.validateListProps(comp, path)
	case "input", "textarea":
		v.validateInputProps(comp, path)
	case "menu":
		v.validateMenuProps(comp, path)
	}
}

// validateTableProps validates table component props.
func (v *ConfigValidator) validateTableProps(comp *Component, path string) {
	props := comp.Props

	// Check if columns are specified
	if columns, ok := props["columns"]; ok {
		switch cols := columns.(type) {
		case []interface{}:
			if len(cols) == 0 {
				v.addError(path+".props.columns", "table columns array is empty")
			}
		case []map[string]interface{}:
			if len(cols) == 0 {
				v.addError(path+".props.columns", "table columns array is empty")
			}
			// Validate each column has required fields
			for i, col := range cols {
				colPath := fmt.Sprintf("%s.props.columns[%d]", path, i)
				if _, hasKey := col["key"]; !hasKey {
					v.addWarning(colPath+".key", "column missing 'key' field")
				}
			}
		default:
			v.addError(path+".props.columns", "columns must be an array")
		}
	} else {
		v.addWarning(path+".props.columns", "table has no columns defined")
	}

	// Check height
	if height, ok := props["height"]; ok {
		if num, ok := height.(float64); ok && num < 1 {
			v.addWarning(path+".props.height", "table height < 1 may not display any rows")
		}
	}
}

// validateListProps validates list component props.
func (v *ConfigValidator) validateListProps(comp *Component, path string) {
	props := comp.Props

	// Validate items count
	if items, ok := props["items"]; ok {
		switch items := items.(type) {
		case []interface{}:
			if len(items) == 0 {
				v.addWarning(path+".props.items", "list items array is empty")
			}
		}
	}
}

// validateInputProps validates input/textarea component props.
func (v *ConfigValidator) validateInputProps(comp *Component, path string) {
	props := comp.Props

	// Validate character limit
	if charLimit, ok := props["charLimit"]; ok {
		if num, ok := charLimit.(float64); ok && num < 1 {
			v.addWarning(path+".props.charLimit", "character limit < 1 may be invalid")
		}
	}
}

// validateMenuProps validates menu component props.
func (v *ConfigValidator) validateMenuProps(comp *Component, path string) {
	props := comp.Props

	// Validate items
	if items, ok := props["items"]; ok {
		switch items := items.(type) {
		case []interface{}:
			if len(items) == 0 {
				v.addWarning(path+".props.items", "menu items array is empty")
			}
		case []map[string]interface{}:
			if len(items) == 0 {
				v.addWarning(path+".props.items", "menu items array is empty")
			}
		}
	}
}

// validateDataBindings validates all data bindings in the config.
func (v *ConfigValidator) validateDataBindings() {
	if v.config.Data == nil {
		v.addWarning("data", "no initial data defined (may cause binding errors)")
		return
	}

	// Validate data is not empty (except for simple cases)
	if len(v.config.Data) == 0 {
		v.addWarning("data", "initial data is empty")
	}
}

// getDataValue retrieves a value from data by dot notation path.
func (v *ConfigValidator) getDataValue(path string) (interface{}, bool) {
	current := interface{}(v.config.Data)
	if current == nil {
		return nil, false
	}

	keys := strings.Split(path, ".")

	for _, key := range keys {
		switch curr := current.(type) {
		case map[string]interface{}:
			val, exists := curr[key]
			if !exists {
				return nil, false
			}
			current = val
		case map[interface{}]interface{}:
			val, exists := curr[key]
			if !exists {
				return nil, false
			}
			current = val
		default:
			return nil, false
		}
	}

	return current, true
}

// addError adds a validation error.
func (v *ConfigValidator) addError(path, message string) {
	v.errors = append(v.errors, ValidationError{
		Path:    path,
		Message: message,
		Level:   "error",
	})
}

// addWarning adds a validation warning.
func (v *ConfigValidator) addWarning(path, message string) {
	v.warnings = append(v.warnings, ValidationError{
		Path:    path,
		Message: message,
		Level:   "warning",
	})
}

// logResults logs all validation results.
func (v *ConfigValidator) logResults() {
	// Log warnings first
	for _, warning := range v.warnings {
		log.Warn("[Config Validator] Warning at %s: %s", warning.Path, warning.Message)
	}

	// Log errors
	for i, err := range v.errors {
		log.Error("[Config Validator] Error #%d at %s: %s", i+1, err.Path, err.Message)
	}

	// Summary
	if len(v.errors) == 0 && len(v.warnings) == 0 {
		log.Trace("[Config Validator] TUI '%s' is valid with no issues", v.config.Name)
	} else if len(v.errors) == 0 {
		log.Info("[Config Validator] TUI '%s' is valid with %d warnings",
			v.config.Name, len(v.warnings))
	} else {
		log.Error("[Config Validator] TUI '%s' has %d errors and %d warnings",
			v.config.Name, len(v.errors), len(v.warnings))
	}
}

// GetErrors returns all validation errors.
func (v *ConfigValidator) GetErrors() []ValidationError {
	return v.errors
}

// GetWarnings returns all validation warnings.
func (v *ConfigValidator) GetWarnings() []ValidationError {
	return v.warnings
}

// GetErrorSummary returns a formatted summary of validation errors.
func (v *ConfigValidator) GetErrorSummary() string {
	if len(v.errors) == 0 && len(v.warnings) == 0 {
		return "Configuration is valid"
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("Validation Summary for TUI '%s':\n", v.config.Name))

	if len(v.errors) > 0 {
		builder.WriteString(fmt.Sprintf("  %d Error(s):\n", len(v.errors)))
		for _, err := range v.errors {
			builder.WriteString(fmt.Sprintf("    - %s: %s\n", err.Path, err.Message))
		}
	}

	if len(v.warnings) > 0 {
		builder.WriteString(fmt.Sprintf("  %d Warning(s):\n", len(v.warnings)))
		for _, warn := range v.warnings {
			builder.WriteString(fmt.Sprintf("    - %s: %s\n", warn.Path, warn.Message))
		}
	}

	return builder.String()
}

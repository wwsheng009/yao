package tui

// DEPRECATED: This file previously contained component registry functionality.
// The component registry has been removed as part of migrating to tui/ui/components.
// ComponentType constants are kept for backward compatibility.

// ComponentType represents the type of a component
type ComponentType string

const (
	// Built-in component types (kept for backward compatibility)
	HeaderComponent     ComponentType = "header"
	TextComponent       ComponentType = "text"
	TableComponent      ComponentType = "table"
	FormComponent       ComponentType = "form"
	InputComponent      ComponentType = "input"
	ViewportComponent   ComponentType = "viewport"
	FooterComponent     ComponentType = "footer"
	ChatComponent       ComponentType = "chat"
	MenuComponent       ComponentType = "menu"
	TimerComponent      ComponentType = "timer"
	StopwatchComponent  ComponentType = "stopwatch"
	FilePickerComponent ComponentType = "filepicker"
	HelpComponent       ComponentType = "help"
	KeyComponent        ComponentType = "key"
	CursorComponent     ComponentType = "cursor"
	ListComponent       ComponentType = "list"
	PaginatorComponent  ComponentType = "paginator"
	ProgressComponent   ComponentType = "progress"
	SpinnerComponent    ComponentType = "spinner"
	TextareaComponent   ComponentType = "textarea"
	CRUDComponent       ComponentType = "crud"
)

// GetGlobalRegistry returns nil (registry has been removed)
func GetGlobalRegistry() *ComponentRegistry {
	return nil
}

// ComponentRegistry is a placeholder for backward compatibility
type ComponentRegistry struct{}

// IsFocusable returns false (registry has been removed)
func (r *ComponentRegistry) IsFocusable(componentType ComponentType) bool {
	return false
}

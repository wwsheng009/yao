package components

import (
	"github.com/yaoapp/yao/tui/core"
)

// Note: NewHeaderComponent and NewTextComponent are already defined in their respective files
// These factory functions only exist for components that don't have their own
// New*Component(id) functions

// NewFooterComponent creates a new Footer component wrapper
func NewFooterComponent(config core.RenderConfig, id string) *FooterComponentWrapper {
	var props FooterProps
	
	// Extract props from config
	if config.Data != nil {
		if dataMap, ok := config.Data.(map[string]interface{}); ok {
			props = ParseFooterProps(dataMap)
		}
	}
	
	// Use defaults if no data provided
	if props.Text == "" {
		props = FooterProps{
			Text: "",
			Height: 0,
			Width: 0,
			Color: "",
		}
	}
	
	model := NewFooterModel(props, id)
	return &FooterComponentWrapper{
		model: &model,
	}
}

// NewInputComponent creates a new Input component wrapper
func NewInputComponent(config core.RenderConfig, id string) *InputComponentWrapper {
	var props InputProps
	
	// Extract props from config
	if config.Data != nil {
		if dataMap, ok := config.Data.(map[string]interface{}); ok {
			props = ParseInputProps(dataMap)
		}
	}
	
	// Use defaults if no data provided
	if props.Placeholder == "" && props.Value == "" {
		props = InputProps{
			Placeholder: "",
			Value:       "",
			Prompt:      "> ",
		}
	}
	
	// Directly call the unified wrapper constructor
	return NewInputComponentWrapper(props, id)
}

// NewTextareaComponent creates a new Textarea component wrapper
func NewTextareaComponent(config core.RenderConfig, id string) *TextareaComponentWrapper {
	var props TextareaProps
	
	// Extract props from config
	if config.Data != nil {
		if dataMap, ok := config.Data.(map[string]interface{}); ok {
			props = ParseTextareaProps(dataMap)
		}
	}
	
	// Use defaults if no data provided
	if props.Placeholder == "" && props.Value == "" {
		props = TextareaProps{
			Placeholder: "",
			Value:       "",
			Prompt:      "> ",
		}
	}
	
	// 直接调用统一的 wrapper 构造函数
	return NewTextareaComponentWrapper(props, id)
}

// NewMenuComponent creates a new Menu component wrapper
func NewMenuComponent(config core.RenderConfig, id string) *MenuComponentWrapper {
	var props MenuProps

	// Extract props from config
	if config.Data != nil {
		if dataMap, ok := config.Data.(map[string]interface{}); ok {
			props = ParseMenuProps(dataMap)
		}
	}

	// Use defaults if no data provided
	if len(props.Items) == 0 {
		props = MenuProps{
			Items:        []MenuItem{},
			Title:        "",
			Height:       0,
			Width:        0,
			Focused:       false,
			ShowStatusBar: true,
			ShowFilter:    false,
		}
	}

	return NewMenuComponentWrapper(props, id)
}

// NewTableComponent creates a new Table component wrapper
func NewTableComponent(config core.RenderConfig, id string) *TableComponentWrapper {
	var props TableProps
	
	// Extract props from config
	if config.Data != nil {
		if dataMap, ok := config.Data.(map[string]interface{}); ok {
			props = ParseTableProps(dataMap)
		}
	}
	
	// Use defaults if no data provided
	if len(props.Columns) == 0 {
		props = TableProps{
			Columns: []Column{},
			Data:    [][]interface{}{},
			Focused:  false,
			Height:   0,
			Width:    0,
		}
	}
	
	// 直接调用统一的 wrapper 构造函数
	return NewTableComponentWrapper(props, id)
}

// NewFormComponent creates a new Form component wrapper
func NewFormComponent(config core.RenderConfig, id string) *FormComponentWrapper {
	var props FormProps
	
	// Extract props from config
	if config.Data != nil {
		if dataMap, ok := config.Data.(map[string]interface{}); ok {
			props = ParseFormProps(dataMap)
		}
	}
	
	// Use defaults if no data provided
	if len(props.Fields) == 0 {
		props = FormProps{
			Fields:       []Field{},
			Title:        "",
			Description:   "",
			SubmitLabel:   "Submit",
			CancelLabel:   "Cancel",
		}
	}
	
	// 直接调用统一的 wrapper 构造函数
	return NewFormComponentWrapper(props, id)
}

// NewListComponent creates a new List component wrapper
func NewListComponent(config core.RenderConfig, id string) *ListComponentWrapper {
	var props ListProps
	
	// Extract props from config
	if config.Data != nil {
		if dataMap, ok := config.Data.(map[string]interface{}); ok {
			props = ParseListProps(dataMap)
		}
	}
	
	// Use defaults if no data provided
	if len(props.Items) == 0 {
		props = ListProps{
			Items:   []ListItem{},
			Title:    "",
			Height:   0,
			Width:    0,
			Focused:  false,
		}
	}
	
	model := NewListModel(props, id)
	return &ListComponentWrapper{
		model: &model,
	}
}

// NewPaginatorComponent creates a new Paginator component wrapper
func NewPaginatorComponent(config core.RenderConfig, id string) *PaginatorComponentWrapper {
	var props PaginatorProps
	
	// Extract props from config
	if config.Data != nil {
		if dataMap, ok := config.Data.(map[string]interface{}); ok {
			props = ParsePaginatorProps(dataMap)
		}
	}
	
	// Use defaults if no data provided
	if props.TotalPages == 0 {
		props = PaginatorProps{
			TotalPages:  1,
			CurrentPage: 1,
			PageSize:     10,
			TotalItems:   0,
			Type:        "dots",
		}
	}
	
	model := NewPaginatorModel(props, id)
	return &PaginatorComponentWrapper{
		model: &model,
	}
}

// NewViewportComponent creates a new Viewport component wrapper
func NewViewportComponent(config core.RenderConfig, id string) *ViewportComponentWrapper {
	var props ViewportProps
	
	// Extract props from config
	if config.Data != nil {
		if dataMap, ok := config.Data.(map[string]interface{}); ok {
			props = ParseViewportProps(dataMap)
		}
	}
	
	// Use defaults if no data provided
	if props.Content == "" && props.Width == 0 {
		props = ViewportProps{
			Content:       "",
			Width:         0,
			Height:        0,
			ShowScrollbar: true,
		}
	}
	
	model := NewViewportModel(props, id)
	return &ViewportComponentWrapper{
		model: &model,
	}
}

// NewChatComponent creates a new Chat component wrapper
func NewChatComponent(config core.RenderConfig, id string) *ChatComponentWrapper {
	var props ChatProps

	// Extract props from config
	if config.Data != nil {
		if dataMap, ok := config.Data.(map[string]interface{}); ok {
			props = ParseChatProps(dataMap)
		}
	}

	// Use defaults if no data provided
	if len(props.Messages) == 0 {
		props = ChatProps{
			Messages:        []Message{},
			InputPlaceholder: "Type your message...",
			ShowInput:       true,
			EnableMarkdown:   true,
			GlamourStyle:    "dark",
			Width:           0,
			Height:          0,
			InputHeight:     3,
		}
	}

	model := NewChatModel(props, id)
	return NewChatComponentWrapper(&model)
}

// NewProgressComponent creates a new Progress component wrapper
func NewProgressComponent(config core.RenderConfig, id string) *ProgressComponentWrapper {
	var props ProgressProps
	
	// Extract props from config
	if config.Data != nil {
		if dataMap, ok := config.Data.(map[string]interface{}); ok {
			props = ParseProgressProps(dataMap)
		}
	}
	
	// Use defaults if no data provided
	// Note: Don't override if props were provided, even if they match default values
	if props.Width == 0 && props.Percent == 0 && props.Color == "" {
		props = ProgressProps{
			Percent:     0,
			Width:       40, // Default width to ensure visibility
			Color:       "",
			Background:  "",
			EmptyColor:  "",
		}
	}
	
	// Ensure width is set (use default if not provided)
	if props.Width == 0 {
		props.Width = 40 // Default width
	}
	
	model := NewProgressModel(props, id)
	return &ProgressComponentWrapper{
		model: &model,
	}
}

// NewSpinnerComponent creates a new Spinner component wrapper
func NewSpinnerComponent(config core.RenderConfig, id string) *SpinnerComponentWrapper {
	var props SpinnerProps
	
	// Extract props from config
	if config.Data != nil {
		if dataMap, ok := config.Data.(map[string]interface{}); ok {
			props = ParseSpinnerProps(dataMap)
		}
	}
	
	// Use defaults if no data provided
	if props.Style == "" {
		props = SpinnerProps{
			Style: "Dot",
			Color: "",
			Background:  "",
			Speed:     100,
			Frames:    nil,
		}
	}
	
	model := NewSpinnerModel(props, id)
	return &SpinnerComponentWrapper{
		model: &model,
	}
}

// NewTimerComponent creates a new Timer component wrapper
func NewTimerComponent(config core.RenderConfig, id string) *TimerComponentWrapper {
	var props TimerProps
	
	// Extract props from config
	if config.Data != nil {
		if dataMap, ok := config.Data.(map[string]interface{}); ok {
			props = ParseTimerProps(dataMap)
		}
	}
	
	// Use defaults if no data provided
	if props.Duration == 0 && props.Format == "" {
		props = TimerProps{
			Duration:   0,
			Format:     "mm:ss",
			Color:      "",
			Background: "",
			Bold:       false,
		}
	}
	
	model := NewTimerModel(props, id)
	return &TimerComponentWrapper{
		model: &model,
	}
}

// NewStopwatchComponent creates a new Stopwatch component wrapper
func NewStopwatchComponent(config core.RenderConfig, id string) *StopwatchComponentWrapper {
	var props StopwatchProps
	
	// Extract props from config
	if config.Data != nil {
		if dataMap, ok := config.Data.(map[string]interface{}); ok {
			props = ParseStopwatchProps(dataMap)
		}
	}
	
	// Use defaults if no data provided
	if props.Format == "" {
		props = StopwatchProps{
			Format:     "mm:ss",
			Color:      "",
			Background: "",
			Bold:       false,
			Italic:     false,
		}
	}
	
	model := NewStopwatchModel(props, id)
	return &StopwatchComponentWrapper{
		model: &model,
	}
}

// NewCursorComponent creates a new Cursor component wrapper
func NewCursorComponent(config core.RenderConfig, id string) *CursorComponentWrapper {
	var props CursorProps
	
	// Extract props from config
	if config.Data != nil {
		if dataMap, ok := config.Data.(map[string]interface{}); ok {
			props = ParseCursorProps(dataMap)
		}
	}
	
	// Use defaults if no data provided
	if props.Color == "" && props.Style == "" {
		props = CursorProps{
			Color:    "",
			Style:    "blink",
		}
	}
	
	model := NewCursorModel(props, id)
	return &CursorComponentWrapper{
		model: &model,
	}
}

// NewFilePickerComponent creates a new FilePicker component wrapper
func NewFilePickerComponent(config core.RenderConfig, id string) *FilePickerComponentWrapper {
	var props FilePickerProps
	
	// Extract props from config
	if config.Data != nil {
		if dataMap, ok := config.Data.(map[string]interface{}); ok {
			props = ParseFilePickerProps(dataMap)
		}
	}
	
	// Use defaults if no data provided
	if props.CurrentDirectory == "" {
		props = FilePickerProps{
			CurrentDirectory: "./",
			AllowedFileTypes: []string{},
			AllowDirectories: false,
			AllowMultiple: false,
			Height: 0,
			Width: 0,
		}
	}
	
	model := NewFilePickerModel(props, id)
	return &FilePickerComponentWrapper{
		model: &model,
	}
}

// NewHelpComponent creates a new Help component wrapper
func NewHelpComponent(config core.RenderConfig, id string) *HelpComponentWrapper {
	var props HelpProps

	// Extract props from config
	if config.Data != nil {
		if dataMap, ok := config.Data.(map[string]interface{}); ok {
			props = ParseHelpProps(dataMap)
		}
	}

	// Use defaults if no data provided (check both KeyMap and Sections)
	if len(props.KeyMap) == 0 && len(props.Sections) == 0 {
		props = HelpProps{
			KeyMap:      map[string]interface{}{},
			Width:       0,
			Height:      0,
			ShowAllKeys: false,
			Style:       "compact",
		}
	}

	model := NewHelpModel(props, id)
	return &HelpComponentWrapper{
		model: &model,
	}
}

// NewKeyComponent creates a new Key component wrapper
func NewKeyComponent(config core.RenderConfig, id string) *KeyComponentWrapper {
	var props KeyProps
	
	// Extract props from config
	if config.Data != nil {
		if dataMap, ok := config.Data.(map[string]interface{}); ok {
			props = ParseKeyProps(dataMap)
		}
	}
	
	// Use defaults if no data provided
	if len(props.Keys) == 0 {
		props = KeyProps{
			Keys:        []string{},
			Description:  "",
			Color:       "",
			Background:  "",
			Bold:        false,
		}
	}
	
	model := NewKeyModel(props, id)
	return &KeyComponentWrapper{
		model: &model,
	}
}


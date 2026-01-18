package components

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/yaoapp/yao/tui/core"
)

// Note: NewHeaderComponent and NewTextComponent are already defined in their respective files
// These factory functions only exist for components that don't have their own
// New*Component(id) functions

// NewFooterComponent creates a new Footer component wrapper
func NewFooterComponent(id string) *FooterComponentWrapper {
	// Create empty props - will be populated during Render()
	props := FooterProps{}
	model := NewFooterModel(props, id)
	return &FooterComponentWrapper{
		model: &model,
	}
}

// NewInputComponent creates a new Input component wrapper
func NewInputComponent(id string) *InputComponentWrapper {
	// Create empty props - will be populated during Render()
	props := InputProps{
		Placeholder: "",
		Value:       "",
		Prompt:      "> ",
	}
	model := NewInputModel(props, id)
	return &InputComponentWrapper{
		model: &model,
	}
}

// NewTextareaComponent creates a new Textarea component wrapper
func NewTextareaComponent(id string) *TextareaComponentWrapper {
	// Create empty props - will be populated during Render()
	props := TextareaProps{
		Placeholder: "",
		Value:       "",
		Prompt:      "> ",
	}
	model := NewTextareaModel(props, id)
	return &TextareaComponentWrapper{
		model: &model,
	}
}

// NewMenuComponent creates a new Menu component wrapper
func NewMenuComponent(id string) *MenuComponentWrapper {
	// Create empty props - will be populated during Render()
	props := MenuProps{
		Items:     []MenuItem{},
		Title:     "",
		Height:    0,
		Width:     0,
		Focused:   false,
		ShowStatusBar: true,
		ShowFilter:    false,
	}
	model := NewMenuInteractiveModel(props)
	model.ID = id
	return &MenuComponentWrapper{
		model: &model,
	}
}

// NewTableComponent creates a new Table component wrapper
func NewTableComponent(id string) *TableComponentWrapper {
	// Create empty props - will be populated during Render()
	props := TableProps{
		Columns: []Column{},
		Data:    [][]interface{}{},
		Focused:  false,
		Height:   0,
		Width:    0,
	}
	model := NewTableModel(props, id)
	return &TableComponentWrapper{
		model: &model,
	}
}

// NewFormComponent creates a new Form component wrapper
func NewFormComponent(id string) *FormComponentWrapper {
	// Create empty props - will be populated during Render()
	props := FormProps{
		Fields:       []Field{},
		Title:        "",
		Description:   "",
		SubmitLabel:   "Submit",
		CancelLabel:   "Cancel",
	}
	model := NewFormModel(props, id)
	return &FormComponentWrapper{
		model: &model,
	}
}

// NewListComponent creates a new List component wrapper
func NewListComponent(id string) *ListComponentWrapper {
	// Create empty props - will be populated during Render()
	props := ListProps{
		Items:   []ListItem{},
		Title:    "",
		Height:   0,
		Width:    0,
		Focused:  false,
	}
	model := NewListModel(props, id)
	return &ListComponentWrapper{
		model: &model,
	}
}

// NewPaginatorComponent creates a new Paginator component wrapper
func NewPaginatorComponent(id string) *PaginatorComponentWrapper {
	// Create empty props - will be populated during Render()
	props := PaginatorProps{
		TotalPages:  1,
		CurrentPage: 1,
		PageSize:     10,
		TotalItems:   0,
		Type:        "dots",
	}
	model := NewPaginatorModel(props, id)
	return &PaginatorComponentWrapper{
		model: &model,
	}
}

// NewViewportComponent creates a new Viewport component wrapper
func NewViewportComponent(id string) *ViewportComponentWrapper {
	// Create empty props - will be populated during Render()
	props := ViewportProps{
		Content:       "",
		Width:         0,
		Height:        0,
		ShowScrollbar: true,
	}
	model := NewViewportModel(props, id)
	return &ViewportComponentWrapper{
		model: &model,
	}
}

// NewChatComponent creates a new Chat component wrapper
func NewChatComponent(id string) *ChatComponentWrapper {
	// Create empty props - will be populated during Render()
	props := ChatProps{
		Messages:        []Message{},
		InputPlaceholder: "Type your message...",
		ShowInput:       true,
		EnableMarkdown:   true,
		GlamourStyle:    "dark",
		Width:           0,
		Height:          0,
		InputHeight:     3,
	}
	model := NewChatModel(props, id)
	return &ChatComponentWrapper{
		model: &model,
	}
}

// NewProgressComponent creates a new Progress component wrapper
func NewProgressComponent(id string) *ProgressComponentWrapper {
	// Create empty props - will be populated during Render()
	props := ProgressProps{
		Percent:     0,
		Width:       0,
		Color:       "",
		Background:  "",
		EmptyColor:  "",
	}
	model := NewProgressModel(props, id)
	return &ProgressComponentWrapper{
		model: &model,
	}
}

// NewSpinnerComponent creates a new Spinner component wrapper
func NewSpinnerComponent(id string) *SpinnerComponentWrapper {
	// Create empty props - will be populated during Render()
	props := SpinnerProps{
		Style: "Dot",
		Color: "",
		Background:  "",
		Speed:     100,
		Frames:    nil,
	}
	model := NewSpinnerModel(props, id)
	return &SpinnerComponentWrapper{
		model: &model,
	}
}

// NewTimerComponent creates a new Timer component wrapper
func NewTimerComponent(id string) *TimerComponentWrapper {
	// Create empty props - will be populated during Render()
	props := TimerProps{
		Duration:   0,
		Format:     "mm:ss",
		Color:      "",
		Background: "",
		Bold:       false,
	}
	model := NewTimerModel(props, id)
	return &TimerComponentWrapper{
		model: &model,
	}
}

// NewStopwatchComponent creates a new Stopwatch component wrapper
func NewStopwatchComponent(id string) *StopwatchComponentWrapper {
	// Create empty props - will be populated during Render()
	props := StopwatchProps{
		Format:     "mm:ss",
		Color:      "",
		Background: "",
		Bold:       false,
		Italic:     false,
	}
	model := NewStopwatchModel(props, id)
	return &StopwatchComponentWrapper{
		model: &model,
	}
}

// NewCursorComponent creates a new Cursor component wrapper
func NewCursorComponent(id string) *CursorComponentWrapper {
	// Create empty props - will be populated during Render()
	props := CursorProps{
		Color:    "",
		Style:    "blink",
	}
	model := NewCursorModel(props, id)
	return &CursorComponentWrapper{
		model: &model,
	}
}

// NewFilePickerComponent creates a new FilePicker component wrapper
func NewFilePickerComponent(id string) *FilePickerComponentWrapper {
	// Create empty props - will be populated during Render()
	props := FilePickerProps{
	}
	model := NewFilePickerModel(props, id)
	return &FilePickerComponentWrapper{
		model: &model,
	}
}

// NewHelpComponent creates a new Help component wrapper
func NewHelpComponent(id string) *HelpComponentWrapper {
	// Create empty props - will be populated during Render()
	props := HelpProps{
		KeyMap:      map[string]interface{}{},
		Width:       0,
		Height:      0,
		ShowAllKeys: false,
		Style:       "compact",
	}
	model := NewHelpModel(props, id)
	return &HelpComponentWrapper{
		model: &model,
	}
}

// NewKeyComponent creates a new Key component wrapper
func NewKeyComponent(id string) *KeyComponentWrapper {
	// Create empty props - will be populated during Render()
	props := KeyProps{
		Keys:        []string{},
		Description:  "",
		Color:       "",
		Background:  "",
		Bold:        false,
	}
	model := NewKeyModel(props, id)
	return &KeyComponentWrapper{
		model: &model,
	}
}

// CRUDComponentWrapper wraps CRUDComponent for unified factory interface
type CRUDComponentWrapper struct {
	component *CRUDComponent
}

// NewCRUDComponentWrapper creates a wrapper around CRUD component
func NewCRUDComponentWrapper(id string) *CRUDComponentWrapper {
	return &CRUDComponentWrapper{
		component: NewCRUDComponent(id, nil),
	}
}

func (w *CRUDComponentWrapper) Init() tea.Cmd {
	return w.component.Init()
}

func (w *CRUDComponentWrapper) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
	return w.component.UpdateMsg(msg)
}

func (w *CRUDComponentWrapper) View() string {
	return w.component.View()
}

func (w *CRUDComponentWrapper) GetID() string {
	return w.component.GetID()
}

func (w *CRUDComponentWrapper) SetFocus(focus bool) {
	w.component.SetFocus(focus)
}

func (w *CRUDComponentWrapper) GetComponentType() string {
	return w.component.GetComponentType()
}

func (w *CRUDComponentWrapper) Render(config core.RenderConfig) (string, error) {
	return w.component.Render(config)
}

func (w *CRUDComponentWrapper) Cleanup() {
	w.component.Cleanup()
}

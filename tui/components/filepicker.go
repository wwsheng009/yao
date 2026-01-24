package components

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yaoapp/yao/tui/core"
)

// FilePickerProps defines the properties for the FilePicker component
type FilePickerProps struct {
	// CurrentDirectory specifies the starting directory
	CurrentDirectory string `json:"currentDirectory"`

	// AllowedFileTypes specifies allowed file extensions
	AllowedFileTypes []string `json:"allowedFileTypes"`

	// AllowDirectories allows directory selection
	AllowDirectories bool `json:"allowDirectories"`

	// AllowMultiple allows multiple file selection
	AllowMultiple bool `json:"allowMultiple"`

	// Height specifies the file picker height
	Height int `json:"height"`

	// Width specifies the file picker width (0 for auto)
	Width int `json:"width"`

	// ShowHiddenFiles shows hidden files
	ShowHiddenFiles bool `json:"showHiddenFiles"`

	// AutoNavigate automatically navigates into directories
	AutoNavigate bool `json:"autoNavigate"`
}

// FilePickerModel wraps the filepicker.Model to handle TUI integration
type FilePickerModel struct {
	filepicker.Model
	props FilePickerProps
	id    string // Unique identifier for this instance
}

// RenderFilePicker renders a file picker component
func RenderFilePicker(props FilePickerProps, width int) string {
	fp := filepicker.New()

	// Set current directory
	if props.CurrentDirectory != "" {
		fp.CurrentDirectory = props.CurrentDirectory
	} else {
		// Use current working directory
		if cwd, err := os.Getwd(); err == nil {
			fp.CurrentDirectory = cwd
		}
	}

	// Set allowed file types
	if len(props.AllowedFileTypes) > 0 {
		fp.AllowedTypes = props.AllowedFileTypes
	}

	// Set height
	if props.Height > 0 {
		fp.Height = props.Height
	}

	// Apply basic styles
	style := lipgloss.NewStyle()

	return style.Render(fp.View())
}

// ParseFilePickerProps converts a generic props map to FilePickerProps using JSON unmarshaling
func ParseFilePickerProps(props map[string]interface{}) FilePickerProps {
	// Set defaults
	fp := FilePickerProps{
		AllowDirectories: true,
		AllowMultiple:    false,
		ShowHiddenFiles:  false,
		AutoNavigate:     true,
		Height:           10,
	}

	// Handle allowed file types
	if types, ok := props["allowedFileTypes"].([]interface{}); ok {
		fp.AllowedFileTypes = make([]string, len(types))
		for i, v := range types {
			if str, ok := v.(string); ok {
				fp.AllowedFileTypes[i] = str
			}
		}
	}

	// Unmarshal remaining properties
	if dataBytes, err := json.Marshal(props); err == nil {
		_ = json.Unmarshal(dataBytes, &fp)
	}

	return fp
}

// NewFilePickerModel creates a new FilePickerModel from FilePickerProps
func NewFilePickerModel(props FilePickerProps, id string) FilePickerModel {
	fp := filepicker.New()

	// Set current directory
	if props.CurrentDirectory != "" {
		fp.CurrentDirectory = props.CurrentDirectory
	} else {
		// Use current working directory
		if cwd, err := os.Getwd(); err == nil {
			fp.CurrentDirectory = cwd
		}
	}

	// Set allowed file types
	if len(props.AllowedFileTypes) > 0 {
		fp.AllowedTypes = props.AllowedFileTypes
	}

	// Set height
	if props.Height > 0 {
		fp.Height = props.Height
	}

	return FilePickerModel{
		Model: fp,
		props: props,
		id:    id,
	}
}

// HandleFilePickerUpdate handles updates for file picker components
func HandleFilePickerUpdate(msg tea.Msg, filePickerModel *FilePickerModel) (FilePickerModel, tea.Cmd) {
	if filePickerModel == nil {
		return FilePickerModel{}, nil
	}

	var cmd tea.Cmd
	filePickerModel.Model, cmd = filePickerModel.Model.Update(msg)
	return *filePickerModel, cmd
}

// Init initializes the file picker model
func (m *FilePickerModel) Init() tea.Cmd {
	return m.Model.Init()
}

// View returns the string representation of the file picker
func (m *FilePickerModel) View() string {
	return m.Model.View()
}

// GetID returns the unique identifier for this component instance
func (m *FilePickerModel) GetID() string {
	return m.id
}

// GetSelectedFiles returns the selected files
func (m *FilePickerModel) GetSelectedFiles() []string {
	// FilePicker doesn't have SelectedFiles field, return current selection
	if m.Path != "" {
		return []string{m.Path}
	}
	return []string{}
}

// SetCurrentDirectory sets the current directory
func (m *FilePickerModel) SetCurrentDirectory(dir string) {
	m.CurrentDirectory = dir
}

// FilePickerComponentWrapper wraps the native filepicker.Model directly
type FilePickerComponentWrapper struct {
	model filepicker.Model
	props FilePickerProps
	id    string
	focus bool
}

// NewFilePickerComponentWrapper creates a wrapper that implements ComponentInterface
func NewFilePickerComponentWrapper(props FilePickerProps, id string) *FilePickerComponentWrapper {
	fp := filepicker.New()

	// Set current directory
	if props.CurrentDirectory != "" {
		fp.CurrentDirectory = props.CurrentDirectory
	} else {
		// Use current working directory
		if cwd, err := os.Getwd(); err == nil {
			fp.CurrentDirectory = cwd
		}
	}

	// Set allowed file types
	if len(props.AllowedFileTypes) > 0 {
		fp.AllowedTypes = props.AllowedFileTypes
	}

	// Set height
	if props.Height > 0 {
		fp.Height = props.Height
	}

	return &FilePickerComponentWrapper{
		model: fp,
		props: props,
		id:    id,
	}
}

func (w *FilePickerComponentWrapper) Init() tea.Cmd {
	// 不要在初始化时自动获取焦点
	// 焦点应该通过框架的焦点管理机制来控制
	// 只有当组件被明确设置焦点时才获取焦点
	// 我们仍然需要初始化底层模型，但不获取焦点
	w.model.Init() // 初始化模型但不返回命令
	return nil
}

func (w *FilePickerComponentWrapper) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
	// Handle targeted messages first
	switch msg := msg.(type) {
	case core.TargetedMsg:
		// Check if this message is targeted to this component
		if msg.TargetID == w.id {
			return w.UpdateMsg(msg.InnerMsg)
		}
		return w, nil, core.Ignored
	case tea.KeyMsg:
		// Handle ESC key for blur
		if msg.Type == tea.KeyEsc && w.focus {
			// Release focus when ESC is pressed
			w.focus = false
			eventCmd := core.PublishEvent(w.id, core.EventEscapePressed, nil)
			return w, eventCmd, core.Ignored
		}
	}

	// For file picker, update the model
	var cmd tea.Cmd
	oldPath := w.model.Path

	w.model, cmd = w.model.Update(msg)

	newPath := w.model.Path

	// Check if selection changed
	if oldPath != newPath {
		eventCmd := core.PublishEvent(w.id, "FILE_SELECTION_CHANGED", map[string]interface{}{
			"selectedFiles":    w.GetSelectedFiles(),
			"currentDirectory": w.model.CurrentDirectory,
		})
		if cmd != nil {
			return w, tea.Batch(cmd, eventCmd), core.Handled
		}
		return w, eventCmd, core.Handled
	}

	return w, cmd, core.Handled
}

func (w *FilePickerComponentWrapper) View() string {
	return w.model.View()
}

func (w *FilePickerComponentWrapper) GetID() string {
	return w.id
}

func (w *FilePickerComponentWrapper) SetFocus(focus bool) {
	w.focus = focus
}

func (w *FilePickerComponentWrapper) GetFocus() bool {
	return w.focus
}

// SetSize sets the allocated size for the component.
func (w *FilePickerComponentWrapper) SetSize(width, height int) {
	// Default implementation: store size if component has width/height fields
	// Components can override this to handle size changes
}

func (w *FilePickerComponentWrapper) GetComponentType() string {
	return "filepicker"
}

func (w *FilePickerComponentWrapper) Render(config core.RenderConfig) (string, error) {
	// Parse configuration data
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("FilePickerComponentWrapper: invalid data type")
	}

	// Parse filepicker properties
	props := ParseFilePickerProps(propsMap)

	// Update component properties
	w.props = props

	// Update current directory if changed
	if props.CurrentDirectory != "" {
		w.model.CurrentDirectory = props.CurrentDirectory
	}

	// Update allowed file types if changed
	if len(props.AllowedFileTypes) > 0 {
		w.model.AllowedTypes = props.AllowedFileTypes
	}

	// Update height if changed
	if props.Height > 0 {
		w.model.Height = props.Height
	}

	return w.View(), nil
}

func (w *FilePickerComponentWrapper) UpdateRenderConfig(config core.RenderConfig) error {
	// Parse configuration data
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("FilePickerComponentWrapper: invalid data type")
	}

	// Parse filepicker properties
	props := ParseFilePickerProps(propsMap)

	// Update component properties
	w.props = props

	// Update current directory if changed
	if props.CurrentDirectory != "" {
		w.model.CurrentDirectory = props.CurrentDirectory
	}

	// Update allowed file types if changed
	if len(props.AllowedFileTypes) > 0 {
		w.model.AllowedTypes = props.AllowedFileTypes
	}

	// Update height if changed
	if props.Height > 0 {
		w.model.Height = props.Height
	}

	return nil
}

// GetSelectedFiles returns the selected files
func (w *FilePickerComponentWrapper) GetSelectedFiles() []string {
	// FilePicker doesn't have SelectedFiles field, return current selection
	if w.model.Path != "" {
		return []string{w.model.Path}
	}
	return []string{}
}

// SetCurrentDirectory sets the current directory
func (w *FilePickerComponentWrapper) SetCurrentDirectory(dir string) {
	w.model.CurrentDirectory = dir
}

func (w *FilePickerComponentWrapper) Cleanup() {
	// FilePicker组件通常不需要清理资源
	// 这是一个空操作
}

// GetStateChanges returns the state changes from this component
func (w *FilePickerComponentWrapper) GetStateChanges() (map[string]interface{}, bool) {
	// FilePicker component state tracking
	return nil, false
}

// GetSubscribedMessageTypes returns the message types this component subscribes to
func (w *FilePickerComponentWrapper) GetSubscribedMessageTypes() []string {
	return []string{
		"core.TargetedMsg",
	}
}

func (m *FilePickerModel) UpdateRenderConfig(config core.RenderConfig) error {
	// Parse configuration data
	propsMap, ok := config.Data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("FilePickerModel: invalid data type")
	}

	// Parse filepicker properties
	props := ParseFilePickerProps(propsMap)

	// Update component properties
	m.props = props

	return nil
}

func (m *FilePickerModel) Cleanup() {
	// FilePicker模型通常不需要清理资源
	// 这是一个空操作
}

func (m *FilePickerModel) Render(config core.RenderConfig) (string, error) {
	// This method is kept for backward compatibility
	// It now delegates to UpdateRenderConfig
	_ = m.UpdateRenderConfig(config)
	return m.View(), nil
}

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
		fp.SetHeight(props.Height)
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
		fp.SetHeight(props.Height)
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

// FilePickerComponentWrapper wraps FilePickerModel to implement ComponentInterface properly
type FilePickerComponentWrapper struct {
	model *FilePickerModel
}

// NewFilePickerComponentWrapper creates a wrapper that implements ComponentInterface
func NewFilePickerComponentWrapper(filePickerModel *FilePickerModel) *FilePickerComponentWrapper {
	return &FilePickerComponentWrapper{
		model: filePickerModel,
	}
}

func (w *FilePickerComponentWrapper) Init() tea.Cmd {
	return w.model.Init()
}

func (w *FilePickerComponentWrapper) UpdateMsg(msg tea.Msg) (core.ComponentInterface, tea.Cmd, core.Response) {
	// Handle targeted messages first
	switch msg := msg.(type) {
	case core.TargetedMsg:
		// Check if this message is targeted to this component
		if msg.TargetID == w.model.id {
			return w.UpdateMsg(msg.InnerMsg)
		}
		return w, nil, core.Ignored
	}

	// For file picker, update the model
	var cmd tea.Cmd
	oldPath := w.model.Path

	w.model.Model, cmd = w.model.Model.Update(msg)

	newPath := w.model.Path

	// Check if selection changed
	if oldPath != newPath {
		eventCmd := core.PublishEvent(w.model.id, "FILE_SELECTION_CHANGED", map[string]interface{}{
			"selectedFiles":    w.model.GetSelectedFiles(),
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
	return w.model.id
}

func (w *FilePickerComponentWrapper) SetFocus(focus bool) {
	// File picker focus is handled internally
}

func (w *FilePickerComponentWrapper) GetComponentType() string {
	return "filepicker"
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

func (w *FilePickerComponentWrapper) Render(config core.RenderConfig) (string, error) {
	return w.model.Render(config)
}

func (w *FilePickerComponentWrapper) UpdateRenderConfig(config core.RenderConfig) error {
	// 委托给底层的 FilePickerModel
	return w.model.UpdateRenderConfig(config)
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

package selection

import (
	"github.com/yaoapp/yao/tui/runtime"
	"github.com/yaoapp/yao/tui/runtime/event"
)

// TextSelection provides a complete text selection system for TUI applications.
// It combines mouse selection, keyboard shortcuts, clipboard integration, and
// selection highlighting into a single, easy-to-use interface.
type TextSelection struct {
	controller   *SelectionController
	mouseHandler *MouseHandler
	keyHandler   *KeyboardHandler
	renderer     *Renderer
	manager      *Manager
	clipboard    *Clipboard
	enabled      bool
	buffer       *runtime.CellBuffer
}

// NewTextSelection creates a new text selection system.
func NewTextSelection(buffer *runtime.CellBuffer) *TextSelection {
	// Create the selection manager with the buffer
	adapter := NewTextBufferAdapter(buffer)
	manager := NewManager(adapter)

	// Create clipboard
	clipboard := NewClipboard()

	// Create components
	mouseHandler := NewMouseHandler(manager)
	keyHandler := NewKeyboardHandler(manager, clipboard)
	renderer := NewRenderer(manager, buffer)
	controller := &SelectionController{
		handler:   mouseHandler,
		manager:   manager,
		clipboard: clipboard,
		enabled:   true,
	}

	return &TextSelection{
		controller:   controller,
		mouseHandler: mouseHandler,
		keyHandler:   keyHandler,
		renderer:     renderer,
		manager:      manager,
		clipboard:    clipboard,
		enabled:      true,
		buffer:       buffer,
	}
}

// HandleEvent processes an event for the text selection system.
// It handles both mouse and keyboard events.
func (s *TextSelection) HandleEvent(ev interface{}) bool {
	if !s.enabled {
		return false
	}

	switch e := ev.(type) {
	case *event.MouseEvent:
		return s.mouseHandler.HandleMouseEvent(e)
	case *event.KeyEvent:
		return s.keyHandler.HandleKeyEvent(e)
	}

	return false
}

// HandleMouseEvent processes a mouse event.
func (s *TextSelection) HandleMouseEvent(ev *event.MouseEvent) bool {
	if !s.enabled {
		return false
	}
	return s.mouseHandler.HandleMouseEvent(ev)
}

// HandleKeyEvent processes a keyboard event.
func (s *TextSelection) HandleKeyEvent(ev *event.KeyEvent) bool {
	if !s.enabled {
		return false
	}
	return s.keyHandler.HandleKeyEvent(ev)
}

// ApplySelection applies the selection highlight to the buffer.
// Call this after rendering to add selection highlights.
func (s *TextSelection) ApplySelection() {
	if !s.enabled {
		return
	}
	s.renderer.ApplySelection()
}

// Copy copies the selected text to the clipboard.
func (s *TextSelection) Copy() (string, error) {
	return s.controller.Copy()
}

// GetSelectedText returns the selected text.
func (s *TextSelection) GetSelectedText() string {
	return s.manager.GetSelectedTextCompact()
}

// GetSelectedTextRaw returns the selected text with trailing whitespace.
func (s *TextSelection) GetSelectedTextRaw() string {
	return s.manager.GetSelectedText()
}

// IsActive returns whether a selection is active.
func (s *TextSelection) IsActive() bool {
	return s.manager.IsActive()
}

// IsSelected returns whether the cell at (x, y) is selected.
func (s *TextSelection) IsSelected(x, y int) bool {
	return s.manager.IsSelected(x, y)
}

// GetSelectionRange returns the normalized selection range.
func (s *TextSelection) GetSelectionRange() (startX, endX, startY, endY int) {
	return s.manager.GetSelectionRange()
}

// Clear clears the current selection.
func (s *TextSelection) Clear() {
	s.manager.Clear()
}

// SelectAll selects all text in the buffer.
func (s *TextSelection) SelectAll() {
	s.manager.SelectAll()
}

// SelectWord selects the word at the given position.
func (s *TextSelection) SelectWord(x, y int) {
	s.manager.SelectWord(x, y)
}

// SelectLine selects the line at the given Y position.
func (s *TextSelection) SelectLine(y int) {
	s.manager.SelectLine(y)
}

// SetEnabled enables or disables text selection.
func (s *TextSelection) SetEnabled(enabled bool) {
	s.enabled = enabled
	s.mouseHandler.SetEnabled(enabled)
	s.keyHandler.SetEnabled(enabled)
	s.controller.SetEnabled(enabled)

	if !enabled {
		s.Clear()
	}
}

// IsEnabled returns whether text selection is enabled.
func (s *TextSelection) IsEnabled() bool {
	return s.enabled
}

// SetHighlightStyle sets the style used for selection highlighting.
func (s *TextSelection) SetHighlightStyle(style CellStyle) {
	s.renderer.SetHighlightStyle(style)
}

// GetHighlightStyle returns the current highlight style.
func (s *TextSelection) GetHighlightStyle() CellStyle {
	return s.renderer.GetHighlightStyle()
}

// SetSelectionMode sets the selection mode.
func (s *TextSelection) SetSelectionMode(mode SelectionMode) {
	s.manager.SetMode(mode)
}

// GetSelectionMode returns the current selection mode.
func (s *TextSelection) GetSelectionMode() SelectionMode {
	return s.manager.GetMode()
}

// GetManager returns the underlying selection manager.
func (s *TextSelection) GetManager() *Manager {
	return s.manager
}

// GetClipboard returns the clipboard instance.
func (s *TextSelection) GetClipboard() *Clipboard {
	return s.clipboard
}

// IsClipboardSupported returns whether clipboard is supported on this platform.
func (s *TextSelection) IsClipboardSupported() bool {
	return s.clipboard.IsSupported()
}

// UpdateBuffer updates the buffer used for selection.
// Call this when the buffer is recreated or resized.
func (s *TextSelection) UpdateBuffer(buffer *runtime.CellBuffer) {
	s.buffer = buffer
	adapter := NewTextBufferAdapter(buffer)
	s.manager.SetBuffer(adapter)
	s.renderer = NewRenderer(s.manager, buffer)
}

// GetSelectedCells returns all cells in the current selection.
func (s *TextSelection) GetSelectedCells() []struct{ X, Y int } {
	return s.manager.GetSelectedCells()
}

// GetSelectionRegion returns the selection as a SelectionRegion.
func (s *TextSelection) GetSelectionRegion() SelectionRegion {
	return s.manager.GetRegion()
}

// ExtendSelection extends the selection to the given position.
func (s *TextSelection) ExtendSelection(x, y int) {
	s.mouseHandler.ExtendSelection(x, y)
}

// MoveSelectionStart moves the selection start by the given delta.
func (s *TextSelection) MoveSelectionStart(dx, dy int) {
	s.manager.MoveStart(dx, dy)
}

// MoveSelectionEnd moves the selection end by the given delta.
func (s *TextSelection) MoveSelectionEnd(dx, dy int) {
	s.manager.MoveEnd(dx, dy)
}

// IsDragging returns whether a drag operation is in progress.
func (s *TextSelection) IsDragging() bool {
	return s.mouseHandler.IsDragging()
}

// SelectionConfig holds configuration for text selection.
type SelectionConfig struct {
	Enabled         bool
	HighlightStyle  CellStyle
	SelectionMode   SelectionMode
	EnableMouse     bool
	EnableKeyboard  bool
	EnableClipboard bool
}

// DefaultSelectionConfig returns the default selection configuration.
func DefaultSelectionConfig() SelectionConfig {
	return SelectionConfig{
		Enabled:         true,
		HighlightStyle:  DefaultHighlightStyle(),
		SelectionMode:   SelectionModeChar,
		EnableMouse:     true,
		EnableKeyboard:  true,
		EnableClipboard: true,
	}
}

// NewTextSelectionWithConfig creates a new text selection system with custom configuration.
func NewTextSelectionWithConfig(buffer *runtime.CellBuffer, config SelectionConfig) *TextSelection {
	selection := NewTextSelection(buffer)

	if !config.Enabled {
		selection.SetEnabled(false)
	}

	selection.SetHighlightStyle(config.HighlightStyle)
	selection.SetSelectionMode(config.SelectionMode)

	// Note: Mouse and keyboard can't be individually disabled in current implementation
	// They can be re-enabled if needed

	return selection
}

// GlobalSelectionManager provides a global instance for convenience.
// This can be used by applications that don't need multiple independent selection systems.
var GlobalSelectionManager *TextSelection

// InitGlobalSelection initializes the global selection manager with a buffer.
func InitGlobalSelection(buffer *runtime.CellBuffer) {
	GlobalSelectionManager = NewTextSelection(buffer)
}

// GetGlobalSelection returns the global selection manager.
func GetGlobalSelection() *TextSelection {
	return GlobalSelectionManager
}

// CopyToClipboardGlobal copies the selected text using the global selection manager.
func CopyToClipboardGlobal() (string, error) {
	if GlobalSelectionManager == nil {
		return "", nil
	}
	return GlobalSelectionManager.Copy()
}

// GetSelectedTextGlobal returns the selected text from the global selection manager.
func GetSelectedTextGlobal() string {
	if GlobalSelectionManager == nil {
		return ""
	}
	return GlobalSelectionManager.GetSelectedText()
}

// ClearSelectionGlobal clears the selection in the global selection manager.
func ClearSelectionGlobal() {
	if GlobalSelectionManager != nil {
		GlobalSelectionManager.Clear()
	}
}

// IsSelectionActiveGlobal returns whether there's an active selection in the global manager.
func IsSelectionActiveGlobal() bool {
	if GlobalSelectionManager == nil {
		return false
	}
	return GlobalSelectionManager.IsActive()
}

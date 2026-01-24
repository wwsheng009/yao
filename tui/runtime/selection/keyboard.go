package selection

import (
	"github.com/yaoapp/yao/tui/runtime/event"
)

// KeyboardHandler handles keyboard shortcuts for text selection operations.
// It supports common shortcuts like Ctrl+C (copy), Ctrl+A (select all), etc.
type KeyboardHandler struct {
	manager    *Manager
	clipboard  *Clipboard
	enabled    bool
}

// NewKeyboardHandler creates a new keyboard handler for selection operations.
func NewKeyboardHandler(manager *Manager, clipboard *Clipboard) *KeyboardHandler {
	return &KeyboardHandler{
		manager:   manager,
		clipboard: clipboard,
		enabled:   true,
	}
}

// SetEnabled enables or disables keyboard shortcuts.
func (h *KeyboardHandler) SetEnabled(enabled bool) {
	h.enabled = enabled
}

// IsEnabled returns whether keyboard shortcuts are enabled.
func (h *KeyboardHandler) IsEnabled() bool {
	return h.enabled
}

// HandleKeyEvent processes a keyboard event for selection operations.
// Returns true if the event was handled.
func (h *KeyboardHandler) HandleKeyEvent(ev *event.KeyEvent) bool {
	if !h.enabled || h.manager == nil || h.clipboard == nil {
		return false
	}

	// Check for Ctrl modifier
	if ev.Type != event.KeyPress || ev.Mod != event.ModCtrl {
		return false
	}

	switch ev.Key {
	case 'c', 'C':
		// Ctrl+C - Copy selection to clipboard
		return h.handleCopy()
	case 'x', 'X':
		// Ctrl+X - Cut (copy and clear selection)
		return h.handleCut()
	case 'a', 'A':
		// Ctrl+A - Select all
		return h.handleSelectAll()
	case 'd', 'D':
		// Ctrl+D - Select word (some terminals use this)
		return h.handleSelectWord()
	case 27:
		// Escape - Clear selection
		return h.handleEscape()
	}

	return false
}

// handleCopy copies the selected text to clipboard.
func (h *KeyboardHandler) handleCopy() bool {
	if !h.manager.IsActive() {
		return false
	}

	text := h.manager.GetSelectedTextCompact()
	if text == "" {
		return false
	}

	err := h.clipboard.Copy(text)
	return err == nil
}

// handleCut copies the selection and clears it.
func (h *KeyboardHandler) handleCut() bool {
	if !h.handleCopy() {
		return false
	}

	// Clear the selection after copying
	h.manager.Clear()
	return true
}

// handleSelectAll selects all text in the buffer.
func (h *KeyboardHandler) handleSelectAll() bool {
	h.manager.SelectAll()
	return true
}

// handleSelectWord extends the selection by word.
// This is a placeholder for more sophisticated word selection.
func (h *KeyboardHandler) handleSelectWord() bool {
	// For now, just return true as the action would need context
	// about where the cursor is to extend the selection
	return false
}

// handleEscape clears the current selection.
func (h *KeyboardHandler) handleEscape() bool {
	h.manager.Clear()
	return true
}

// CopyToClipboard is a convenience method to copy the current selection.
func (h *KeyboardHandler) CopyToClipboard() (string, error) {
	if h.manager == nil || h.clipboard == nil {
		return "", nil
	}

	text := h.manager.GetSelectedTextCompact()
	if text == "" {
		return "", nil
	}

	err := h.clipboard.Copy(text)
	return text, err
}

// GetSelectedText returns the currently selected text.
func (h *KeyboardHandler) GetSelectedText() string {
	if h.manager == nil {
		return ""
	}
	return h.manager.GetSelectedText()
}

// ClearSelection clears the current selection.
func (h *KeyboardHandler) ClearSelection() {
	if h.manager != nil {
		h.manager.Clear()
	}
}

// HasSelection returns whether there is an active selection.
func (h *KeyboardHandler) HasSelection() bool {
	return h.manager != nil && h.manager.IsActive()
}

// KeyBindings defines keyboard shortcuts for selection operations.
type KeyBindings struct {
	Copy      rune
	Cut       rune
	SelectAll rune
	Escape    rune
}

// DefaultKeyBindings returns the default key bindings.
func DefaultKeyBindings() KeyBindings {
	return KeyBindings{
		Copy:      'c',
		Cut:       'x',
		SelectAll: 'a',
		Escape:    27,
	}
}

// ConfigurableKeyboardHandler is a keyboard handler with configurable key bindings.
type ConfigurableKeyboardHandler struct {
	manager     *Manager
	clipboard   *Clipboard
	enabled     bool
	bindings    KeyBindings
}

// NewConfigurableKeyboardHandler creates a new configurable keyboard handler.
func NewConfigurableKeyboardHandler(manager *Manager, clipboard *Clipboard, bindings KeyBindings) *ConfigurableKeyboardHandler {
	if bindings.Copy == 0 {
		bindings.Copy = 'c'
	}
	if bindings.Cut == 0 {
		bindings.Cut = 'x'
	}
	if bindings.SelectAll == 0 {
		bindings.SelectAll = 'a'
	}
	if bindings.Escape == 0 {
		bindings.Escape = 27
	}

	return &ConfigurableKeyboardHandler{
		manager:   manager,
		clipboard: clipboard,
		enabled:   true,
		bindings:  bindings,
	}
}

// HandleKeyEvent processes a keyboard event with configurable bindings.
func (h *ConfigurableKeyboardHandler) HandleKeyEvent(ev *event.KeyEvent) bool {
	if !h.enabled || h.manager == nil || h.clipboard == nil {
		return false
	}

	if ev.Type != event.KeyPress || ev.Mod != event.ModCtrl {
		return false
	}

	switch ev.Key {
	case h.bindings.Copy:
		return h.handleCopy()
	case h.bindings.Cut:
		return h.handleCut()
	case h.bindings.SelectAll:
		return h.handleSelectAll()
	case h.bindings.Escape:
		return h.handleEscape()
	}

	return false
}

func (h *ConfigurableKeyboardHandler) handleCopy() bool {
	if !h.manager.IsActive() {
		return false
	}

	text := h.manager.GetSelectedTextCompact()
	if text == "" {
		return false
	}

	err := h.clipboard.Copy(text)
	return err == nil
}

func (h *ConfigurableKeyboardHandler) handleCut() bool {
	if !h.handleCopy() {
		return false
	}
	h.manager.Clear()
	return true
}

func (h *ConfigurableKeyboardHandler) handleSelectAll() bool {
	h.manager.SelectAll()
	return true
}

func (h *ConfigurableKeyboardHandler) handleEscape() bool {
	h.manager.Clear()
	return true
}

// SetBindings updates the key bindings.
func (h *ConfigurableKeyboardHandler) SetBindings(bindings KeyBindings) {
	h.bindings = bindings
}

// GetBindings returns the current key bindings.
func (h *ConfigurableKeyboardHandler) GetBindings() KeyBindings {
	return h.bindings
}

// SetEnabled enables or disables the keyboard handler.
func (h *ConfigurableKeyboardHandler) SetEnabled(enabled bool) {
	h.enabled = enabled
}

// IsEnabled returns whether the keyboard handler is enabled.
func (h *ConfigurableKeyboardHandler) IsEnabled() bool {
	return h.enabled
}

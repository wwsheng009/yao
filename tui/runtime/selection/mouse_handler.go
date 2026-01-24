package selection

import (
	"github.com/yaoapp/yao/tui/runtime/event"
)

// MouseHandler manages mouse-based text selection.
// It tracks mouse button states and coordinates to handle
// click, drag, and release operations for text selection.
type MouseHandler struct {
	manager      *Manager
	isDragging   bool
	dragStartX   int
	dragStartY   int
	lastX        int
	lastY        int
	clickCount   int      // For double/triple click detection
	lastClickTime int64    // For double/triple click detection
	lastClickX   int
	lastClickY   int
	enabled      bool
}

// NewMouseHandler creates a new mouse handler for text selection.
func NewMouseHandler(manager *Manager) *MouseHandler {
	return &MouseHandler{
		manager:    manager,
		isDragging: false,
		enabled:    true,
	}
}

// SetEnabled enables or disables mouse selection.
func (h *MouseHandler) SetEnabled(enabled bool) {
	h.enabled = enabled
	if !enabled {
		h.isDragging = false
	}
}

// IsEnabled returns whether mouse selection is enabled.
func (h *MouseHandler) IsEnabled() bool {
	return h.enabled
}

// HandleMouseEvent processes a mouse event for text selection.
// Returns true if the event was handled.
func (h *MouseHandler) HandleMouseEvent(ev *event.MouseEvent) bool {
	if !h.enabled || h.manager == nil {
		return false
	}

	switch ev.Type {
	case event.MousePress:
		return h.handlePress(ev)
	case event.MouseRelease:
		return h.handleRelease(ev)
	case event.MouseMove:
		return h.handleMove(ev)
	}

	return false
}

// handlePress handles mouse button press events.
func (h *MouseHandler) handlePress(ev *event.MouseEvent) bool {
	if ev.Click != event.MouseLeft {
		return false
	}

	// Check for double/triple click
	h.clickCount++

	// Start selection
	h.isDragging = true
	h.dragStartX = ev.X
	h.dragStartY = ev.Y
	h.lastX = ev.X
	h.lastY = ev.Y

	switch h.clickCount {
	case 1:
		// Single click - start character selection
		h.manager.SetMode(SelectionModeChar)
		h.manager.Start(ev.X, ev.Y)
	case 2:
		// Double click - select word
		h.manager.SetMode(SelectionModeWord)
		h.manager.SelectWord(ev.X, ev.Y)
	case 3:
		// Triple click - select line
		h.manager.SetMode(SelectionModeLine)
		h.manager.SelectLine(ev.Y)
		h.clickCount = 0 // Reset for quadruple click
	}

	return true
}

// handleRelease handles mouse button release events.
func (h *MouseHandler) handleRelease(ev *event.MouseEvent) bool {
	if ev.Click != event.MouseLeft {
		return false
	}

	if h.isDragging {
		h.isDragging = false
		return true
	}

	return false
}

// handleMove handles mouse move events.
func (h *MouseHandler) handleMove(ev *event.MouseEvent) bool {
	if !h.isDragging {
		return false
	}

	// Only update if position changed
	if ev.X != h.lastX || ev.Y != h.lastY {
		h.manager.Update(ev.X, ev.Y)
		h.lastX = ev.X
		h.lastY = ev.Y
		return true
	}

	return false
}

// IsDragging returns whether a drag operation is in progress.
func (h *MouseHandler) IsDragging() bool {
	return h.isDragging
}

// GetDragStart returns the starting position of the current drag.
func (h *MouseHandler) GetDragStart() (x, y int) {
	return h.dragStartX, h.dragStartY
}

// Clear resets the mouse handler state.
func (h *MouseHandler) Clear() {
	h.isDragging = false
	h.dragStartX = 0
	h.dragStartY = 0
	h.lastX = 0
	h.lastY = 0
	h.clickCount = 0
	h.manager.Clear()
}

// ExtendSelection extends the current selection to the new position.
// This is used for Shift+Click selections.
func (h *MouseHandler) ExtendSelection(x, y int) {
	if h.manager.IsActive() {
		h.manager.Extend(x, y)
	} else {
		h.manager.Start(x, y)
	}
	h.lastX = x
	h.lastY = y
}

// SelectionController provides a high-level interface for text selection.
type SelectionController struct {
	handler    *MouseHandler
	manager    *Manager
	clipboard  *Clipboard
	enabled    bool
}

// NewSelectionController creates a new selection controller.
func NewSelectionController(buffer TextBuffer) *SelectionController {
	manager := NewManager(buffer)
	handler := NewMouseHandler(manager)
	clipboard := NewClipboard()

	return &SelectionController{
		handler:   handler,
		manager:   manager,
		clipboard: clipboard,
		enabled:   true,
	}
}

// NewSelectionControllerWithBuffer creates a selection controller from a runtime CellBuffer.
func NewSelectionControllerWithBuffer(buffer interface {
	GetCell(x, y int) struct {
		Char   rune
		Style  interface{}
		ZIndex int
		NodeID string
	}
	Width() int
	Height() int
}) *SelectionController {
	adapter := &cellBufferWrapper{buffer: buffer}
	return NewSelectionController(adapter)
}

// cellBufferWrapper wraps runtime.CellBuffer to implement TextBuffer.
type cellBufferWrapper struct {
	buffer interface {
		GetCell(x, y int) struct {
			Char   rune
			Style  interface{}
			ZIndex int
			NodeID string
		}
		Width() int
		Height() int
	}
}

func (w *cellBufferWrapper) GetCell(x, y int) Cell {
	cell := w.buffer.GetCell(x, y)
	return Cell{
		Char:  cell.Char,
		Empty: cell.Char == ' ' || cell.Char == 0,
	}
}

func (w *cellBufferWrapper) Width() int {
	return w.buffer.Width()
}

func (w *cellBufferWrapper) Height() int {
	return w.buffer.Height()
}

// HandleEvent handles a mouse event through the controller.
func (c *SelectionController) HandleEvent(ev *event.MouseEvent) bool {
	if !c.enabled {
		return false
	}
	return c.handler.HandleMouseEvent(ev)
}

// GetManager returns the selection manager.
func (c *SelectionController) GetManager() *Manager {
	return c.manager
}

// GetClipboard returns the clipboard instance.
func (c *SelectionController) GetClipboard() *Clipboard {
	return c.clipboard
}

// Copy copies the selected text to the clipboard.
func (c *SelectionController) Copy() (string, error) {
	return CopySelectionToClipboard(c.manager, c.clipboard)
}

// IsEnabled returns whether selection is enabled.
func (c *SelectionController) IsEnabled() bool {
	return c.enabled
}

// SetEnabled enables or disables selection.
func (c *SelectionController) SetEnabled(enabled bool) {
	c.enabled = enabled
	c.handler.SetEnabled(enabled)
	if !enabled {
		c.manager.Clear()
	}
}

// Clear clears the current selection.
func (c *SelectionController) Clear() {
	c.handler.Clear()
}

// IsActive returns whether a selection is active.
func (c *SelectionController) IsActive() bool {
	return c.manager.IsActive()
}

// GetSelectedText returns the selected text.
func (c *SelectionController) GetSelectedText() string {
	return c.manager.GetSelectedText()
}

// SelectAll selects all text.
func (c *SelectionController) SelectAll() {
	c.manager.SelectAll()
}

// SelectWord selects the word at the given position.
func (c *SelectionController) SelectWord(x, y int) {
	c.manager.SelectWord(x, y)
}

// SelectLine selects the line at the given Y position.
func (c *SelectionController) SelectLine(y int) {
	c.manager.SelectLine(y)
}

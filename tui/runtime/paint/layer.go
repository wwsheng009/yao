package paint

import "github.com/yaoapp/yao/tui/runtime/style"

// LayerType represents the type of layer
type LayerType int

const (
	// LayerBackground is for static UI (menus, borders)
	LayerBackground LayerType = iota
	// LayerContent is for forms, tables
	LayerContent
	// LayerStream is for high-frequency logs
	LayerStream
	// LayerOverlay is for modals, popups
	LayerOverlay
)

// String returns the string representation of LayerType
func (l LayerType) String() string {
	switch l {
	case LayerBackground:
		return "Background"
	case LayerContent:
		return "Content"
	case LayerStream:
		return "Stream"
	case LayerOverlay:
		return "Overlay"
	default:
		return "Unknown"
	}
}

// Layer represents an independent rendering layer
type Layer struct {
	ID       string
	Type     LayerType
	ZIndex   int
	Buffer   *Buffer
	Dirty    bool
	Rect     Rect
	Enabled  bool
	Visible  bool
}

// NewLayer creates a new layer
func NewLayer(id string, layerType LayerType, zIndex int, width, height int) *Layer {
	return &Layer{
		ID:      id,
		Type:    layerType,
		ZIndex:  zIndex,
		Buffer:  NewBuffer(width, height),
		Dirty:   true,
		Enabled: true,
		Visible: true,
		Rect: Rect{
			X:      0,
			Y:      0,
			Width:  width,
			Height: height,
		},
	}
}

// NewLayerWithRect creates a new layer with a specific rectangle
func NewLayerWithRect(id string, layerType LayerType, zIndex int, rect Rect) *Layer {
	return &Layer{
		ID:      id,
		Type:    layerType,
		ZIndex:  zIndex,
		Buffer:  NewBuffer(rect.Width, rect.Height),
		Dirty:   true,
		Enabled: true,
		Visible: true,
		Rect:    rect,
	}
}

// MarkDirty marks the layer as dirty
func (l *Layer) MarkDirty() {
	l.Dirty = true
}

// ClearDirty clears the dirty flag
func (l *Layer) ClearDirty() {
	l.Dirty = false
}

// IsDirty returns true if the layer needs rendering
func (l *Layer) IsDirty() bool {
	return l.Enabled && l.Visible && l.Dirty
}

// SetRect sets the layer's rectangle
func (l *Layer) SetRect(rect Rect) {
	l.Rect = rect
	// Resize buffer if needed
	if rect.Width != l.Buffer.Width || rect.Height != l.Buffer.Height {
		l.Buffer = NewBuffer(rect.Width, rect.Height)
		l.MarkDirty()
	}
}

// GetRect returns the layer's rectangle
func (l *Layer) GetRect() Rect {
	return l.Rect
}

// Enable enables the layer
func (l *Layer) Enable() {
	l.Enabled = true
	l.MarkDirty()
}

// Disable disables the layer
func (l *Layer) Disable() {
	l.Enabled = false
}

// Show shows the layer
func (l *Layer) Show() {
	l.Visible = true
	l.MarkDirty()
}

// Hide hides the layer
func (l *Layer) Hide() {
	l.Visible = false
}

// SetPosition sets the layer's position
func (l *Layer) SetPosition(x, y int) {
	l.Rect.X = x
	l.Rect.Y = y
	l.MarkDirty()
}

// SetSize sets the layer's size
func (l *Layer) SetSize(width, height int) {
	l.Rect.Width = width
	l.Rect.Height = height
	if width != l.Buffer.Width || height != l.Buffer.Height {
		l.Buffer = NewBuffer(width, height)
	}
	l.MarkDirty()
}

// Clear clears the layer's buffer
func (l *Layer) Clear() {
	for y := 0; y < l.Buffer.Height; y++ {
		for x := 0; x < l.Buffer.Width; x++ {
			l.Buffer.Cells[y][x] = Cell{}
		}
	}
	l.MarkDirty()
}

// Fill fills the layer with a character and style
func (l *Layer) Fill(char rune, st style.Style) {
	l.Buffer.Fill(l.Rect, char, st)
	l.MarkDirty()
}

package paint

import (
	"bytes"
	"sort"

	"github.com/yaoapp/yao/tui/runtime/style"
)

// Compositor manages multiple layers and composites them
type Compositor struct {
	layers      []*Layer
	width       int
	height      int
	batch       *CommandBatch
	styleState  *StyleStateMachine
}

// NewCompositor creates a new compositor
func NewCompositor(width, height int) *Compositor {
	return &Compositor{
		layers:     make([]*Layer, 0, 4),
		width:      width,
		height:     height,
		batch:      NewCommandBatch(),
		styleState: NewStyleStateMachine(),
	}
}

// AddLayer adds a layer to the compositor
func (c *Compositor) AddLayer(layer *Layer) {
	c.layers = append(c.layers, layer)
	c.sortLayers()
}

// RemoveLayer removes a layer by ID
func (c *Compositor) RemoveLayer(id string) bool {
	for i, layer := range c.layers {
		if layer.ID == id {
			c.layers = append(c.layers[:i], c.layers[i+1:]...)
			return true
		}
	}
	return false
}

// GetLayer returns a layer by ID
func (c *Compositor) GetLayer(id string) *Layer {
	for _, layer := range c.layers {
		if layer.ID == id {
			return layer
		}
	}
	return nil
}

// GetLayerByType returns the first layer of the given type
func (c *Compositor) GetLayerByType(layerType LayerType) *Layer {
	for _, layer := range c.layers {
		if layer.Type == layerType {
			return layer
		}
	}
	return nil
}

// sortLayers sorts layers by Z-index
func (c *Compositor) sortLayers() {
	sort.Slice(c.layers, func(i, j int) bool {
		return c.layers[i].ZIndex < c.layers[j].ZIndex
	})
}

// RenderDirty renders only dirty layers and returns the composited output
func (c *Compositor) RenderDirty() string {
	var buf bytes.Buffer

	for _, layer := range c.layers {
		if !layer.IsDirty() {
			continue
		}

		// Output layer content
		buf.WriteString(c.renderLayer(layer))

		layer.ClearDirty()
	}

	return buf.String()
}

// renderLayer renders a single layer with region scrolling
func (c *Compositor) renderLayer(layer *Layer) string {
	var output string

	// For stream layers, use scroll optimization
	if layer.Type == LayerStream {
		// Set scroll region
		output = "\x1b[" + itoa(layer.Rect.Y+1) + ";" +
			itoa(layer.Rect.Y+layer.Rect.Height) + "r"
	}

	// Output layer buffer content
	output += c.bufferToString(layer)

	// Reset scroll region
	if layer.Type == LayerStream {
		output += "\x1b[r"
	}

	return output
}

// bufferToString converts a layer's buffer to ANSI output
func (c *Compositor) bufferToString(layer *Layer) string {
	c.batch.Clear()
	c.styleState.Reset()

	buf := layer.Buffer
	for y := 0; y < buf.Height; y++ {
		for x := 0; x < buf.Width; x++ {
			cell := buf.Cells[y][x]
			if cell.Char != 0 {
				c.batch.AddCell(
					layer.Rect.X+x,
					layer.Rect.Y+y,
					cell.Char,
					cell.Style,
				)
			}
		}
	}

	return c.batch.Flush()
}

// Composite creates a composite buffer from all layers
func (c *Compositor) Composite() *Buffer {
	buffer := NewBuffer(c.width, c.height)

	for _, layer := range c.layers {
		if !layer.Enabled || !layer.Visible {
			continue
		}

		c.blitLayer(buffer, layer)
	}

	return buffer
}

// blitLayer blits a layer onto the composite buffer
func (c *Compositor) blitLayer(dst *Buffer, src *Layer) {
	for y := 0; y < src.Rect.Height; y++ {
		for x := 0; x < src.Rect.Width; x++ {
			srcX := src.Rect.X + x
			srcY := src.Rect.Y + y

			if srcX >= dst.Width || srcY >= dst.Height {
				continue
			}

			if x >= src.Buffer.Width || y >= src.Buffer.Height {
				continue
			}

			cell := src.Buffer.Cells[y][x]
			if cell.Char != 0 {
				dst.Cells[srcY][srcX] = cell
			}
		}
	}
}

// MarkAllDirty marks all layers as dirty
func (c *Compositor) MarkAllDirty() {
	for _, layer := range c.layers {
		layer.MarkDirty()
	}
}

// MarkTypeDirty marks all layers of a given type as dirty
func (c *Compositor) MarkTypeDirty(layerType LayerType) {
	for _, layer := range c.layers {
		if layer.Type == layerType {
			layer.MarkDirty()
		}
	}
}

// Resize handles window resize
func (c *Compositor) Resize(width, height int) {
	c.width = width
	c.height = height

	for _, layer := range c.layers {
		if layer.Rect.Width > width {
			layer.Rect.Width = width
		}
		if layer.Rect.Height > height {
			layer.Rect.Height = height
		}
		layer.SetRect(layer.Rect)
	}
}

// GetLayerCount returns the number of layers
func (c *Compositor) GetLayerCount() int {
	return len(c.layers)
}

// GetLayers returns all layers
func (c *Compositor) GetLayers() []*Layer {
	return c.layers
}

// Clear clears all layers
func (c *Compositor) Clear() {
	for _, layer := range c.layers {
		layer.Clear()
	}
}

// ClearType clears all layers of a given type
func (c *Compositor) ClearType(layerType LayerType) {
	for _, layer := range c.layers {
		if layer.Type == layerType {
			layer.Clear()
		}
	}
}

// Fill fills a layer with a character and style
func (c *Compositor) Fill(id string, char rune, st style.Style) bool {
	layer := c.GetLayer(id)
	if layer == nil {
		return false
	}
	layer.Fill(char, st)
	return true
}

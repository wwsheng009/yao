package paint

import (
	"bytes"
	"sort"

	"github.com/yaoapp/yao/tui/runtime/style"
)

// DrawCmd represents a single drawing command
type DrawCmd struct {
	X, Y  int
	Text  string
	Style style.Style
}

// CommandBatch batches draw commands to minimize terminal IO
type CommandBatch struct {
	cmds    []DrawCmd
	styleVM *StyleStateMachine
}

// NewCommandBatch creates a new command batch
func NewCommandBatch() *CommandBatch {
	return &CommandBatch{
		cmds:    make([]DrawCmd, 0, 256),
		styleVM: NewStyleStateMachine(),
	}
}

// Add adds a draw command
func (b *CommandBatch) Add(x, y int, text string, st style.Style) {
	b.cmds = append(b.cmds, DrawCmd{
		X:     x,
		Y:     y,
		Text:  text,
		Style: st,
	})
}

// AddCell adds a single cell command
func (b *CommandBatch) AddCell(x, y int, char rune, st style.Style) {
	b.Add(x, y, string(char), st)
}

// Flush merges commands and generates the final output
func (b *CommandBatch) Flush() string {
	if len(b.cmds) == 0 {
		return ""
	}

	var buf bytes.Buffer
	b.styleVM.Reset()

	// Sort by Y then X for linear traversal
	b.sortCommands()

	// Merge adjacent commands with same style
	merged := b.mergeCommands()

	// Generate output with style state machine
	lastX, lastY := -1, -1
	for _, cmd := range merged {
		// Move cursor if needed
		if cmd.X != lastX || cmd.Y != lastY {
			buf.WriteString(b.moveCursor(cmd.X, cmd.Y))
			lastX, lastY = cmd.X, cmd.Y
		}

		// Apply style if changed
		if b.styleVM.NeedsUpdate(cmd.Style) {
			buf.WriteString(b.styleVM.Update(cmd.Style))
		}

		// Write text
		buf.WriteString(cmd.Text)
		lastX += len(cmd.Text)
	}

	// Reset style at end
	buf.WriteString("\x1b[0m")

	return buf.String()
}

// mergeCommands merges adjacent commands that can be combined
func (b *CommandBatch) mergeCommands() []DrawCmd {
	if len(b.cmds) == 0 {
		return nil
	}

	merged := make([]DrawCmd, 0, len(b.cmds))
	current := b.cmds[0]

	for i := 1; i < len(b.cmds); i++ {
		next := b.cmds[i]

		// Check if we can merge
		if b.canMerge(current, next) {
			current.Text += next.Text
		} else {
			merged = append(merged, current)
			current = next
		}
	}

	merged = append(merged, current)
	return merged
}

// canMerge checks if two commands can be merged
func (b *CommandBatch) canMerge(a, c DrawCmd) bool {
	// Must be on same line
	if a.Y != c.Y {
		return false
	}

	// Must be adjacent
	if a.X+len(a.Text) != c.X {
		return false
	}

	// Must have same style
	return a.Style == c.Style
}

// sortCommands sorts commands by Y then X
func (b *CommandBatch) sortCommands() {
	sort.Slice(b.cmds, func(i, j int) bool {
		if b.cmds[i].Y != b.cmds[j].Y {
			return b.cmds[i].Y < b.cmds[j].Y
		}
		return b.cmds[i].X < b.cmds[j].X
	})
}

// moveCursor generates ANSI cursor movement
func (b *CommandBatch) moveCursor(x, y int) string {
	return "\x1b[" + itoa(y+1) + ";" + itoa(x+1) + "H"
}

// Clear clears all commands
func (b *CommandBatch) Clear() {
	b.cmds = b.cmds[:0]
}

// Count returns the number of commands
func (b *CommandBatch) Count() int {
	return len(b.cmds)
}

// Reserve reserves space for additional commands
func (b *CommandBatch) Reserve(n int) {
	if cap(b.cmds) < len(b.cmds)+n {
		newCmds := make([]DrawCmd, len(b.cmds), cap(b.cmds)+n+256)
		copy(newCmds, b.cmds)
		b.cmds = newCmds
	}
}

// MergeFrom merges another batch into this one
func (b *CommandBatch) MergeFrom(other *CommandBatch) {
	if other == nil || len(other.cmds) == 0 {
		return
	}
	b.Reserve(len(other.cmds))
	b.cmds = append(b.cmds, other.cmds...)
}

// EstimateSize estimates the size of the flushed output
func (b *CommandBatch) EstimateSize() int {
	if len(b.cmds) == 0 {
		return 0
	}

	// Rough estimate: each character + cursor moves + style codes
	size := 0
	for _, cmd := range b.cmds {
		size += len(cmd.Text) + 20 // Approx for cursor and style
	}
	return size
}

// WriteToBuffer writes all commands to a paint Buffer
func (b *CommandBatch) WriteToBuffer(buf *Buffer) {
	for _, cmd := range b.cmds {
		buf.SetString(cmd.X, cmd.Y, cmd.Text, cmd.Style)
	}
}

// WriteToBufferWithOffset writes all commands to a paint Buffer with offset
func (b *CommandBatch) WriteToBufferWithOffset(buf *Buffer, offsetX, offsetY int) {
	for _, cmd := range b.cmds {
		buf.SetString(cmd.X+offsetX, cmd.Y+offsetY, cmd.Text, cmd.Style)
	}
}

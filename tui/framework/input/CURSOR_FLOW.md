# Cursor Position Calculation and Rendering Flow

## Overview

This document describes how the TextInput component responds to user input and calculates cursor position for rendering.

## Architecture

```
TextInput (Parent Component)
    ├── value: string          // Current text value
    ├── cursor: int            // Logical cursor position (character index)
    └── cursorComp: *cursor.Cursor  // Independent cursor component
         ├── x, y: int         // Relative position in parent
         ├── visible: bool     // Current visibility (for blinking)
         └── style: style.Style
```

## Input Response Flow

### 1. User Input Handling

```
User types 'a'
    ↓
KeyEvent received (Key='a', Special=KeyUnknown)
    ↓
HandleEvent(ev Event) bool
    ↓
HandleAction(ActionInputChar with Payload='a')
    ↓
handleInputChar('a')
    ├── value = insert 'a' at cursor position
    └── cursor++ (increment cursor position)
    ↓
return true (state changed)
```

**Code:** `textinput.go:376-438`

### 2. Cursor Position Update

```go
// handleInputChar - lines 553-562
func (t *TextInput) handleInputChar(char rune) bool {
    runes := []rune(t.value)
    t.value = string(append(runes[:t.cursor], append([]rune{char}, runes[t.cursor:]...)...))
    t.cursor++  // Move cursor after the inserted character
    return true
}
```

Example:
- Initial: `value="test"`, `cursor=2` (between 'e' and 's')
- Input 'x': `value="texst"`, `cursor=3` (between 'x' and 's')

### 3. Render Phase

```
MarkDirty() called
    ↓
Framework schedules re-render
    ↓
Paint(ctx, buf) called
    ↓
Read current state (value, cursor, isFocused)
    ↓
Calculate positions and render
```

**Code:** `textinput.go:276-370`

### 4. Position Calculation

```go
// Paint method - lines 350-369
if isFocused {
    // cursorX is relative position within TextInput
    cursorX := 1 + cursorPos  // 1 for left bracket '['

    // cursorY is relative Y (always 0 for single-line TextInput)
    cursorY := y - ctx.Y

    // Clamp to content bounds
    if cursorX < 1 {
        cursorX = 1
    }
    if cursorX > 1+contentWidth-1 {
        cursorX = 1 + contentWidth - 1
    }

    // Update cursor component position
    t.cursorComp.SetPosition(cursorX, cursorY)

    // Cursor.Paint adds ctx.X/Y for absolute position
    t.cursorComp.Paint(ctx, buf)
}
```

### 5. Cursor Component Rendering

```go
// cursor.go:60-85
func (c *Cursor) Paint(ctx component.PaintContext, buf *paint.Buffer) {
    if !c.IsVisible() {
        return
    }

    c.mu.RLock()
    x := ctx.X + c.x  // Convert relative to absolute
    y := ctx.Y + c.y
    cursorStyle := c.style
    shape := c.shape
    c.mu.RUnlock()

    // Boundary check
    if x < 0 || x >= buf.Width || y < 0 || y >= buf.Height {
        return
    }

    switch shape {
    case ShapeBlock:
        cell := buf.Cells[y][x]
        reverseStyle := cursorStyle.Reverse(true)
        buf.SetCell(x, y, cell.Char, reverseStyle)
    // ... other shapes
    }
}
```

## Example: Typing "hello"

| Step | Value | Cursor (logical) | Render | Screen X |
|------|-------|------------------|--------|----------|
| Init | ""    | 0                | `[]`    | 1 (on `[`) |
| Type 'h' | "h" | 1                | `[h]`   | 2 (on `h`) |
| Type 'e' | "he" | 2                | `[he]`  | 3 (on `e`) |
| Type 'l' | "hel" | 3               | `[hel]` | 4 (on `l`) |
| Type 'l' | "hell" | 4              | `[hell]`| 5 (on `l`) |
| Type 'o' | "hello" | 5            | `[hello]`| 6 (on `o`) |

## Key Design Principles

1. **Separation of Concerns**: Cursor is an independent component with its own state
2. **Relative Positioning**: Cursor position is relative to parent, converted to absolute during Paint
3. **Atomic Operations**: Input handling locks mutex, ensures consistent state
4. **Self-Contained Blinking**: Cursor manages its own blink timing without external Tick
5. **Thread Safety**: All state access is protected by mutex

## Files

- `tui/framework/cursor/cursor.go` - Independent cursor component
- `tui/framework/input/textinput.go` - TextInput with cursor integration
- `tui/framework/input/textinput_test.go` - Unit tests
- `tui/framework/input/rendering_test.go` - Rendering tests
- `tui/framework/cursor/blink_test.go` - Blink behavior tests

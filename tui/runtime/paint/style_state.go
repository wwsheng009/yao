package paint

import (
	"strings"

	"github.com/yaoapp/yao/tui/runtime/style"
)

// StyleStateMachine minimizes VT code changes by tracking current style
type StyleStateMachine struct {
	current style.Style
}

// NewStyleStateMachine creates a new style state machine
func NewStyleStateMachine() *StyleStateMachine {
	return &StyleStateMachine{
		current: style.Style{},
	}
}

// Reset resets the current style
func (s *StyleStateMachine) Reset() {
	s.current = style.Style{}
}

// NeedsUpdate checks if the style needs to be updated
func (s *StyleStateMachine) NeedsUpdate(st style.Style) bool {
	return s.current != st
}

// Update generates VT codes to transition to the new style
func (s *StyleStateMachine) Update(st style.Style) string {
	codes := s.buildDiffCodes(s.current, st)
	s.current = st
	return codes
}

// buildDiffCodes builds only the codes that need to change
func (s *StyleStateMachine) buildDiffCodes(from, to style.Style) string {
	if from == to {
		return ""
	}

	// If completely different or target is empty, emit reset first
	if from == (style.Style{}) {
		return s.fullStyle(to)
	}

	// If target is empty, emit reset
	if to == (style.Style{}) {
		return "\x1b[0m"
	}

	var codes []string

	// Check for reset needed - if many things changed, just reset and start fresh
	changes := 0
	if from.FG != to.FG && to.FG != "" {
		changes++
	}
	if from.BG != to.BG && to.BG != "" {
		changes++
	}
	if from.IsBold() != to.IsBold() && to.IsBold() {
		changes++
	}
	if from.IsItalic() != to.IsItalic() && to.IsItalic() {
		changes++
	}
	if from.IsUnderline() != to.IsUnderline() && to.IsUnderline() {
		changes++
	}
	if from.IsReverse() != to.IsReverse() && to.IsReverse() {
		changes++
	}

	// If many changes, reset and rebuild
	if changes >= 4 {
		return "\x1b[0m" + s.fullStyle(to)
	}

	// Otherwise, emit only changes
	// Bold
	if to.IsBold() && !from.IsBold() {
		codes = append(codes, "1")
	}

	// Italic
	if to.IsItalic() && !from.IsItalic() {
		codes = append(codes, "3")
	}

	// Underline
	if to.IsUnderline() && !from.IsUnderline() {
		codes = append(codes, "4")
	}

	// Reverse
	if to.IsReverse() && !from.IsReverse() {
		codes = append(codes, "7")
	}

	// Colors
	if from.FG != to.FG && to.FG != "" {
		codes = append(codes, colorCode(to.FG, false))
	}
	if from.BG != to.BG && to.BG != "" {
		codes = append(codes, colorCode(to.BG, true))
	}

	if len(codes) == 0 {
		return ""
	}

	return "\x1b[" + strings.Join(codes, ";") + "m"
}

// fullStyle generates full style codes
func (s *StyleStateMachine) fullStyle(st style.Style) string {
	var codes []string

	if st.IsBold() {
		codes = append(codes, "1")
	}
	if st.IsItalic() {
		codes = append(codes, "3")
	}
	if st.IsUnderline() {
		codes = append(codes, "4")
	}
	if st.IsReverse() {
		codes = append(codes, "7")
	}
	if st.FG != "" {
		codes = append(codes, colorCode(st.FG, false))
	}
	if st.BG != "" {
		codes = append(codes, colorCode(st.BG, true))
	}

	if len(codes) == 0 {
		return ""
	}

	return "\x1b[" + strings.Join(codes, ";") + "m"
}

// colorCode converts a color to ANSI code
func colorCode(color style.Color, isBackground bool) string {
	c := string(color)
	// Handle hex colors
	if strings.HasPrefix(c, "#") {
		// For now, skip hex colors - would require extended ANSI codes
		return ""
	}

	// Handle RGB format
	if strings.HasPrefix(c, "rgb(") {
		return ""
	}

	// Standard colors
	code, ok := colorToAnsi[strings.ToLower(c)]
	if !ok {
		return ""
	}

	if isBackground {
		return itoa(code + 40)
	}
	return itoa(code + 30)
}

// colorToAnsi maps color names to ANSI color codes
var colorToAnsi = map[string]int{
	"black":         0,
	"red":           1,
	"green":         2,
	"yellow":        3,
	"blue":          4,
	"magenta":       5,
	"cyan":          6,
	"white":         7,
	"bright-black":   8,
	"bright-red":     9,
	"bright-green":  10,
	"bright-yellow": 11,
	"bright-blue":   12,
	"bright-magenta": 13,
	"bright-cyan":   14,
	"bright-white":  15,
}

// itoa converts int to string (faster than strconv for small numbers)
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var buf [20]byte
	var n int
	for i > 0 {
		buf[n] = byte('0' + i%10)
		i /= 10
		n++
	}
	// Reverse
	for i := 0; i < n/2; i++ {
		buf[i], buf[n-1-i] = buf[n-1-i], buf[i]
	}
	return string(buf[:n])
}

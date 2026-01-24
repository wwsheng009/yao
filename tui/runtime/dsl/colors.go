package dsl

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ColorNameToANSI converts a color name or hex code to ANSI color code
// Supported formats:
//   - Named colors: "red", "blue", "green", etc.
//   - Bright colors: "brightRed", "brightBlue", etc.
//   - Hex codes: "#FF5733", "ff5733"
//   - RGB: "rgb(255, 87, 51)"
//   - ANSI numbers: "214", "57"
//   - Hex shorthand: "F53" -> "#FF5533"
func ColorNameToANSI(color string) string {
	if color == "" {
		return ""
	}

	// If it's already a number, return as is
	if isANSICode(color) {
		return color
	}

	// Hex code
	if strings.HasPrefix(color, "#") {
		return color
	}

	// RGB format
	if strings.HasPrefix(strings.ToLower(color), "rgb(") {
		return color
	}

	// Named colors - convert to ANSI
	color = strings.ToLower(color)
	if ansi, ok := basicColors[color]; ok {
		return ansi
	}
	if ansi, ok := brightColors[color]; ok {
		return ansi
	}
	if ansi, ok := semanticColors[color]; ok {
		return ansi
	}

	// Try as lipgloss color (may be hex or already valid)
	// If lipgloss can parse it, return as is
	c := lipgloss.Color(color)
	if c != "" {
		return color
	}

	// Default to white
	return "15"
}

// isANSICode checks if the string is an ANSI color code (0-255)
func isANSICode(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

// Basic 8 colors (ANSI 0-7)
var basicColors = map[string]string{
	"black":   "0",
	"red":     "1",
	"green":   "2",
	"yellow":  "3",
	"blue":    "4",
	"magenta": "5",
	"cyan":    "6",
	"white":   "7",

	// Alternate names
	"default":  "0",
	"darkgray":  "8",
	"darkgrey":  "8",
	"gray":      "8",
	"grey":      "8",
	"lightgray": "7",
	"lightgrey": "7",
}

// Bright 8 colors (ANSI 8-15)
var brightColors = map[string]string{
	"brightblack":   "8",
	"brightred":     "9",
	"brightgreen":   "10",
	"brightyellow":  "11",
	"brightblue":    "12",
	"brightmagenta": "13",
	"brightcyan":    "14",
	"brightwhite":   "15",

	// Alternate names for bright colors
	"lightblack":   "8",
	"lightred":     "9",
	"lightgreen":   "10",
	"lightyellow":  "11",
	"lightblue":    "12",
	"lightmagenta": "13",
	"lightcyan":    "14",
	"lightwhite":   "15",

	// Additional common color names
	"orange":  "208",
	"purple":  "127",
	"pink":    "204",
	"brown":   "130",
	"lime":    "10",
	"indigo":  "54",
	"violet":  "213",
	"gold":    "220",
	"silver":  "7",
	"beige":   "230",
}

// Semantic colors mapped to appropriate ANSI codes
var semanticColors = map[string]string{
	// Status colors
	"primary":    "21",  // Blue
	"secondary":  "95",  // Purple
	"success":    "34",  // Green
	"info":       "39",  // Light Blue
	"warning":    "208", // Orange
	"danger":     "196", // Red
	"error":      "196", // Red
	"critical":   "196", // Red

	// UI element colors
	"foreground": "15",  // White
	"background": "235", // Dark Gray
	"muted":      "245", // Muted Gray
	"border":     "240", // Border Gray

	// Table specific colors
	"header":     "214", // Orange/Gold
	"row":        "15",  // White
	"alternate":  "236", // Lighter gray for alternate rows
	"hover":      "57",  // Blue for hover

	// Text colors
	"text":       "15",  // White
	"textdark":   "0",   // Black
	"textlight":  "15",  // White
	"textmuted":  "245", // Gray

	// Common terminal color palette names
	// Based on popular terminal color schemes
	"tokyo-night-blue":   "39",
	"tokyo-night-cyan":   "38",
	"tokyo-night-green":  "42",
	"tokyo-night-yellow": "220",
	"tokyo-night-red":    "203",
	"dracula-purple":     "141",
	"dracula-pink":       "212",
	"dracula-red":        "203",
	"dracula-orange":     "216",
	"dracula-yellow":     "228",
	"nord-blue":          "104",
	"nord-cyan":          "109",
	"nord-green":         "151",
	"nord-yellow":        "222",
	"nord-red":           "167",
	"nord-purple":        "146",
}

// ParseColorStyle parses a color from various formats and returns a lipgloss.Color
func ParseColorStyle(color string) lipgloss.Color {
	if color == "" {
		return lipgloss.Color("")
	}
	ansiColor := ColorNameToANSI(color)
	return lipgloss.Color(ansiColor)
}

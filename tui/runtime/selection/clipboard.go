package selection

import (
	"errors"
	"os/exec"
	"runtime"
	"strings"
)

// Clipboard provides clipboard operations for different platforms.
type Clipboard struct {
	// Platform-specific commands
	commands Commands
}

// Commands holds the platform-specific clipboard commands.
type Commands struct {
	Copy      string
	CopyArgs  []string
	Paste     string
	PasteArgs []string
}

// NewClipboard creates a new Clipboard instance for the current platform.
func NewClipboard() *Clipboard {
	return &Clipboard{
		commands: getPlatformCommands(),
	}
}

// getPlatformCommands returns clipboard commands based on the OS.
func getPlatformCommands() Commands {
	switch runtime.GOOS {
	case "windows":
		return Commands{
			Copy:     "cmd",
			CopyArgs: []string{"/c", "echo", "<text>", "|", "clip"},
			// Note: The Windows implementation uses a different approach
		}
	case "darwin":
		return Commands{
			Copy:     "pbcopy",
			CopyArgs: []string{},
			Paste:    "pbpaste",
			PasteArgs: []string{},
		}
	case "linux":
		// Try to detect which clipboard utility is available
		if commandExists("wl-copy") {
			// Wayland
			return Commands{
				Copy:     "wl-copy",
				CopyArgs: []string{},
				Paste:    "wl-paste",
				PasteArgs: []string{},
			}
		} else if commandExists("xclip") {
			// X11 with xclip
			return Commands{
				Copy:     "xclip",
				CopyArgs: []string{"-selection", "clipboard"},
				Paste:    "xclip",
				PasteArgs: []string{"-selection", "clipboard", "-o"},
			}
		} else if commandExists("xsel") {
			// X11 with xsel
			return Commands{
				Copy:     "xsel",
				CopyArgs: []string{"--clipboard", "--input"},
				Paste:    "xsel",
				PasteArgs: []string{"--clipboard", "--output"},
			}
		}
	}

	// Default: no clipboard support
	return Commands{}
}

// commandExists checks if a command exists in PATH.
func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// Copy copies text to the clipboard.
func (c *Clipboard) Copy(text string) error {
	if text == "" {
		return errors.New("cannot copy empty text")
	}

	switch runtime.GOOS {
	case "windows":
		return c.copyWindows(text)
	case "darwin":
		return c.copyCommand(text, c.commands.Copy, c.commands.CopyArgs)
	case "linux":
		if c.commands.Copy == "" {
			return errors.New("no clipboard utility found (install wl-copy, xclip, or xsel)")
		}
		return c.copyCommand(text, c.commands.Copy, c.commands.CopyArgs)
	default:
		return errors.New("clipboard not supported on this platform")
	}
}

// copyWindows handles clipboard copy on Windows using PowerShell.
func (c *Clipboard) copyWindows(text string) error {
	// Use PowerShell to set clipboard content
	// This is more reliable than the echo | clip method
	psCommand := "Set-Clipboard"
	if strings.Contains(text, "'") {
		// Use double quotes if text contains single quotes
		psCommand = `$input | Set-Clipboard`
	}

	cmd := exec.Command("powershell", "-NoProfile", "-Command", psCommand)
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

// copyCommand copies text using a generic command.
func (c *Clipboard) copyCommand(text, cmd string, args []string) error {
	if cmd == "" {
		return errors.New("no copy command configured")
	}

	cmdExec := exec.Command(cmd, args...)
	cmdExec.Stdin = strings.NewReader(text)
	return cmdExec.Run()
}

// Paste retrieves text from the clipboard.
func (c *Clipboard) Paste() (string, error) {
	switch runtime.GOOS {
	case "windows":
		return c.pasteWindows()
	case "darwin":
		return c.pasteCommand(c.commands.Paste, c.commands.PasteArgs)
	case "linux":
		if c.commands.Paste == "" {
			return "", errors.New("no clipboard utility found")
		}
		return c.pasteCommand(c.commands.Paste, c.commands.PasteArgs)
	default:
		return "", errors.New("clipboard not supported on this platform")
	}
}

// pasteWindows retrieves clipboard content on Windows.
func (c *Clipboard) pasteWindows() (string, error) {
	cmd := exec.Command("powershell", "-NoProfile", "-Command", "Get-Clipboard")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(string(output), "\r\n"), nil
}

// pasteCommand retrieves clipboard content using a generic command.
func (c *Clipboard) pasteCommand(cmd string, args []string) (string, error) {
	if cmd == "" {
		return "", errors.New("no paste command configured")
	}

	cmdExec := exec.Command(cmd, args...)
	output, err := cmdExec.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(string(output), "\n"), nil
}

// IsSupported returns whether clipboard operations are supported on this platform.
func (c *Clipboard) IsSupported() bool {
	switch runtime.GOOS {
	case "windows", "darwin":
		return true
	case "linux":
		return c.commands.Copy != ""
	default:
		return false
	}
}

// CopyWithFallback copies text to clipboard with a fallback for unsupported platforms.
// If clipboard is not supported, it returns an error but the text is still returned
// for potential alternative handling (like logging to a file).
func (c *Clipboard) CopyWithFallback(text string) error {
	if !c.IsSupported() {
		return &UnsupportedError{Platform: runtime.GOOS}
	}
	return c.Copy(text)
}

// UnsupportedError is returned when clipboard is not supported on a platform.
type UnsupportedError struct {
	Platform string
}

func (e *UnsupportedError) Error() string {
	return "clipboard not supported on " + e.Platform
}

// IsUnsupported returns whether an error indicates unsupported clipboard.
func IsUnsupported(err error) bool {
	_, ok := err.(*UnsupportedError)
	return ok
}

// Global clipboard instance for convenience.
var globalClipboard = NewClipboard()

// CopyToClipboard is a convenience function to copy text using the global clipboard.
func CopyToClipboard(text string) error {
	return globalClipboard.Copy(text)
}

// GetFromClipboard is a convenience function to paste text using the global clipboard.
func GetFromClipboard() (string, error) {
	return globalClipboard.Paste()
}

// IsClipboardSupported returns whether clipboard is supported on this platform.
func IsClipboardSupported() bool {
	return globalClipboard.IsSupported()
}

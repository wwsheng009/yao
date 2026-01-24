package runtime

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// Clipboard provides clipboard operations for different platforms.
type Clipboard struct {
	// Platform detection is done at runtime
}

// NewClipboard creates a new Clipboard instance.
func NewClipboard() *Clipboard {
	return &Clipboard{}
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
		return c.copyCommand(text, "pbcopy", []string{})
	case "linux":
		// Try to detect which clipboard utility is available
		if commandExists("wl-copy") {
			return c.copyCommand(text, "wl-copy", []string{})
		} else if commandExists("xclip") {
			return c.copyCommand(text, "xclip", []string{"-selection", "clipboard"})
		} else if commandExists("xsel") {
			return c.copyCommand(text, "xsel", []string{"--clipboard", "--input"})
		} else {
			return errors.New("no clipboard utility found (install wl-copy, xclip, or xsel)")
		}
	default:
		return errors.New("clipboard not supported on this platform")
	}
}

// copyWindows handles clipboard copy on Windows using PowerShell.
func (c *Clipboard) copyWindows(text string) error {
	// Use PowerShell Set-Clipboard with -Value parameter
	// This is more reliable than piping through stdin
	psScript := fmt.Sprintf("Set-Clipboard -Value %s", escapePowerShellString(text))
	cmd := exec.Command("powershell", "-NoProfile", "-Command", psScript)
	err := cmd.Run()
	if err == nil {
		return nil
	}

	// Fallback to using stdin with PowerShell
	cmd2 := exec.Command("powershell", "-NoProfile", "-Command", "Set-Clipboard")
	cmd2.Stdin = strings.NewReader(text)
	err = cmd2.Run()
	if err != nil {
		return fmt.Errorf("both PowerShell methods failed: %w", err)
	}
	return nil
}

// escapePowerShellString escapes a string for use in PowerShell command arguments
func escapePowerShellString(s string) string {
	// Replace single quotes with double single quotes for PowerShell escaping
	// In PowerShell: 'don''t' represents "don't"
	s = strings.ReplaceAll(s, "'", "''")
	return "'" + s + "'"
}

// copyCommand copies text using a generic command.
func (c *Clipboard) copyCommand(text, cmd string, args []string) error {
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
		return c.pasteCommand("pbpaste", []string{})
	case "linux":
		if commandExists("wl-paste") {
			return c.pasteCommand("wl-paste", []string{})
		} else if commandExists("xclip") {
			return c.pasteCommand("xclip", []string{"-selection", "clipboard", "-o"})
		} else if commandExists("xsel") {
			return c.pasteCommand("xsel", []string{"--clipboard", "--output"})
		} else {
			return "", errors.New("no clipboard utility found")
		}
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
		return commandExists("wl-copy") || commandExists("xclip") || commandExists("xsel")
	default:
		return false
	}
}

// commandExists checks if a command exists in PATH.
func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
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

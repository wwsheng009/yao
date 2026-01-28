package form

import (
	"fmt"
	stdtesting "testing"
	"time"

	"github.com/yaoapp/yao/tui/framework/component"
	"github.com/yaoapp/yao/tui/framework/cursor"
	"github.com/yaoapp/yao/tui/framework/event"
	"github.com/yaoapp/yao/tui/framework/input"
	"github.com/yaoapp/yao/tui/runtime/style"
	"github.com/yaoapp/yao/tui/framework/validation"
	"github.com/yaoapp/yao/tui/runtime/paint"
)

// TestRuntimeIntegration simulates the full app runtime flow
// This test closely mimics what happens in app.Run() to find issues
// that don't appear in simple unit tests.
func TestRuntimeIntegration(t *stdtesting.T) {
	fmt.Println("\n========== Runtime Integration Test ==========")

	// Step 1: Create form (simulates createLoginForm in interactive.go)
	loginForm := createLoginFormForRuntime()

	// Step 2: Create context with dirty callback (simulates app.NewApp())
	ctx := component.NewComponentContext()
	ctx.SetDirtyCallback(func() {
		fmt.Println("  [DirtyCallback triggered]")
	})

	// Call MountWithContext on the form
	loginForm.MountWithContext(nil, ctx)
	fmt.Println("  [Form mounted with context]")

	// Step 3: Init - call OnFocus on root (simulates app.Init())
	// This is what app does at line 222-225
	loginForm.OnFocus()
	fmt.Println("  [Root OnFocus called]")

	// Verify: Check if TextInput cursor is enabled
	fields := loginForm.GetFields()
	usernameInput, ok := fields["username"]
	if !ok {
		t.Fatal("username field not found")
	}
	txtInput, ok := usernameInput.Input.(*input.TextInput)
	if !ok {
		t.Fatal("username input is not TextInput")
	}

	cursorComp := txtInput.GetCursorComp()
	if cursorComp == nil {
		t.Fatal("Cursor component is nil")
	}

	// Check if cursor is enabled after focus
	if !cursorComp.IsBlinkEnabled() {
		t.Error("Cursor blink should be enabled after OnFocus!")
	}
	fmt.Printf("  [Cursor blink enabled: %v]\n", cursorComp.IsBlinkEnabled())

	// Step 4: Simulate typing (simulates event pump delivering key events)
	fmt.Println("\n[Simulating key input 'test']")

	for i, ch := range "test" {
		// Create KeyEvent (as event pump would)
		keyEv := event.NewKeyEvent(event.Key{
			Rune: ch,
		})
		keyEv.Special = event.KeyUnknown

		// Handle event (as app.handleEvent does)
		handled := loginForm.HandleEvent(keyEv)
		if !handled {
			t.Errorf("Character '%c' not handled", ch)
		}

		expectedValue := "test"[:i+1]
		actualValue := txtInput.GetValue()
		actualCursor := txtInput.GetCursor()

		fmt.Printf("  After '%c': value='%s', cursor=%d, focused=%v\n",
			ch, actualValue, actualCursor, txtInput.IsFocused())

		if actualValue != expectedValue {
			t.Errorf("Expected value '%s', got '%s'", expectedValue, actualValue)
		}
		if actualCursor != i+1 {
			t.Errorf("Expected cursor %d, got %d", i+1, actualCursor)
		}
	}

	// Step 5: Render (simulates app.render())
	fmt.Println("\n[Rendering form]")
	buf := paint.NewBuffer(80, 24)
	ctx2 := component.NewPaintContext(buf, 0, 0, 80, 24)

	loginForm.Paint(ctx2, buf)
	fmt.Println("  [Paint completed]")

	// Verify: Check if 'test' appears in render
	foundTest := false
	foundCursor := false

	for y := 0; y < buf.Height; y++ {
		for x := 0; x < buf.Width-3; x++ {
			// Look for "test" in the buffer
			if buf.Cells[y][x].Char == 't' &&
				buf.Cells[y][x+1].Char == 'e' &&
				buf.Cells[y][x+2].Char == 's' &&
				buf.Cells[y][x+3].Char == 't' {
				foundTest = true
				fmt.Printf("  [Found 'test' at row %d, col %d]\n", y, x)
			}
		}
		// Look for cursor (reverse style)
		for x := 0; x < buf.Width; x++ {
			if buf.Cells[y][x].Style.IsReverse() {
				foundCursor = true
				fmt.Printf("  [Found cursor at row %d, col %d]\n", y, x)

				// Show what character is at cursor
				if buf.Cells[y][x].Char != 0 {
					fmt.Printf("  [Cursor over character: '%c']\n", buf.Cells[y][x].Char)
				}
			}
		}
	}

	if !foundTest {
		t.Error("The word 'test' not found in render output")
	}
	if !foundCursor {
		t.Error("Cursor (reverse style) not found in render output")
	}

	// Step 6: Simulate cursor blink (cursor handles its own timing)
	fmt.Println("\n[Checking cursor self-blink]")

	// Cursor should be visible initially
	if !cursorComp.IsVisible() {
		t.Error("Cursor should be visible initially")
	}
	fmt.Printf("  [Cursor visible: %v]\n", cursorComp.IsVisible())

	// Wait for blink interval
	cursorComp.SetBlinkInterval(100 * time.Millisecond)
	time.Sleep(150 * time.Millisecond)

	// Check visibility after blink
	visibleAfterBlink := cursorComp.IsVisible()
	fmt.Printf("  [Cursor visible after blink: %v]\n", visibleAfterBlink)

	// Step 7: Final state check
	fmt.Println("\n[Final State]")
	fmt.Printf("  Value: '%s'\n", txtInput.GetValue())
	fmt.Printf("  Cursor: %d\n", txtInput.GetCursor())
	fmt.Printf("  Focused: %v\n", txtInput.IsFocused())
	fmt.Printf("  Visible: %v\n", txtInput.IsVisible())

	fmt.Println("\n========== Runtime Integration Complete ==========")
}

// TestCursorBlink verifies cursor self-blink works correctly
func TestCursorBlink(t *stdtesting.T) {
	fmt.Println("\n========== Cursor Blink Test ==========")

	// Create a new cursor
	c := cursor.NewCursor()
	c.SetBlinkInterval(100 * time.Millisecond)

	// Initially visible
	if !c.IsVisible() {
		t.Error("Cursor should be visible initially")
	}
	fmt.Println("  [OK: Cursor visible initially]")

	// Wait for blink
	time.Sleep(150 * time.Millisecond)
	visibleAfter := c.IsVisible()
	fmt.Printf("  [Cursor visible after blink: %v]\n", visibleAfter)

	// Check multiple blinks
	changeCount := 0
	lastVisible := visibleAfter
	for i := 0; i < 5; i++ {
		time.Sleep(120 * time.Millisecond)
		nowVisible := c.IsVisible()
		if nowVisible != lastVisible {
			changeCount++
			lastVisible = nowVisible
		}
	}

	if changeCount == 0 {
		t.Error("Cursor did not blink")
	} else {
		fmt.Printf("  [OK: Cursor blinked %d times]\n", changeCount)
	}

	// Test disabling blink
	c.SetBlinkEnabled(false)
	if !c.IsVisible() {
		t.Error("Cursor should always be visible when blink is disabled")
	}
	fmt.Println("  [OK: Cursor visible when blink disabled]")

	fmt.Println("\n========== Cursor Blink Test Complete ==========")
}

// createLoginFormForRuntime creates a login form for runtime testing
func createLoginFormForRuntime() *Form {
	f := NewForm()
	f.SetID("login-form")
	f.SetLabelStyle(style.Style{}.Foreground(style.Cyan))

	// Username field
	usernameInput := input.NewTextInput()
	usernameInput.SetID("username-input")
	usernameInput.SetPlaceholder("Enter username")

	usernameField := NewFormField("username")
	usernameField.Label = "Username: *"
	usernameField.Input = usernameInput
	usernameField.HelpText = "At least 3 characters"
	usernameField.Validators = []validation.Validator{
		validation.NewFuncValidator(func(value interface{}) error {
			if value == nil || value == "" {
				return fmt.Errorf("username required")
			}
			return nil
		}, "required"),
	}
	f.AddField(usernameField)

	return f
}

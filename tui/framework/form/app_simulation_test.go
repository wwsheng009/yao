package form

import (
	"fmt"
	stdtesting "testing"

	"github.com/yaoapp/yao/tui/framework/component"
	"github.com/yaoapp/yao/tui/framework/event"
	"github.com/yaoapp/yao/tui/framework/input"
	"github.com/yaoapp/yao/tui/framework/validation"
	"github.com/yaoapp/yao/tui/runtime/paint"
)

// TestAppSimulation simulates the actual App.Run() flow
func TestAppSimulation(t *stdtesting.T) {
	// Create form
	loginForm := createTestLoginForm()

	fmt.Println("\n========== App Simulation Test ==========")

	// Test 1: Simulate app initialization flow
	fmt.Println("\n[Test 1] App initialization")

	// Check if form was mounted with context
	fields := loginForm.GetFields()
	fmt.Printf("  Fields count: %d\n", len(fields))

	for name, field := range fields {
		if txt, ok := field.Input.(*input.TextInput); ok {
			fmt.Printf("  %s: value='%s', cursor=%d, focused=%v\n",
				name, txt.GetValue(), txt.GetCursor(), txt.IsFocused())
		}
	}

	// Test 2: Simulate the event flow from app.Run()
	fmt.Println("\n[Test 2] Simulate app event flow")

	// In app.Run(), after Init(), the root component gets focus
	loginForm.OnFocus()

	// Check if username input also got focus
	usernameInput := getUsernameInputFromFields(loginForm)
	if usernameInput == nil {
		t.Fatal("username-input not found")
	}

	fmt.Printf("  Form.IsFocused() = %v\n", loginForm.IsFocused())
	fmt.Printf("  usernameInput.IsFocused() = %v\n", usernameInput.IsFocused())

	// Test 3: Simulate typing through event pump
	fmt.Println("\n[Test 3] Simulate typing 'test'")

	inputSequence := "test"
	for i, ch := range inputSequence {
		fmt.Printf("\n  Typing '%c' (character %d of %d):\n", ch, i+1, len(inputSequence))

		// Create KeyEvent (as event pump would)
		keyEv := event.NewKeyEvent(event.Key{
			Rune: ch,
		})
		keyEv.Special = event.KeyUnknown

		// App routes keypress to root component
		handled := loginForm.HandleEvent(keyEv)
		fmt.Printf("    HandleEvent returned: %v\n", handled)

		// Check state after each character
		value := usernameInput.GetValue()
		cursor := usernameInput.GetCursor()
		fmt.Printf("    State: value='%s', cursor=%d\n", value, cursor)

		expectedValue := inputSequence[:i+1]
		if value != expectedValue {
			t.Errorf("After typing '%c': expected value='%s', got '%s'", ch, expectedValue, value)
		}
		if cursor != i+1 {
			t.Errorf("After typing '%c': expected cursor=%d, got %d", ch, i+1, cursor)
		}
	}

	// Test 4: Render after typing
	fmt.Println("\n[Test 4] Render after typing")

	buf := paint.NewBuffer(80, 24)
	ctx := component.NewPaintContext(buf, 0, 0, 80, 24)

	loginForm.Paint(ctx, buf)
	fmt.Println("  Render complete")

	// Check if "test" appears in render
	foundTest := false
	for y := 0; y < buf.Height; y++ {
		for x := 0; x < buf.Width; x++ {
			cell := buf.Cells[y][x]
			if cell.Char == 't' && x < buf.Width-3 && buf.Cells[y][x+1].Char == 'e' && buf.Cells[y][x+2].Char == 's' && buf.Cells[y][x+3].Char == 't' {
				foundTest = true
				fmt.Printf("  Found 'test' at row %d, col %d\n", y, x)
				break
			}
		}
		if foundTest {
			break
		}
	}

	if !foundTest {
		t.Error("The word 'test' not found in render output")
	}

	// Test 5: Check cursor position in render
	fmt.Println("\n[Test 5] Check cursor position in render")

	// Calculate expected cursor position:
	// Form.Paint calls TextInput.Paint with X=2 (offset for form layout)
	// TextInput draws '[' at X, then content at X+1, X+2, ...
	// At cursor position, '_' is drawn as the cursor
	cursorVal := usernameInput.GetCursor()
	// Position in buffer: 2 (form offset) + 1 (after '[') + cursor position
	// Note: when cursor=4 and len("test")=4, cursor is after content, so '_' at position 2+1+4=7
	expectedCursorCol := 2 + 1 + cursorVal
	foundCursor := false
	actualCursorCol := -1

	for y := 0; y < buf.Height; y++ {
		for x := 0; x < buf.Width; x++ {
			if buf.Cells[y][x].Style.IsReverse() {
				actualCursorCol = x
				foundCursor = true
				fmt.Printf("  Cursor found at row %d, col %d\n", y, x)
				break
			}
		}
		if foundCursor {
			break
		}
	}

	if !foundCursor {
		t.Error("Cursor (reverse style) not found in render")
	} else if actualCursorCol == expectedCursorCol {
		fmt.Printf("  OK: Cursor at expected position %d\n", actualCursorCol)
	} else {
		t.Errorf("Cursor position wrong: expected %d, got %d", expectedCursorCol, actualCursorCol)
	}

	// Test 6: Test backspace
	fmt.Println("\n[Test 6] Test backspace")

	backspaceEv := event.NewSpecialKeyEvent(event.KeyBackspace)

	loginForm.HandleEvent(backspaceEv)

	valueAfterBackspace := usernameInput.GetValue()
	cursorAfterBackspace := usernameInput.GetCursor()
	fmt.Printf("  After backspace: value='%s', cursor=%d\n", valueAfterBackspace, cursorAfterBackspace)

	if valueAfterBackspace != "tes" {
		t.Errorf("Backspace failed: expected 'tes', got '%s'", valueAfterBackspace)
	}
	if cursorAfterBackspace != 3 {
		t.Errorf("Cursor after backspace wrong: expected 3, got %d", cursorAfterBackspace)
	}

	fmt.Println("\n========== App Simulation Complete ==========")
}

// TestDirtyCallback tests if dirty callback triggers re-render
func TestDirtyCallback(t *stdtesting.T) {
	fmt.Println("\n========== Dirty Callback Test ==========")

	// Create form
	f := NewForm()

	// Add a field
	textInput := input.NewTextInput()
	textInput.SetID("test-input")

	field := NewFormField("test")
	field.Label = "Test"
	field.Input = textInput
	f.AddField(field)

	// Set dirty callback with flag
	dirtyCalled := false
	f.SetDirtyCallback(func() {
		dirtyCalled = true
		fmt.Printf("  Dirty callback called\n")
	})

	// Trigger markDirty which should call the dirty callback
	f.markDirty()

	if dirtyCalled {
		fmt.Println("  OK: Dirty callback works")
	} else {
		t.Error("Dirty callback not triggered by markDirty")
	}

	fmt.Println("\n========== Dirty Callback Test Complete ==========")
}

// =============================================================================
// Helper functions
// =============================================================================

func createTestLoginForm() *Form {
	f := NewForm()
	f.SetID("login-form")

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

func getUsernameInputFromFields(f *Form) *input.TextInput {
	fields := f.GetFields()
	if field, ok := fields["username"]; ok {
		if txt, ok := field.Input.(*input.TextInput); ok {
			return txt
		}
	}
	return nil
}

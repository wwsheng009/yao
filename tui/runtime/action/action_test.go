package action

import (
	"testing"
)

func TestActionCreation(t *testing.T) {
	a := NewAction(ActionInputText).
		WithPayload("hello").
		WithTarget("input-1").
		WithSource("keyboard")

	if a.Type != ActionInputText {
		t.Errorf("expected type %s, got %s", ActionInputText, a.Type)
	}

	if a.Payload != "hello" {
		t.Errorf("expected payload 'hello', got '%v'", a.Payload)
	}

	if a.Target != "input-1" {
		t.Errorf("expected target 'input-1', got '%s'", a.Target)
	}

	if a.Source != "keyboard" {
		t.Errorf("expected source 'keyboard', got '%s'", a.Source)
	}
}

func TestActionString(t *testing.T) {
	tests := []struct {
		name     string
		action   *Action
		expected string
	}{
		{
			name:     "action with target",
			action:   NewAction(ActionSubmit).WithTarget("form-1"),
			expected: "submit{form-1}",
		},
		{
			name:     "action without target",
			action:   NewAction(ActionQuit),
			expected: "quit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.action.String()
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestActionClone(t *testing.T) {
	original := NewAction(ActionInputText).
		WithPayload("hello").
		WithTarget("input-1")

	clone := original.Clone()

	// 验证克隆值相等
	if clone.Type != original.Type {
		t.Errorf("clone type mismatch")
	}
	if clone.Payload != original.Payload {
		t.Errorf("clone payload mismatch")
	}
	if clone.Target != original.Target {
		t.Errorf("clone target mismatch")
	}

	// 修改原始不影响克隆
	original.Payload = "world"
	if clone.Payload != "hello" {
		t.Errorf("clone was affected by original modification")
	}
}

func TestActionTypes(t *testing.T) {
	// 验证导航 Actions
	tests := []struct {
		actionType ActionType
		expected   string
	}{
		{ActionNavigateUp, "navigate_up"},
		{ActionNavigateDown, "navigate_down"},
		{ActionNavigateLeft, "navigate_left"},
		{ActionNavigateRight, "navigate_right"},
		{ActionNavigateNext, "navigate_next"},
		{ActionNavigatePrev, "navigate_prev"},
		{ActionNavigateFirst, "navigate_first"},
		{ActionNavigateLast, "navigate_last"},
		{ActionNavigatePageUp, "navigate_page_up"},
		{ActionNavigatePageDown, "navigate_page_down"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if string(tt.actionType) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, tt.actionType)
			}
		})
	}
}

func TestTargetFunc(t *testing.T) {
	called := false
	var receivedAction *Action

	target := NewTargetFunc("test-1", func(a *Action) bool {
		called = true
		receivedAction = a
		return true
	})

	if target.ID() != "test-1" {
		t.Errorf("expected ID 'test-1', got '%s'", target.ID())
	}

	action := NewAction(ActionSubmit)
	result := target.HandleAction(action)

	if !called {
		t.Errorf("handler was not called")
	}

	if !result {
		t.Errorf("expected true result")
	}

	if receivedAction != action {
		t.Errorf("received action mismatch")
	}
}

func TestTargetChain(t *testing.T) {
	firstCalled := false
	secondCalled := false

	chain := NewTargetChain("chain",
		NewTargetFunc("first", func(a *Action) bool {
			firstCalled = true
			return false // 不处理，传递给下一个
		}),
		NewTargetFunc("second", func(a *Action) bool {
			secondCalled = true
			return true // 处理
		}),
	)

	action := NewAction(ActionSubmit)
	result := chain.HandleAction(action)

	if !firstCalled {
		t.Errorf("first target was not called")
	}

	if !secondCalled {
		t.Errorf("second target was not called")
	}

	if !result {
		t.Errorf("expected true result")
	}
}

func TestNoOpTarget(t *testing.T) {
	target := NewNoOpTarget("noop")

	if target.ID() != "noop" {
		t.Errorf("expected ID 'noop', got '%s'", target.ID())
	}

	action := NewAction(ActionSubmit)
	result := target.HandleAction(action)

	if result {
		t.Errorf("NoOpTarget should return false")
	}
}

func TestErrorCreation(t *testing.T) {
	action := NewAction(ActionSubmit).WithTarget("form-1")

	err := NewError(ErrTargetDisabled, "component is disabled", action).
		WithTarget("submit-btn").
		WithComponentType("Button").
		WithDetail("disabled_reason", "form incomplete")

	if err.Type != ErrTargetDisabled {
		t.Errorf("expected error type %s, got %s", ErrTargetDisabled, err.Type)
	}

	if err.Target != "submit-btn" {
		t.Errorf("expected target 'submit-btn', got '%s'", err.Target)
	}

	if err.ComponentType != "Button" {
		t.Errorf("expected component type 'Button', got '%s'", err.ComponentType)
	}

	if val := err.Details["disabled_reason"]; val != "form incomplete" {
		t.Errorf("expected detail 'form incomplete', got '%v'", val)
	}
}

func TestErrorHelpers(t *testing.T) {
	t.Run("ErrTargetNotFound", func(t *testing.T) {
		act := NewAction(ActionSubmit)
		err := NewErrTargetNotFound("missing-id", act)

		if err.Type != ErrTargetNotFound {
			t.Errorf("expected error type %s, got %s", ErrTargetNotFound, err.Type)
		}
	})

	t.Run("ErrInvalidPayload", func(t *testing.T) {
		act := NewAction(ActionInputText).WithPayload(123)
		err := NewErrInvalidPayload("string", act)

		if err.Type != ErrInvalidPayload {
			t.Errorf("expected error type %s, got %s", ErrInvalidPayload, err.Type)
		}
	})

	t.Run("ErrActionNotSupported", func(t *testing.T) {
		act := NewAction(ActionDeleteChar)
		err := NewErrActionNotSupported("Label", "delete_char", act)

		if err.Type != ErrActionNotSupported {
			t.Errorf("expected error type %s, got %s", ErrActionNotSupported, err.Type)
		}

		if err.ComponentType != "Label" {
			t.Errorf("expected component type 'Label', got '%s'", err.ComponentType)
		}
	})
}

func TestPayloadValidation(t *testing.T) {
	t.Run("ValidateStringPayload", func(t *testing.T) {
		validAction := NewAction(ActionInputText).WithPayload("hello")
		s, ok := ValidateStringPayload(validAction)
		if !ok || s != "hello" {
			t.Errorf("failed to validate string payload")
		}

		invalidAction := NewAction(ActionInputText).WithPayload(123)
		_, ok = ValidateStringPayload(invalidAction)
		if ok {
			t.Errorf("should not validate int as string")
		}
	})

	t.Run("ValidateRunePayload", func(t *testing.T) {
		validAction := NewAction(ActionInputChar).WithPayload('a')
		r, ok := ValidateRunePayload(validAction)
		if !ok || r != 'a' {
			t.Errorf("failed to validate rune payload")
		}

		invalidAction := NewAction(ActionInputChar).WithPayload("abc")
		_, ok = ValidateRunePayload(invalidAction)
		if ok {
			t.Errorf("should not validate string as rune")
		}
	})

	t.Run("ValidateIntPayload", func(t *testing.T) {
		validAction := NewAction(ActionNavigateDown).WithPayload(1)
		i, ok := ValidateIntPayload(validAction)
		if !ok || i != 1 {
			t.Errorf("failed to validate int payload")
		}

		invalidAction := NewAction(ActionNavigateDown).WithPayload("1")
		_, ok = ValidateIntPayload(invalidAction)
		if ok {
			t.Errorf("should not validate string as int")
		}
	})
}

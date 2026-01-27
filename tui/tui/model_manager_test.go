package tui

import (
	"os"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yaoapp/gou/process"
)

// MockModel for testing
type MockModel struct{}

func (m MockModel) Init() tea.Cmd { return nil }
func (m MockModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) { 
	// Handle quit message specially in tests
	if _, ok := msg.(tea.QuitMsg); ok {
		return m, tea.Quit
	}
	return m, nil 
}
func (m MockModel) View() string { return "mock view" }

func TestModelManager(t *testing.T) {
	// Create a simple config for testing
	config := &Config{
		ID:   "test-model",
		Name: "Test Model",
		Data: map[string]interface{}{
			"test": "value",
		},
	}

	// Create a mock program for testing
	mockProgram := tea.NewProgram(MockModel{})
	
	// Create a model (this should register it automatically)
	_ = NewModel(config, mockProgram)
	
	// Verify the model was registered
	registeredModel := GetModel("test-model")
	if registeredModel == nil {
		t.Fatalf("Expected model to be registered, but got nil")
	}
	
	// Verify we can access the data
	if val, ok := registeredModel.State["test"]; !ok || val != "value" {
		t.Errorf("Expected state to contain test=value, got %v", val)
	}
	
	// Test getting a non-existent model
	nilModel := GetModel("non-existent")
	if nilModel != nil {
		t.Errorf("Expected nil for non-existent model, got %v", nilModel)
	}
	
	// Test listing model IDs
	ids := ListModelIDs()
	found := false
	for _, id := range ids {
		if id == "test-model" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected to find test-model in model IDs list, got %v", ids)
	}
	
	// Clean up
	UnregisterModel("test-model")
	
	// Verify cleanup
	cleanedModel := GetModel("test-model")
	if cleanedModel != nil {
		t.Errorf("Expected model to be unregistered, but still got %v", cleanedModel)
	}
}

// TestProcessFunctions tests the process functions without causing actual quit in tests
func TestProcessWithModelManager(t *testing.T) {
	// Create a simple config for testing
	config := &Config{
		ID:   "test-process-model",
		Name: "Test Process Model",
		Data: map[string]interface{}{
			"test": "initial",
		},
	}
	os.Setenv("TESTING_TUI", "true")
	defer os.Unsetenv("TESTING_TUI")
	// Create a mock program for testing
	mockProgram := tea.NewProgram(MockModel{})
	
	// Create a model (this should register it automatically)
	_ = NewModel(config, mockProgram)
	
	// Test ProcessQuit with the registered model ID
	proc := process.New("tui.quit", "test-process-model")
	result := ProcessQuit(proc)
	
	// The result should not be an error since the model exists
	if resultMap, ok := result.(map[string]interface{}); ok {
		if _, hasError := resultMap["error"]; hasError {
			t.Errorf("Expected no error, but got: %v", resultMap)
		}
	}
	
	// Test ProcessQuit with a non-existent model
	proc2 := process.New("tui.quit", "non-existent-model")
	result2 := ProcessQuit(proc2)
	
	// The result should be an error since the model doesn't exist
	if resultMap, ok := result2.(map[string]interface{}); ok {
		if _, hasError := resultMap["error"]; !hasError {
			t.Errorf("Expected an error for non-existent model, but didn't get one")
		}
	}
	
	// Clean up
	UnregisterModel("test-process-model")
}

// TestModelRegistrationOnCreation tests that models are automatically registered when created
func TestModelRegistrationOnCreation(t *testing.T) {
	config := &Config{
		ID:   "auto-register-test",
		Name: "Auto Register Test",
		Data: map[string]interface{}{
			"data": "test-value",
		},
	}
	
	mockProgram := tea.NewProgram(MockModel{})
	model := NewModel(config, mockProgram)
	
	// Check that the model was automatically registered
	registeredModel := GetModel("auto-register-test")
	if registeredModel == nil {
		t.Fatalf("Model should have been automatically registered on creation")
	}
	
	// Check that it's the same model instance
	if registeredModel != model {
		t.Errorf("Registered model should be the same instance as created model")
	}
	
	// Clean up
	UnregisterModel("auto-register-test")
}
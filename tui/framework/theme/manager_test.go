package theme

import (
	"testing"
)

func TestNewManager(t *testing.T) {
	mgr := NewManager()

	if mgr == nil {
		t.Fatal("NewManager() should not return nil")
	}

	if mgr.current != nil {
		t.Error("NewManager() should have no current theme initially")
	}

	if mgr.themes == nil {
		t.Error("NewManager() should initialize themes map")
	}

	if mgr.listeners == nil {
		t.Error("NewManager() should initialize listeners slice")
	}
}

func TestManagerRegister(t *testing.T) {
	mgr := NewManager()

	theme1 := NewTheme("theme1")
	theme2 := NewTheme("theme2")

	mgr.Register(theme1)
	mgr.Register(theme2)

	// First registered theme should become current
	if mgr.current != theme1 {
		t.Error("First registered theme should become current")
	}

	// Check themes are registered
	if _, ok := mgr.themes["theme1"]; !ok {
		t.Error("theme1 should be registered")
	}
	if _, ok := mgr.themes["theme2"]; !ok {
		t.Error("theme2 should be registered")
	}
}

func TestManagerGet(t *testing.T) {
	mgr := NewManager()

	theme1 := NewTheme("theme1")
	mgr.Register(theme1)

	// Test getting existing theme
	got, ok := mgr.Get("theme1")
	if !ok {
		t.Error("Get() should return true for existing theme")
	}
	if got != theme1 {
		t.Error("Get() should return the registered theme")
	}

	// Test getting non-existing theme
	_, ok = mgr.Get("nonexistent")
	if ok {
		t.Error("Get() should return false for non-existing theme")
	}
}

func TestManagerSet(t *testing.T) {
	mgr := NewManager()

	theme1 := NewTheme("theme1")
	theme2 := NewTheme("theme2")

	mgr.Register(theme1)
	mgr.Register(theme2)

	// Set to theme2
	err := mgr.Set("theme2")
	if err != nil {
		t.Errorf("Set() should not return error, got %v", err)
	}

	if mgr.current != theme2 {
		t.Error("Set() should change current theme")
	}

	// Try to set non-existing theme
	err = mgr.Set("nonexistent")
	if err == nil {
		t.Error("Set() should return error for non-existing theme")
	}
}

func TestManagerToggle(t *testing.T) {
	mgr := NewManager()

	theme1 := NewTheme("theme1")
	theme2 := NewTheme("theme2")
	theme3 := NewTheme("theme3")

	mgr.RegisterMultiple([]*Theme{theme1, theme2, theme3})

	// Start with theme1
	mgr.Set("theme1")

	// Toggle to theme2
	mgr.Toggle()
	if mgr.current.Name != "theme2" {
		t.Errorf("Toggle() should go to theme2, got %v", mgr.current.Name)
	}

	// Toggle to theme3
	mgr.Toggle()
	if mgr.current.Name != "theme3" {
		t.Errorf("Toggle() should go to theme3, got %v", mgr.current.Name)
	}

	// Toggle back to theme1 (wrap around)
	mgr.Toggle()
	if mgr.current.Name != "theme1" {
		t.Errorf("Toggle() should wrap to theme1, got %v", mgr.current.Name)
	}
}

func TestManagerList(t *testing.T) {
	mgr := NewManager()

	theme3 := NewTheme("theme3")
	theme1 := NewTheme("theme1")
	theme2 := NewTheme("theme2")

	mgr.RegisterMultiple([]*Theme{theme3, theme1, theme2})

	// List should return sorted names
	names := mgr.List()

	if len(names) != 3 {
		t.Errorf("List() should return 3 names, got %d", len(names))
	}

	// Check sorted
	if names[0] != "theme1" || names[1] != "theme2" || names[2] != "theme3" {
		t.Errorf("List() should return sorted names, got %v", names)
	}
}

func TestManagerHas(t *testing.T) {
	mgr := NewManager()

	theme1 := NewTheme("theme1")
	mgr.Register(theme1)

	if !mgr.Has("theme1") {
		t.Error("Has() should return true for existing theme")
	}

	if mgr.Has("nonexistent") {
		t.Error("Has() should return false for non-existing theme")
	}
}

func TestManagerCount(t *testing.T) {
	mgr := NewManager()

	if mgr.Count() != 0 {
		t.Errorf("Count() should be 0 initially, got %d", mgr.Count())
	}

	mgr.Register(NewTheme("theme1"))
	if mgr.Count() != 1 {
		t.Errorf("Count() should be 1, got %d", mgr.Count())
	}

	mgr.Register(NewTheme("theme2"))
	if mgr.Count() != 2 {
		t.Errorf("Count() should be 2, got %d", mgr.Count())
	}
}

func TestManagerSubscribe(t *testing.T) {
	mgr := NewManager()

	theme1 := NewTheme("theme1")
	theme2 := NewTheme("theme2")

	mgr.Register(theme1)
	mgr.Register(theme2)

	called := false
	listener := func(old, new *Theme) {
		called = true
		if old == theme1 && new == theme2 {
			// Correct
		} else {
			t.Errorf("Listener called with wrong themes: old=%v, new=%v", old, new)
		}
	}

	mgr.Subscribe(listener)

	// Trigger change
	mgr.Set("theme2")

	if !called {
		t.Error("Listener should be called when theme changes")
	}
}

func TestManagerUnsubscribe(t *testing.T) {
	mgr := NewManager()

	theme1 := NewTheme("theme1")
	theme2 := NewTheme("theme2")

	mgr.Register(theme1)
	mgr.Register(theme2)

	callCount := 0
	listener := func(old, new *Theme) {
		callCount++
	}

	unsubscribe := mgr.Subscribe(listener)

	// Unsubscribe
	unsubscribe()

	// Trigger change
	mgr.Set("theme2")

	if callCount != 0 {
		t.Errorf("Unsubscribed listener should not be called, was called %d times", callCount)
	}
}

func TestManagerGetColor(t *testing.T) {
	mgr := NewManager()

	theme1 := NewTheme("test")
	theme1.Colors.Primary = Blue

	mgr.Register(theme1)
	mgr.Set("test")

	color := mgr.GetColor("primary")
	if !color.Equals(Blue) {
		t.Errorf("GetColor() = %v, want %v", color, Blue)
	}
}

func TestManagerGetStyle(t *testing.T) {
	mgr := NewManager()

	theme1 := NewTheme("test")
	styleConfig := StyleConfig{
		Foreground: &Blue,
		Bold:       true,
	}
	theme1.SetStyle("test.style", styleConfig)

	mgr.Register(theme1)
	mgr.Set("test")

	style := mgr.GetStyle("", "test.style")
	if style.Foreground == nil || !style.Foreground.Equals(Blue) {
		t.Error("GetStyle() should return configured style")
	}
	if !style.Bold {
		t.Error("GetStyle() should return bold style")
	}
}

func TestManagerGetComponentStyle(t *testing.T) {
	mgr := NewManager()

	theme1 := NewTheme("test")
	baseStyle := StyleConfig{
		Foreground: &Blue,
	}
	focusedStyle := StyleConfig{
		Foreground: &Green,
	}

	theme1.SetComponentStyle("button", baseStyle, map[string]StyleConfig{
		"focused": focusedStyle,
	})

	mgr.Register(theme1)
	mgr.Set("test")

	// Test getting base style
	base := mgr.GetComponentStyle("button", "")
	if base.Foreground == nil || !base.Foreground.Equals(Blue) {
		t.Error("GetComponentStyle() should return base style")
	}

	// Test getting state style
	focused := mgr.GetComponentStyle("button", "focused")
	if focused.Foreground == nil || !focused.Foreground.Equals(Green) {
		t.Error("GetComponentStyle() should return focused style")
	}
}

func TestManagerRefresh(t *testing.T) {
	mgr := NewManager()

	theme1 := NewTheme("test")
	mgr.Register(theme1)

	called := false
	listener := func(old, new *Theme) {
		called = true
		if old == new {
			// Refresh should call with same theme
		} else {
			t.Error("Refresh should call with same theme")
		}
	}

	mgr.Subscribe(listener)
	mgr.Refresh()

	if !called {
		t.Error("Refresh() should trigger listeners")
	}
}

func TestManagerUnregister(t *testing.T) {
	mgr := NewManager()

	theme1 := NewTheme("theme1")
	theme2 := NewTheme("theme2")

	mgr.RegisterMultiple([]*Theme{theme1, theme2})
	mgr.Set("theme1")

	// Unregister theme2
	mgr.Unregister("theme2")

	if mgr.Has("theme2") {
		t.Error("Unregister() should remove theme")
	}

	// Current theme should still be theme1
	if mgr.current.Name != "theme1" {
		t.Error("Unregister() non-current theme should not change current")
	}

	// Unregister current theme
	mgr.Unregister("theme1")

	if mgr.Has("theme1") {
		t.Error("Unregister() should remove current theme")
	}

	// Current should change to remaining theme or nil
	if mgr.current != nil {
		// If there's another theme, it should become current
	}
}

func TestBuiltinThemes(t *testing.T) {
	themes := BuiltinThemes()

	if len(themes) < 4 {
		t.Errorf("BuiltinThemes() should return at least 4 themes, got %d", len(themes))
	}

	for _, theme := range themes {
		if theme == nil {
			t.Error("BuiltinThemes() should not contain nil themes")
		}
		if theme.Name == "" {
			t.Error("BuiltinThemes() themes should have names")
		}
	}
}

func TestGetBuiltinTheme(t *testing.T) {
	tests := []struct {
		name    string
		theme   string
		wantNil bool
	}{
		{"light theme", "light", false},
		{"dark theme", "dark", false},
		{"dracula theme", "dracula", false},
		{"nord theme", "nord", false},
		{"unknown theme", "unknown", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetBuiltinTheme(tt.theme)
			if (got == nil) != tt.wantNil {
				t.Errorf("GetBuiltinTheme() = %v, wantNil %v", got, tt.wantNil)
			}
			if got != nil && got.Name != tt.theme {
				t.Errorf("GetBuiltinTheme() Name = %v, want %v", got.Name, tt.theme)
			}
		})
	}
}

func TestBuiltinThemeNames(t *testing.T) {
	names := BuiltinThemeNames()

	if len(names) < 4 {
		t.Errorf("BuiltinThemeNames() should return at least 4 names, got %d", len(names))
	}

	// Check that all names are unique
	seen := make(map[string]bool)
	for _, name := range names {
		if seen[name] {
			t.Errorf("BuiltinThemeNames() contains duplicate: %s", name)
		}
		seen[name] = true
	}
}

func TestTransition(t *testing.T) {
	from := DarkTheme
	to := LightTheme

	transition := NewTransition(from, to, 300*1000000) // 300ms in nanoseconds

	if transition.from != from {
		t.Error("NewTransition() should set from theme")
	}
	if transition.to != to {
		t.Error("NewTransition() should set to theme")
	}

	// Test GetProgress
	progress := transition.GetProgress()
	if progress < 0 || progress > 1 {
		t.Errorf("GetProgress() should be between 0 and 1, got %f", progress)
	}

	// Test Update (without waiting)
	done, progress := transition.Update(0)
	if done {
		t.Error("Update() with 0 time should not be done")
	}
	if progress < 0 || progress > 1 {
		t.Errorf("Update() progress should be between 0 and 1, got %f", progress)
	}

	// Test InterpolateColor
	color := transition.InterpolateColor("primary")
	if color.IsNone() {
		t.Error("InterpolateColor() should not return none color")
	}
}

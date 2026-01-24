// Package framework provides a complete, independent TUI (Terminal User Interface)
// framework that does not depend on Bubble Tea. It reuses the existing tui/runtime
// layout engine and provides a rich set of components for building terminal applications.
//
// # Architecture
//
// The framework is organized into several layers:
//
//   - Application Layer (app.go): Main application lifecycle and event loop
//   - Component Layer (component/): Reusable UI components
//   - Event Layer (event/): Event handling and routing
//   - Screen Layer (screen/): Terminal screen management and rendering
//   - Style Layer (style/): Styling and theming
//   - Platform Layer (platform/): Platform-specific terminal operations
//
// # Quick Start
//
//	package main
//
//	import (
//	    "github.com/yaoapp/yao/tui/framework"
//	    "github.com/yaoapp/yao/tui/framework/display"
//	    "github.com/yaoapp/yao/tui/framework/event"
//	)
//
//	func main() {
//	    app := framework.NewApp()
//
//	    // Build UI
//	    root := display.NewText("Hello, World!")
//	    app.SetRoot(root)
//
//	    // Handle keyboard events
//	    app.OnKey(event.KeyEscape, app.Quit)
//	    app.OnKey('q', app.Quit)
//
//	    // Run
//	    if err := app.Run(); err != nil {
//	        panic(err)
//	    }
//	}
//
// # Components
//
// The framework provides the following components:
//
//   - Display: Text, Paragraph, Code, Separator
//   - Interactive: Button, Checkbox, Radio, Toggle, Slider
//   - Input: TextInput, TextArea, PasswordInput, NumberInput, Select
//   - Container: Box, Flex, Grid, Stack, Tabs
//   - Collections: List, Table, Tree, Calendar
//   - Widgets: ProgressBar, Spinner, Meter, Gauge, Chart
//   - Form: Form, Field, Label, Validation, Schema
//
// # Styling
//
// Components can be styled using the style package:
//
//	text := display.NewText("Hello").
//	    WithStyle(style.NewBuilder().
//	        Foreground("blue").
//	        Bold().
//	        Build())
//
// # Themes
//
// Pre-built themes are available:
//
//	app.SetTheme(style.DefaultTheme)  // or LightTheme, DraculaTheme, NordTheme
//
// # Layout
//
// Use containers to create complex layouts:
//
//	root := layout.NewFlex(layout.Column).
//	    WithChildren(
//	        display.NewText("Header"),
//	        layout.NewBox().
//	            WithPadding(1).
//	            WithChild(display.NewText("Content")),
//	        display.NewText("Footer"),
//	    )
package framework

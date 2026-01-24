package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yaoapp/yao/tui/ui/components"
	"github.com/yaoapp/yao/tui/ui/layouts"
)

// Todo represents a single todo item
type Todo struct {
	ID        string
	Text      string
	Completed bool
}

// TodoModel is the main application model
type TodoModel struct {
	todos     []Todo
	input     *components.InputComponent
	addButton *components.ButtonComponent
	list      *components.ListComponent
	header    *components.HeaderComponent
	footer    *components.FooterComponent
	focused    focusedComponent
	err       error
}

type focusedComponent int

const (
	focusedInput focusedComponent = iota
	focusedList
	focusedAddButton
)

// Init initializes the application
func (m *TodoModel) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m *TodoModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case tea.MouseMsg:
		return m, nil
	case error:
		m.err = msg
		return m, nil
	}

	return m, nil
}

// handleKey handles keyboard input
func (m *TodoModel) handleKey(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg := msg.(tea.KeyMsg)

	switch keyMsg.Type {
	case tea.KeyEnter, tea.KeyCtrlC:
		// Enter key on input adds a todo
		if m.focused == focusedInput && m.input.GetValue() != "" {
			m.addTodo(m.input.GetValue())
			m.input.SetValue("")
		}
		return m, nil

	case tea.KeyTab, tea.KeyShiftTab:
		// Tab cycles focus
		if keyMsg.Type == tea.KeyTab {
			m.cycleFocus(1)
		} else {
			m.cycleFocus(-1)
		}
		return m, nil

	case tea.KeyCtrlC, tea.KeyEsc:
		// Quit
		return m, tea.Quit

	case tea.KeyUp, tea.KeyDown:
		// Arrow keys in list
		if m.focused == focusedList {
			m.list.HandleKey(&components.KeyEvent{Key: int(keyMsg.Type)})
		}
		return m, nil

	case 'j':
		// j to move down in list
		if m.focused == focusedList {
			m.list.HandleKey(&components.KeyEvent{Key: 'j'})
		}
		return m, nil

	case 'k':
		// k to move up in list
		if m.focused == focusedList {
			m.list.HandleKey(&components.KeyEvent{Key: 'k'})
		}
		return m, nil

	case tea.KeyDelete, tea.KeyBackspace:
		// Delete/backspace to remove selected todo
		if m.focused == focusedList {
			m.removeSelectedTodo()
		}
		return m, nil

	case ' ':
		// Space to toggle completed
		if m.focused == focusedList {
			m.toggleSelectedTodo()
		}
		return m, nil

	default:
		// Pass through to focused component
		if m.focused == focusedInput {
			m.input.UpdateMsg(msg)
		} else if m.focused == focusedAddButton && keyMsg.Type == tea.KeyEnter {
			m.addButton.Click()
		}
		return m, nil
	}
}

// View renders the UI
func (m TodoModel) View() string {
	var builder strings.Builder

	// Render header
	builder.WriteString(m.header.View())
	builder.WriteString("\n")

	// Render input row
	builder.WriteString(m.input.View())
	builder.WriteString(" ")
	builder.WriteString(m.addButton.View())
	builder.WriteString("\n\n")

	// Render list
	builder.WriteString(m.list.View())
	builder.WriteString("\n")

	// Render footer
	builder.WriteString(m.footer.View())

	return builder.String()
}

// addTodo adds a new todo item
func (m *TodoModel) addTodo(text string) {
	todo := Todo{
		ID:        fmt.Sprintf("todo-%d", len(m.todos)+1),
		Text:      text,
		Completed: false,
	}
	m.todos = append(m.todos, todo)
	m.updateList()
	m.updateFooter()
}

// removeSelectedTodo removes the selected todo
func (m *TodoModel) removeSelectedTodo() {
	selectedIdx := m.list.GetSelectedIdx()
	if selectedIdx >= 0 && selectedIdx < len(m.todos) {
		m.todos = append(m.todos[:selectedIdx], m.todos[selectedIdx+1:]...)
		m.updateList()
		m.updateFooter()
	}
}

// toggleSelectedTodo toggles the completed status
func (m *TodoModel) toggleSelectedTodo() {
	selectedIdx := m.list.GetSelectedIdx()
	if selectedIdx >= 0 && selectedIdx < len(m.todos) {
		m.todos[selectedIdx].Completed = !m.todos[selectedIdx].Completed
		m.updateList()
		m.updateFooter()
	}
}

// updateList refreshes the list display
func (m *TodoModel) updateList() {
	items := make([]*components.ListItem, len(m.todos))
	for i, todo := range m.todos {
		prefix := "[ ]"
		if todo.Completed {
			prefix = "[✓]"
		}
		items[i] = &components.ListItem{
			ID:    todo.ID,
			Label: prefix + " " + todo.Text,
		}
	}
	m.list.WithItems(items)
}

// updateFooter updates the footer info
func (m *TodoModel) updateFooter() {
	total := len(m.todos)
	completed := 0
	for _, todo := range m.todos {
		if todo.Completed {
			completed++
		}
	}
	remaining := total - completed
	m.footer.WithContent(fmt.Sprintf("Total: %d | Completed: %d | Remaining: %d", total, completed, remaining))
}

// cycleFocus cycles focus between components
func (m *TodoModel) cycleFocus(direction int) {
	components := []focusedComponent{focusedInput, focusedAddButton, focusedList}
	currentIdx := int(m.focused)

	newIdx := currentIdx + direction
	if newIdx < 0 {
		newIdx = len(components) - 1
	} else if newIdx >= len(components) {
		newIdx = 0
	}

	m.focused = components[newIdx]

	// Update focus states
	m.input.SetFocus(m.focused == focusedInput)
	m.addButton.SetFocus(m.focused == focusedAddButton)
	m.list.SetFocus(m.focused == focusedList)
}

// NewTodoModel creates a new todo application model
func NewTodoModel() *TodoModel {
	// Create components
	input := components.NewInput().
		WithID("todo-input").
		WithPlaceholder("Add a new todo...").
		WithWidth(50)

	addButton := components.NewButton("Add").
		WithID("add-button").
		WithOnClick(func() {
			// Handled by Update method
		})

	list := components.NewList().
		WithID("todo-list").
		WithTitle("My Todos").
		WithSize(60, 20)

	header := components.NewHeader("📝 Todo App").
		WithAlign("center").
		WithBold(true).
		WithColor("#00aaff")

	footer := components.NewFooter("Total: 0 | Completed: 0 | Remaining: 0").
		WithAlign("center")

	model := &TodoModel{
		todos:     []Todo{},
		input:     input,
		addButton: addButton,
		list:      list,
		header:    header,
		footer:    footer,
		focused:    focusedInput,
	}

	// Add some sample todos
	model.addTodo("Buy groceries")
	model.addTodo("Complete TUI component implementation")
	model.addTodo("Write tests for all components")
	model.addTodo("Create demo application")

	return model
}

func main() {
	p := tea.NewProgram(
		NewTodoModel(),
		tea.WithAltScreen(),       // Use alternate screen
		tea.WithMouseCellMotion(), // Enable mouse support
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
	}
}

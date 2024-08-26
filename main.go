package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// memory represents a memorized CLI command with it's context.
type memory struct {
	command     string
	description string
}

// Basic Model for https://github.com/charmbracelet/bubbletea
type model struct {
	memories []memory
	cursor   int
}

func initialModel() model {
	return model{
		memories: []memory{},
		cursor:   0,
	}
}

// Init implements tea.Model.
func (m model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

// View implements tea.Model.
func (m model) View() string {
	header := "Please select a command"
	body := ""
	for _, memory := range m.memories {
		body += fmt.Sprintf("[%5.5s] [%s]", memory.command, memory.description)
	}
	return header + "\n" + body
}


func main() {
    p := tea.NewProgram(initialModel())
    if _, err := p.Run(); err != nil {
        fmt.Printf("exited with error: %v", err)
        os.Exit(1)
    }
}

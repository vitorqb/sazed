package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"

	tea "github.com/charmbracelet/bubbletea"
)

// Memory represents a memorized CLI command with it's context.
type Memory struct {
	Command     string
	Description string
}

// Basic Model for https://github.com/charmbracelet/bubbletea
type model struct {
	memories []Memory
	cursor   int
}

func initialModel() model {
	return model{
		memories: []Memory{},
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
		body += fmt.Sprintf("[%5.5s] [%s]", memory.Command, memory.Description)
	}
	return header + "\n" + body
}

type CLIOptions struct {
	MemoriesFile string
}

func ParseCliArgs(x []string) (CLIOptions, error) {
	var cliOpts CLIOptions
	flagSet := flag.NewFlagSet("sazed", flag.ContinueOnError)
	flagSet.StringVar(&cliOpts.MemoriesFile, "memories-file", "", "File to read memories from")
	err := flagSet.Parse(x)
	if err != nil {
		return cliOpts, err
	}
	if (cliOpts.MemoriesFile == "") {
		return cliOpts, fmt.Errorf("Missing memories file (--memories-file)")
	}
	return cliOpts, nil
}

func LoadMemoriesFromYaml(source io.Reader) ([]Memory, error) {
	memories := []Memory{}
	err := yaml.NewDecoder(source).Decode(&memories)
	return memories, err
}

func main() {
    p := tea.NewProgram(initialModel())
    if _, err := p.Run(); err != nil {
        fmt.Printf("exited with error: %v", err)
        os.Exit(1)
    }
}

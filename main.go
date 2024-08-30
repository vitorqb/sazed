package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"

	tea "github.com/charmbracelet/bubbletea"
)

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

// QuitWithErr signals that the program should quit
type QuitWithErr error

// The error printed when after quitting the program
var QuitErr QuitWithErr

// Memory represents a memorized CLI command with it's context.
type Memory struct {
	Command     string
	Description string
}

// Basic Model for https://github.com/charmbracelet/bubbletea
type Model struct {
	cliOpts  CLIOptions
	Memories []Memory
	cursor   int
}

// Returns the initial model
func InitialModel(cliOpts CLIOptions) Model {
	return Model{
		cliOpts:  cliOpts,
		Memories: []Memory{},
		cursor:   0,
	}
}

type LoadedMemories []Memory

func LoadMemoriesFromYaml(source io.Reader) ([]Memory, error) {
	memories := []Memory{}
	err := yaml.NewDecoder(source).Decode(&memories)
	return memories, err
}

func InitLoadMemories(cliOpts CLIOptions) tea.Cmd {
	return func() tea.Msg {
		memoriesFile, err := os.Open(cliOpts.MemoriesFile)
		if err != nil {
			return QuitWithErr(fmt.Errorf("failed to load memoriesFile: %w", err))
		}
		memories, err := LoadMemoriesFromYaml(memoriesFile)
		if err != nil {
			return QuitWithErr(fmt.Errorf("failed to load memories from yaml: %w", err))
		}
		return LoadedMemories(memories)
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return InitLoadMemories(m.cliOpts)
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	case QuitWithErr:
		// Impure, but the only way I found to quit nicely with an error
		QuitErr = msg
		return m, tea.Quit
	case LoadedMemories:
		m.Memories = msg
		return m, nil
	}
	return m, nil
}

// View implements tea.Model.
func (m Model) View() string {
	header := "Please select a command"
	body := ""
	for _, memory := range m.Memories {
		body += fmt.Sprintf("[%-35s] %s\n", memory.Command, memory.Description)
	}
	return header + "\n" + body
}

func exitWithErr(msg string, err error) {
	fmt.Fprintf(os.Stderr, "%s: %s", msg, err)
	os.Exit(1)
}

func main() {
	cliOpts, err := ParseCliArgs(os.Args[1:])
	if err != nil {
		exitWithErr("failed to parse CLI args", err)
	}
	model := InitialModel(cliOpts)
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		exitWithErr("exited with error: %v", err)
	}
	if QuitErr != nil {
		exitWithErr("error", QuitErr)
	}

}

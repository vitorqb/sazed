package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/caarlos0/env/v11"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type AppOptions struct {
	MemoriesFile string `env:"SAZED_MEMORIES_FILE"`
}

// ParseAppOptions parses the app options from (a) CLI Arguments and (b)
// a base AppOptions containing values from Env variables.
func ParseAppOptions(cliArgs []string, envOptions AppOptions) (AppOptions, error) {
	// parse CLI options
	var cliOpts AppOptions
	flagSet := flag.NewFlagSet("sazed", flag.ContinueOnError)
	flagSet.StringVar(&cliOpts.MemoriesFile, "memories-file", "", "File to read memories from")
	err := flagSet.Parse(cliArgs)
	if err != nil {
		return cliOpts, err
	}

	// join CLI and Env, priority to CLI
	appOptions := envOptions
	if cliOpts.MemoriesFile != "" {
		appOptions.MemoriesFile = cliOpts.MemoriesFile
	}

	// sanity checks
	if appOptions.MemoriesFile == "" {
		return appOptions, fmt.Errorf("Missing memories file (--memories-file)")
	}
	return appOptions, nil
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
	TextInput textinput.Model
	cliOpts   AppOptions
	Memories  []Memory
	Cursor    int
	fuzzy     IFuzzy
}

// Returns the initial model
func InitialModel(cliOpts AppOptions) Model {
	textInput := textinput.New()
	textInput.Focus()

	return Model{
		TextInput: textInput,
		cliOpts:   cliOpts,
		Memories:  []Memory{},
		Cursor:    0,
		fuzzy:     NewFuzzy(),
	}
}

type LoadedMemories []Memory

func LoadMemoriesFromYaml(source io.Reader) ([]Memory, error) {
	memories := []Memory{}
	err := yaml.NewDecoder(source).Decode(&memories)
	return memories, err
}

func InitLoadMemories(cliOpts AppOptions) tea.Cmd {
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
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
		switch msg.Type {
		case tea.KeyDown:
			m = IncreaseCursor(m)
			return m, nil
		case tea.KeyUp:
			m = DecreaseCursor(m)
			return m, nil
		}
	case QuitWithErr:
		// Impure, but the only way I found to quit nicely with an error
		QuitErr = msg
		return m, tea.Quit
	case LoadedMemories:
		m.Memories = msg
		return m, nil
	}
	m.TextInput, cmd = m.TextInput.Update(msg)
	return m, cmd
}

// View implements tea.Model.
func (m Model) View() string {
	body := "Please select a command\n"
	body += m.TextInput.View() + "\n"
	body += "----------------------\n"
	memories := m.Memories

	if inputStr := m.TextInput.Value(); inputStr != "" {
		m.fuzzy.SortByMatch(memories, inputStr)
	}

	for i, memory := range memories {
		cursor := " "
		if i == m.Cursor {
			cursor = ">"
		}
		body += fmt.Sprintf("%-2s[%-35s] %s\n", cursor, memory.Command, memory.Description)
	}

	return body
}

func exitWithErr(msg string, err error) {
	fmt.Fprintf(os.Stderr, "%s: %s", msg, err)
	os.Exit(1)
}

func main() {
	var envOpts AppOptions
	err := env.Parse(&envOpts)
	if err != nil {
		exitWithErr("failed to parse env vars", err)
	}
	appOpts, err := ParseAppOptions(os.Args[1:], envOpts)
	if err != nil {
		exitWithErr("failed to parse CLI args", err)
	}
	model := InitialModel(appOpts)
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		exitWithErr("exited with error: %v", err)
	}
	if QuitErr != nil {
		exitWithErr("error", QuitErr)
	}

}

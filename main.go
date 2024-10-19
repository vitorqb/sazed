package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path"

	"gopkg.in/yaml.v3"

	"github.com/caarlos0/env/v11"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

const DefaultCommandPrintLength = 75

type AppOptions struct {
	MemoriesFile       string `env:"SAZED_MEMORIES_FILE"`
	CommandPrintLength int    `env:"SAZED_COMMAND_PRINT_LENGTH"`
}

// Instantiates AppOptions from environmental variables
func NewAppOptionsFromEnv() (AppOptions, error) {
	var appOpts AppOptions
	err := env.Parse(&appOpts)
	if err != nil {
		return appOpts, fmt.Errorf("failed to parse from env: %w", err)
	}
	return appOpts, nil
}

// ParseAppOptions parses the app options from (a) CLI Arguments and (b)
// a base AppOptions containing values from Env variables.
func ParseAppOptions(cliArgs []string, envOptions AppOptions) (AppOptions, error) {
	// parse CLI options
	var cliOpts AppOptions
	flagSet := flag.NewFlagSet("sazed", flag.ContinueOnError)
	flagSet.StringVar(&cliOpts.MemoriesFile, "memories-file", "", "File to read memories from")
	flagSet.IntVar(&cliOpts.CommandPrintLength, "command-print-length", 0, "How many characters to print for Commands")
	err := flagSet.Parse(cliArgs)
	if err != nil {
		return cliOpts, err
	}

	// join CLI and Env, priority to CLI
	appOptions := envOptions
	if cliOpts.MemoriesFile != "" {
		appOptions.MemoriesFile = cliOpts.MemoriesFile
	}
	if cliOpts.CommandPrintLength != 0 {
		appOptions.CommandPrintLength = cliOpts.CommandPrintLength
	}

	// defaults
	if appOptions.CommandPrintLength == 0 {
		appOptions.CommandPrintLength = DefaultCommandPrintLength
	}
	if appOptions.MemoriesFile == "" {
		homeDir, _ := os.UserHomeDir()
		appOptions.MemoriesFile = path.Join(homeDir, ".config/sazed/memories.yaml")
	}

	// sanity check

	return appOptions, nil
}

// QuitWithOutput signals that the program should quit and print something to stdout.
type QuitWithOutput string

// QuitWithErr signals that the program should quit
type QuitWithErr error

// The error printed when after quitting the program
var QuitErr QuitWithErr

// An output to print when quitting
var QuitOutput QuitWithOutput

// Memory represents a memorized CLI command with it's context.
type Memory struct {
	Command     string
	Description string
}

// Basic Model for https://github.com/charmbracelet/bubbletea
type Model struct {
	// Models & Updaters
	TextInput     textinput.Model
	UpdateMatches func(memories []Memory, input string, cleanCache bool) tea.Cmd

	// Fields
	AppOpts  AppOptions
	Memories []Memory
	Matches  []Match
	Cursor   int
}

// Returns the initial model
func InitialModel(cliOpts AppOptions) Model {
	textInput := textinput.New()
	textInput.Focus()

	fuzzy := NewFuzzy()

	return Model{
		TextInput:     textInput,
		UpdateMatches: UpdateMatches(fuzzy),
		AppOpts:       cliOpts,
		Memories:      []Memory{},
		Matches:       []Match{},
		Cursor:        0,
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

type SetMatched []Match

// UpdateMatches is a curried function that returns a command re-calcualte matches
func UpdateMatches(fuzzy IFuzzy) func(memories []Memory, input string, cleanCache bool) tea.Cmd {
	first := true
	inputCache := ""
	return func(memories []Memory, input string, cleanCache bool) tea.Cmd {
		return func() tea.Msg {
			if cleanCache {
				first = true
				inputCache = ""
			}

			// If not first run and cache matches, return nil
			if !first {
				if input == inputCache {
					return nil
				}
			}

			first = false
			inputCache = input
			return SetMatched(fuzzy.GetMatches(memories, input))
		}
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return tea.Sequence(InitLoadMemories(m.AppOpts))
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
		case tea.KeyEnter:
			msg := QuitWithOutput(m.Matches[m.Cursor].Memory.Command)
			return m, func() tea.Msg { return msg }
		}
	case QuitWithErr:
		// Impure, but the only way I found to quit nicely with an error
		QuitErr = msg
		return m, tea.Quit
	case QuitWithOutput:
		// Impure, but the only way I found to quit nicely with an output
		QuitOutput = msg
		return m, tea.Quit
	case LoadedMemories:
		m.Memories = msg
		return m, m.UpdateMatches(m.Memories, m.TextInput.Value(), true)
	case SetMatched:
		m.Matches = msg
		return m, nil
	}

	// Update the text input (since it might have changed)
	m.TextInput, cmd = m.TextInput.Update(msg)

	// Update the matches to keep it in sync with the Memories/Input that may have changed
	cmd = tea.Batch(cmd, m.UpdateMatches(m.Memories, m.TextInput.Value(), false))

	return m, cmd
}

// View implements tea.Model.
func (m Model) View() string {
	body := "Please select a command\n"
	body += m.TextInput.View() + "\n"
	body += "----------------------\n"

	for i, match := range m.Matches {
		cursor := " "
		if i == m.Cursor {
			cursor = ">>"
		}
		printLength := fmt.Sprintf("%d", m.AppOpts.CommandPrintLength)

		// Prints command on first line
		format := "%-2s %-" + printLength + "." + printLength + "s\n"
		body += fmt.Sprintf(format, cursor, match.Memory.Command)

		// Prints description on second line
		body += fmt.Sprintf("      |%s\n", match.Memory.Description)
	}

	return body
}

func exitWithErr(msg string, err error) {
	fmt.Fprintf(os.Stderr, "%s: %s", msg, err)
	os.Exit(1)
}

// getOutputFile returns the output to use for the program. It tries to write to
// tty, and defaults to stdout if it can't find it.
func getOutputFile() *os.File {
	// tty is preferable because it works from subshells: echo "$(sazed)"
	if tty, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0); err == nil {
		return tty
	}
	return os.Stderr
}

func main() {
	envOpts, err := NewAppOptionsFromEnv()
	if err != nil {
		exitWithErr("failed to parse env vars", err)
	}

	appOpts, err := ParseAppOptions(os.Args[1:], envOpts)
	if err != nil {
		exitWithErr("failed to parse CLI args", err)
	}

	model := InitialModel(appOpts)

	outputFile := getOutputFile()
	defer outputFile.Close()

	p := tea.NewProgram(model, tea.WithOutput(outputFile))
	if _, err := p.Run(); err != nil {
		exitWithErr("exited with error: %v", err)
	}

	if QuitErr != nil {
		exitWithErr("error", QuitErr)
	}

	if QuitOutput != "" {
		fmt.Print(QuitOutput)
	}
}

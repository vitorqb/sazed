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

// ParseAppOptions parses the app options from CLI Arguments a map of environmental variables
func ParseAppOptions(cliArgs []string, envMap map[string]string) (AppOptions, error) {
	// parse env vars
	var opts AppOptions
	err := env.ParseWithOptions(&opts, env.Options{Environment: envMap})
	if err != nil {
		return opts, fmt.Errorf("failed to parse env vars: %w", err)
	}

	// parse CLI options
	flagSet := flag.NewFlagSet("sazed", flag.ContinueOnError)
	flagSet.StringVar(&opts.MemoriesFile, "memories-file", opts.MemoriesFile, "File to read memories from")
	flagSet.IntVar(&opts.CommandPrintLength, "command-print-length", opts.CommandPrintLength, "How many characters to print for Commands")
	err = flagSet.Parse(cliArgs)
	if err != nil {
		return opts, fmt.Errorf("failed to parse cli args: %w", err)
	}

	// defaults
	if opts.CommandPrintLength == 0 {
		opts.CommandPrintLength = DefaultCommandPrintLength
	}
	if opts.MemoriesFile == "" {
		homeDir, _ := os.UserHomeDir()
		opts.MemoriesFile = path.Join(homeDir, ".config/sazed/memories.yaml")
	}

	return opts, nil
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

// Page represents the possible pages the user is interacting with
type Page string

const PageSelect Page = "PageSelect"
const PageEdit Page = "PageEdit"

// Basic Model for https://github.com/charmbracelet/bubbletea
type Model struct {
	// Models & Updaters
	TextInput     textinput.Model
	UpdateMatches func(memories []Memory, input string, cleanCache bool) tea.Cmd

	// Fields
	AppOpts     AppOptions
	Memories    []Memory
	Matches     []Match
	MatchCursor int
	CurrentPage Page
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
		MatchCursor:   0,
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

// HandleMemorySelected is a function that reacts to the user selecting a memory.
func HandleMemorySelected(m Model) (Model, tea.Cmd) {
	SelectedMemory := m.Matches[m.MatchCursor].Memory
	countOfPlaceholders := CountPlaceholders(SelectedMemory)
	if countOfPlaceholders == 0 || !FeatureFlagPlaceholder {
		return m, func() tea.Msg {
			return QuitWithOutput(SelectedMemory.Command)
		}
	}
	m.CurrentPage = PageEdit
	return m, nil
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
			m = IncreaseMatchCursor(m)
			return m, nil
		case tea.KeyUp:
			m = DecreaseMatchCursor(m)
			return m, nil
		case tea.KeyEnter:
			m, cmd := HandleMemorySelected(m)
			return m, cmd
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
	if m.CurrentPage == PageEdit {
		return ViewCommandEdit(m)
	}
	return ViewCommandSelection(m)
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
	appOpts, err := ParseAppOptions(os.Args[1:], env.ToMap(os.Environ()))
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

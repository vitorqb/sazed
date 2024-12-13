package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path"

	"gopkg.in/yaml.v3"

	"github.com/caarlos0/env/v11"
	"github.com/charmbracelet/bubbles/cursor"
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

// The error printed when after quitting the program
var QuitErr error

// Quits the program with an error
func QuitWithErr(err error) tea.Msg {
	// Impure, but the only way I found to quit nicely with an error
	QuitErr = err
	return tea.Quit()
}

// An output to print when quitting
var QuitOutput string

// Quits the program with an output
func QuitWithOutput(output string) tea.Cmd {
	return func() tea.Msg {
		QuitOutput = output
		return tea.Quit()
	}
}

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
	SearchTextInput textinput.Model
	EditTextInputs  []textinput.Model
	UpdateMatches   func(m Model, cleanCache bool) Model
	LoadMemories    func(AppOptions) tea.Cmd

	// Fields
	AppOpts        AppOptions
	Memories       []Memory
	Matches        []Match
	MatchCursor    int
	CurrentPage    Page
	SelectedMemory Memory
}

// Returns the initial model
func InitialModel(cliOpts AppOptions) Model {
	textInput := textinput.New()
	textInput.Focus()
	textInput.Cursor.SetMode(cursor.CursorStatic)

	fuzzy := NewFuzzy()

	return Model{
		// Models & Updaters
		SearchTextInput: textInput,
		EditTextInputs:  []textinput.Model{},
		UpdateMatches:   UpdateMatches(fuzzy),
		LoadMemories:    InitLoadMemories,

		// Fields
		CurrentPage:    PageSelect,
		AppOpts:        cliOpts,
		Memories:       []Memory{},
		Matches:        []Match{},
		MatchCursor:    0,
		SelectedMemory: Memory{},
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

// UpdateMatches recalculate the matches for the given user input and list of memories
func UpdateMatches(fuzzy IFuzzy) func(m Model, cleanCache bool) Model {
	first := true
	inputCache := ""
	return func(m Model, cleanCache bool) Model {
		memories := m.Memories
		input := m.SearchTextInput.Value()
		if cleanCache {
			first = true
			inputCache = ""
		}
		if !first {
			if input == inputCache {
				return m
			}
		}
		first = false
		inputCache = input
		m.Matches = fuzzy.GetMatches(memories, input)
		return m
	}
}

// SelectCursorMemory is the logic fo when a new memory is selected based on
// existing cursor.
func SelectCursorMemory(m Model) (newModel Model, quitCmd tea.Cmd) {
	m.SelectedMemory = m.Matches[m.MatchCursor].Memory
	if !NeedsEdit(m.SelectedMemory) {
		return m, QuitWithOutput(m.SelectedMemory.Command)
	}
	m = SetupEditTextInputs(m)
	m.CurrentPage = PageEdit
	return m, nil
}

// SubmitPlaceholderValueFromInput is called when an user submits the value
// of the current placeholder value input
func SubmitPlaceholderValueFromInput(m Model) (Model, tea.Cmd) {

	// Find focusde input
	focusedInputIndex := -1
	for i, input := range m.EditTextInputs {
		if input.Focused() {
			focusedInputIndex = i
			break
		}
	}

	// If none default to the first input
	if focusedInputIndex == -1 {
		focusedInputIndex = 0
	}

	hasNextInput := len(m.EditTextInputs) >= (focusedInputIndex + 2)

	// No next input, render and return
	if !hasNextInput {
		placeholderValues := make([]string, len(m.EditTextInputs))
		for i, input := range m.EditTextInputs {
			placeholderValues[i] = input.Value()
		}
		rendered := Render(m.SelectedMemory.Command, placeholderValues)
		return m, QuitWithOutput(rendered)
	}

	// Focus next input
	m.EditTextInputs[focusedInputIndex].Blur()
	return m, tea.Batch(m.EditTextInputs[focusedInputIndex+1].Focus())
}

// SetupEditTextInputs prepares the TextInputs for the Edit page
func SetupEditTextInputs(m Model) Model {
	mem := m.SelectedMemory
	m.EditTextInputs = make([]textinput.Model, CountPlaceholders(mem.Command))
	for i, placeholder := range GetPlaceholders(mem.Command) {
		m.EditTextInputs[i] = textinput.New()
		m.EditTextInputs[i].Cursor.SetMode(cursor.CursorStatic)
		m.EditTextInputs[i].Prompt = placeholder.Name + ": "
		if i == 0 {
			m.EditTextInputs[i].Focus()
		}
	}
	return m
}

// LoadMemories handle memories loaded
func LoadMemories(m Model, mems []Memory) Model {
	m.Memories = mems
	m = m.UpdateMatches(m, true)
	return m
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd { return m.LoadMemories(m.AppOpts) }

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
		switch m.CurrentPage {
		case PageSelect:
			switch msg.Type {
			case tea.KeyDown:
				return IncreaseMatchCursor(m), nil
			case tea.KeyUp:
				return DecreaseMatchCursor(m), nil
			case tea.KeyEnter:
				return SelectCursorMemory(m)
			}
		case PageEdit:
			switch msg.Type {
			case tea.KeyEnter:
				return SubmitPlaceholderValueFromInput(m)
			}
		}
	case LoadedMemories:
		return LoadMemories(m, msg), nil
	case SetMatched:
		m.Matches = msg
		return m, nil
	}

	// Cmd to return
	var cmd tea.Cmd

	// Update the text input (since it might have changed)
	if m.CurrentPage == PageSelect {
		m.SearchTextInput, cmd = m.SearchTextInput.Update(msg)
	}

	// Update the Edit view text inputs
	if m.CurrentPage == PageEdit {
		var editTextInputsCmds tea.Cmd
		m.EditTextInputs, editTextInputsCmds = m.UpdateEditTextInputs(msg)
		cmd = tea.Batch(cmd, editTextInputsCmds)
	}

	// Update the matches to keep it in sync with the Memories/Input that may have changed
	m = m.UpdateMatches(m, false)

	return m, cmd
}

// UpdateEditTextInputs updates the text inputs for the edit page
func (m *Model) UpdateEditTextInputs(msg tea.Msg) ([]textinput.Model, tea.Cmd) {
	if m.CurrentPage != PageEdit {
		return m.EditTextInputs, nil
	}
	var cmd tea.Cmd
	for i := 0; i < len(m.EditTextInputs); i++ {
		var aCmd tea.Cmd
		m.EditTextInputs[i], aCmd = m.EditTextInputs[i].Update(msg)
		cmd = tea.Batch(cmd, aCmd)
	}
	return m.EditTextInputs, cmd
}

// View implements tea.Model.
func (m Model) View() string {
	if m.CurrentPage == PageEdit {
		return ViewCommandEdit(m)
	}
	return ViewCommandSelection(m)
}

func (m Model) GetPlaceholderValues() []string {
	out := make([]string, len(m.EditTextInputs))
	for i, textInp := range m.EditTextInputs {
		out[i] = textInp.Value()
	}
	return out
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

// NeedsEdit returns True if a memory needs to be edited before returning
func NeedsEdit(m Memory) bool {
	if !FeatureFlagPlaceholder {
		return false
	}
	if countOfPlace := CountPlaceholders(m.Command); countOfPlace == 0 {
		return false
	}
	return true
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

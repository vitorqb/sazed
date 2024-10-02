package main_test

import (
	"os"
	"path"
	"strings"
	"testing"

	"github.com/caarlos0/env/v11"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	sazed "github.com/vitorqb/sazed"
)

func cleanup() {
	// Cleanup the global QuitErr
	sazed.QuitErr = nil
}
func memory1() sazed.Memory {
	return sazed.Memory{Command: "cmd1", Description: "Memory 1"}
}

func memory2() sazed.Memory {
	return sazed.Memory{Command: "foo", Description: "Bar"}
}

func memory3() sazed.Memory {
	return sazed.Memory{Command: "not foo", Description: "not bar"}
}

func newTestModel() sazed.Model {
	return sazed.InitialModel(sazed.AppOptions{
		CommandPrintLength: sazed.DefaultCommandPrintLength,
	})
}

func update(m tea.Model, msg tea.Msg) tea.Model {
	return batchUpdate(m, func() tea.Msg { return msg })
}

func batchUpdate(m tea.Model, cmd tea.Cmd) tea.Model {
	if cmd == nil {
		return m
	}
	msg := cmd()
	m2, cmd := m.Update(msg)
	return batchUpdate(m2, cmd)
}

func Test__NewAppOptionsFromEnv(t *testing.T) {
	t.Cleanup(cleanup)
	t.Run("instantiates from env vars", func(t *testing.T) {
		t.Setenv("SAZED_MEMORIES_FILE", "/foo")
		t.Setenv("SAZED_COMMAND_PRINT_LENGTH", "999")

		envOpts, err := sazed.NewAppOptionsFromEnv()

		assert.Nil(t, err)
		assert.Equal(t, sazed.AppOptions{
			MemoriesFile:       "/foo",
			CommandPrintLength: 999,
		}, envOpts)
	})
	t.Run("fails because wrong format of CommandPrintLength", func(t *testing.T) {
		t.Setenv("SAZED_COMMAND_PRINT_LENGTH", "aaa")

		_, err := sazed.NewAppOptionsFromEnv()

		assert.ErrorContains(t, err, "CommandPrintLength")
	})
}

func Test__ParseAppOptions(t *testing.T) {
	t.Cleanup(cleanup)
	t.Run("parse all args", func(t *testing.T) {
		// Priority should be given to `args`, not `envOpts`
		args := []string{
			"--memories-file", "/tmp/foo",
			"--command-print-length", "45",
		}
		envOpts := sazed.AppOptions{MemoriesFile: "/tmp/bar"}
		parsed, err := sazed.ParseAppOptions(args, envOpts)
		assert.Nil(t, err)
		expected := sazed.AppOptions{
			MemoriesFile:       "/tmp/foo",
			CommandPrintLength: 45,
		}
		assert.Equal(t, expected, parsed)
	})
	t.Run("parse from env", func(t *testing.T) {
		args := []string{}
		envOpts := sazed.AppOptions{MemoriesFile: "/tmp/foo"}
		parsed, err := sazed.ParseAppOptions(args, envOpts)
		assert.Nil(t, err)
		expected := sazed.AppOptions{
			MemoriesFile:       "/tmp/foo",
			CommandPrintLength: sazed.DefaultCommandPrintLength,
		}
		assert.Equal(t, expected, parsed)
	})
	t.Run("missing memories file", func(t *testing.T) {
		args := []string{}
		_, err := sazed.ParseAppOptions(args, sazed.AppOptions{})
		assert.ErrorContains(t, err, "--memories-file")
	})
}

func Test__AppOptions(t *testing.T) {
	t.Cleanup(cleanup)
	t.Run("reads from env", func(t *testing.T) {
		t.Setenv("SAZED_MEMORIES_FILE", "foo")
		var appOpts sazed.AppOptions
		err := env.Parse(&appOpts)
		assert.Nil(t, err)
		assert.Equal(t, "foo", appOpts.MemoriesFile)
	})
}

func Test__LoadMemoriesFromYaml(t *testing.T) {
	t.Cleanup(cleanup)
	t.Run("loads empty array", func(t *testing.T) {
		yamlContent := "[]"
		reader := strings.NewReader(yamlContent)
		memories, err := sazed.LoadMemoriesFromYaml(reader)
		assert.Nil(t, err)
		assert.Equal(t, []sazed.Memory{}, memories)
	})
	t.Run("loads two memories", func(t *testing.T) {
		yamlContent := ""
		yamlContent += "- {command: \"foo\", description: \"bar\"}\n"
		yamlContent += "- {command: \"bar\", description: \"baz\"}\n"
		reader := strings.NewReader(yamlContent)
		memories, err := sazed.LoadMemoriesFromYaml(reader)
		assert.Nil(t, err)
		assert.Equal(t, []sazed.Memory{
			{Command: "foo", Description: "bar"},
			{Command: "bar", Description: "baz"},
		}, memories)
	})
}

func Test__InitLoadMemories(t *testing.T) {
	t.Cleanup(cleanup)
	t.Run("load memories from yaml", func(t *testing.T) {
		memoriesFile := path.Join(t.TempDir(), "foo")
		memoriesFileContent := "- {command: foo, description: bar}"
		os.WriteFile(memoriesFile, []byte(memoriesFileContent), 0644)
		appOpts := sazed.AppOptions{MemoriesFile: memoriesFile}

		msg := sazed.InitLoadMemories(appOpts)()

		assert.Equal(t, msg, sazed.LoadedMemories([]sazed.Memory{
			{Command: "foo", Description: "bar"},
		}))
	})
	t.Run("report error if loaded from file", func(t *testing.T) {
		memoriesFile := path.Join(t.TempDir(), "foo")
		appOpts := sazed.AppOptions{MemoriesFile: memoriesFile}

		msg := sazed.InitLoadMemories(appOpts)()

		assert.ErrorContains(t, msg.(error), "file")
	})
	t.Run("report error if invalid yaml", func(t *testing.T) {
		memoriesFile := path.Join(t.TempDir(), "foo")
		memoriesFileContent := "INV{A}LID{YAML"
		os.WriteFile(memoriesFile, []byte(memoriesFileContent), 0644)
		appOpts := sazed.AppOptions{MemoriesFile: memoriesFile}

		msg := sazed.InitLoadMemories(appOpts)()

		assert.ErrorContains(t, msg.(error), "unmarshal error")
	})
}

func Test__Update(t *testing.T) {
	t.Cleanup(cleanup)
	t.Run("handles LoadedMemories", func(t *testing.T) {
		var model tea.Model

		model = newTestModel()
		msg := sazed.LoadedMemories([]sazed.Memory{
			{Command: "foo", Description: "bar"},
		})
		model = update(model, msg)
		assert.Equal(t, []sazed.Memory(msg), model.(sazed.Model).Memories)
	})
	t.Run("handles enter key", func(t *testing.T) {
		model := sazed.Model{}
		model.Memories = []sazed.Memory{memory1()}
		msg := tea.KeyMsg{Type: tea.KeyEnter}
		newModel, cmd := model.Update(msg)

		// Model is unchanged
		assert.Equal(t, model, newModel)

		// Runs the command. It must return a msg telling the program to quit
		// with output.
		newMsg := cmd()
		assert.Equal(t, sazed.QuitWithOutput(memory1().Command), newMsg)

	})
}

func Test__View(t *testing.T) {
	t.Cleanup(cleanup)
	t.Run("renders view with a single memory", func(t *testing.T) {
		msg := sazed.LoadedMemories([]sazed.Memory{memory1()})
		model := update(newTestModel(), msg)

		rendered := strings.Split(model.View(), "\n")

		assert.Equal(t, "Please select a command", rendered[0])
		assert.Contains(t, rendered[3], "cmd1")
		assert.Contains(t, rendered[4], "Memory 1")
	})
	t.Run("renders an input field", func(t *testing.T) {
		model := update(newTestModel(), tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{'a'},
		})

		rendered := strings.Split(model.View(), "\n")

		assert.Equal(t, "> a ", rendered[1])
	})
	t.Run("commands with long width are trimmed", func(t *testing.T) {
		model := newTestModel()
		memory := memory1()
		memory.Command = strings.Repeat("x", 100)
		model.Matches = []sazed.Match{{Memory: memory}}
		cmdPrintLen := 30
		model.AppOpts.CommandPrintLength = cmdPrintLen

		rendered := strings.Split(model.View(), "\n")

		expected := memory.Command[0:30]
		assert.Contains(t, rendered[3], expected)
		assert.NotContains(t, rendered[3], memory.Command)
	})
	t.Run("sort memories by fuzzy search", func(t *testing.T) {
		model := newTestModel()
		model.Memories = []sazed.Memory{memory1(), memory2(), memory3()}

		// Simulate user writting foo
		msg := tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{'b', 'a', 'r'},
		}
		model = update(model, msg).(sazed.Model)

		// Expect ordered
		rendered := strings.Split(model.View(), "\n")
		assert.Contains(t, rendered[4], "Bar")
		assert.Contains(t, rendered[6], "not bar")

		// First command should not be there
		assert.Len(t, rendered, 8)
	})
	t.Run("moves cursor around", func(t *testing.T) {
		// Load memories
		memories := sazed.LoadedMemories([]sazed.Memory{memory1(), memory2(), memory3()})
		model := update(newTestModel(), memories)

		// Simulate kew down
		msg := tea.KeyMsg{Type: tea.KeyDown}
		model = update(model, msg)

		// Find the cursor at second memory
		rendered := strings.Split(model.View(), "\n")
		assert.Contains(t, rendered[3], "   cmd1")
		assert.Contains(t, rendered[5], ">> foo")

		// Simulate kew up
		model = update(model, tea.KeyMsg{Type: tea.KeyUp})

		// Find the cursor at the first memory
		rendered = strings.Split(model.View(), "\n")
		assert.Contains(t, rendered[3], ">> cmd1")
		assert.Contains(t, rendered[5], "   foo")
	})
}

type FakeFuzzy struct {
	mockResult []sazed.Match
}

func (f *FakeFuzzy) GetMatches(memories []sazed.Memory, input string) []sazed.Match {
	return f.mockResult
}

func Test__UpdateMatches(t *testing.T) {
	t.Run("no memories no input", func(t *testing.T) {
		fuzzy := &FakeFuzzy{[]sazed.Match{}}
		memories := []sazed.Memory{}
		input := ""
		updateMatches := sazed.UpdateMatches(fuzzy)

		result := updateMatches(memories, input, false)()

		assert.Equal(t, sazed.SetMatched([]sazed.Match{}), result)
	})

	t.Run("return nil if input unchanged input", func(t *testing.T) {
		matches := []sazed.Match{{Score: 20}}
		fuzzy := &FakeFuzzy{matches}
		memories := []sazed.Memory{}
		input := "foo"
		updateMatches := sazed.UpdateMatches(fuzzy)

		resultOne := updateMatches(memories, input, false)()
		assert.Equal(t, sazed.SetMatched(matches), resultOne)

		resultTwo := updateMatches(memories, input, false)()
		assert.Equal(t, nil, resultTwo)
	})

	t.Run("return update result if input changed", func(t *testing.T) {
		matches := []sazed.Match{{Score: 20}}
		fuzzy := &FakeFuzzy{matches}
		memories := []sazed.Memory{}
		updateMatches := sazed.UpdateMatches(fuzzy)

		inputOne := "foo"
		resultOne := updateMatches(memories, inputOne, false)()
		assert.Equal(t, sazed.SetMatched(matches), resultOne)

		inputTwo := "foo2"
		resultTwo := updateMatches(memories, inputTwo, false)()
		assert.Equal(t, sazed.SetMatched(matches), resultTwo)
	})

	t.Run("cleans cache", func(t *testing.T) {
		matches := []sazed.Match{{Score: 20}}
		fuzzy := &FakeFuzzy{matches}
		memories := []sazed.Memory{}
		updateMatches := sazed.UpdateMatches(fuzzy)

		input := "foo"
		result := updateMatches(memories, input, true)()
		assert.Equal(t, sazed.SetMatched(matches), result)
		assert.Equal(t, sazed.SetMatched(matches), result)
	})
}

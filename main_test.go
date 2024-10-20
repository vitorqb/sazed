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

func Test__ParseAppOptions(t *testing.T) {
	t.Cleanup(cleanup)

	t.Run("all default values", func(t *testing.T) {
		env := map[string]string{}
		args := []string{}

		opts, err := sazed.ParseAppOptions(args, env)

		assert.Nil(t, err)
		assert.Contains(t, opts.MemoriesFile, ".config/sazed/memories.yaml")
		assert.Equal(t, opts.CommandPrintLength, sazed.DefaultCommandPrintLength)
	})

	t.Run("args have preference over env", func(t *testing.T) {
		env := map[string]string{
			"SAZED_MEMORIES_FILE":        "/foo",
			"SAZED_COMMAND_PRINT_LENGTH": "40",
		}
		args := []string{
			"--memories-file=/bar",
			"--command-print-length=999",
		}

		opts, err := sazed.ParseAppOptions(args, env)

		assert.Nil(t, err)
		assert.Equal(t, opts.MemoriesFile, "/bar")
		assert.Equal(t, opts.CommandPrintLength, 999)
	})

	t.Run("from env if args not set", func(t *testing.T) {
		env := map[string]string{
			"SAZED_MEMORIES_FILE":        "/foo",
			"SAZED_COMMAND_PRINT_LENGTH": "40",
		}
		args := []string{}

		opts, err := sazed.ParseAppOptions(args, env)

		assert.Nil(t, err)
		assert.Equal(t, opts.MemoriesFile, "/foo")
		assert.Equal(t, opts.CommandPrintLength, 40)
	})

	t.Run("errors if unexpected type (env)", func(t *testing.T) {
		env := map[string]string{
			"SAZED_COMMAND_PRINT_LENGTH": "a", // Should be int
		}
		args := []string{}

		_, err := sazed.ParseAppOptions(args, env)

		assert.ErrorContains(t, err, "failed to parse env vars")
	})

	t.Run("errors if unexpected type (cli)", func(t *testing.T) {
		env := map[string]string{}
		args := []string{
			"--command-print-length=aaa", // Should be int
		}

		_, err := sazed.ParseAppOptions(args, env)

		assert.ErrorContains(t, err, "failed to parse cli args")
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
		_ = os.WriteFile(memoriesFile, []byte(memoriesFileContent), 0644)
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
		_ = os.WriteFile(memoriesFile, []byte(memoriesFileContent), 0644)
		appOpts := sazed.AppOptions{MemoriesFile: memoriesFile}

		msg := sazed.InitLoadMemories(appOpts)()

		assert.ErrorContains(t, msg.(error), "unmarshal error")
	})
}

func Test__Update(t *testing.T) {
	t.Cleanup(cleanup)
	t.Run("handles LoadedMemories", func(t *testing.T) {
		model := newTestModel()
		msg := sazed.LoadedMemories([]sazed.Memory{
			{Command: "foo", Description: "bar"},
		})
		model = update(model, msg).(sazed.Model)
		assert.Equal(t, []sazed.Memory(msg), model.Memories)
	})
	t.Run("handles enter key", func(t *testing.T) {
		model := sazed.Model{}
		model.Matches = []sazed.Match{{Memory: memory1()}}
		msg := tea.KeyMsg{Type: tea.KeyEnter}
		newModel, cmd := model.Update(msg)

		// Model is unchanged
		assert.Equal(t, model, newModel)

		// Runs the command. It must return a msg telling the program to quit
		// with output.
		newMsg := cmd()
		assert.Equal(t, sazed.QuitWithOutput(memory1().Command), newMsg)

	})
	t.Run("selects memory from user input", func(t *testing.T) {
		// Load memories
		memories := []sazed.Memory{memory1(), memory2()}
		m := update(newTestModel(), sazed.LoadedMemories(memories))

		// User inputs Bar and selects second memory
		m = update(m, tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{'b', 'a', 'r'},
		})

		// User hits enter
		_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		// QuitWithOutput is sent
		assert.Equal(t, sazed.QuitWithOutput(memory2().Command), cmd())
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

package main_test

import (
	"os"
	"path"
	"strings"
	"testing"

	"github.com/caarlos0/env/v11"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	sazed "github.com/vitorqb/sazed"
)

func cleanup() {
	// Cleanup the global QuitErr
	sazed.QuitErr = nil
	sazed.QuitOutput = ""
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

func memory4() sazed.Memory {
	return sazed.Memory{Command: "echo {{value}}", Description: "not bar"}
}

func memory5() sazed.Memory {
	return sazed.Memory{Command: "echo {{value1}} {{value2}} end", Description: "not bar"}
}

func newTestModel() sazed.Model {
	return sazed.InitialModel(sazed.AppOptions{
		CommandPrintLength: sazed.DefaultCommandPrintLength,
	})
}

func update(m sazed.Model, msg tea.Msg) sazed.Model {
	return batchUpdate(m, func() tea.Msg { return msg })
}

func batchUpdate(m sazed.Model, cmd tea.Cmd) sazed.Model {
	if cmd == nil {
		return m
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); ok {
		return m
	}
	m2, cmd := m.Update(msg)
	return batchUpdate(m2.(sazed.Model), cmd)
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
		defer cleanup()
		memoriesFile := path.Join(t.TempDir(), "foo")
		appOpts := sazed.AppOptions{MemoriesFile: memoriesFile}

		msg := sazed.InitLoadMemories(appOpts)()

		assert.Equal(t, tea.QuitMsg{}, msg)
		assert.ErrorContains(t, sazed.QuitErr, "file")
	})
	t.Run("report error if invalid yaml", func(t *testing.T) {
		defer cleanup()
		memoriesFile := path.Join(t.TempDir(), "foo")
		memoriesFileContent := "INV{A}LID{YAML"
		_ = os.WriteFile(memoriesFile, []byte(memoriesFileContent), 0644)
		appOpts := sazed.AppOptions{MemoriesFile: memoriesFile}

		msg := sazed.InitLoadMemories(appOpts)()

		assert.Equal(t, tea.QuitMsg{}, msg)
		assert.ErrorContains(t, sazed.QuitErr, "unmarshal error")
	})
}

func Test__LoadMemories(t *testing.T) {
	t.Cleanup(cleanup)

	memories := []sazed.Memory{memory1()}
	m := newTestModel()
	m = sazed.LoadMemories(m, memories)
	assert.Equal(t, memories, m.Memories)
	assert.Len(t, m.Matches, 1)
}

func Test__Update(t *testing.T) {
	t.Cleanup(cleanup)

	t.Run("handles enter key", func(t *testing.T) {
		defer cleanup()
		model := newTestModel()
		// Load memories
		model = sazed.LoadMemories(model, []sazed.Memory{memory1()})
		// User writes and matches first one
		model.SearchTextInput.SetValue(memory1().Description)
		// Update matches with user input
		model = model.UpdateMatches(model, true)
		// User hits enter
		teaModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

		// We quit with msg after setting the selected memory
		assert.Equal(t, tea.QuitMsg{}, cmd())
		assert.Equal(t, memory1().Command, sazed.QuitOutput)
		assert.Equal(t, memory1(), teaModel.(sazed.Model).SelectedMemory)
	})
	t.Run("selects memory from user input (no placeholder)", func(t *testing.T) {
		defer cleanup()
		model := newTestModel()
		// Load memories
		memories := []sazed.Memory{memory1(), memory2()}
		model = sazed.LoadMemories(model, memories)

		// User inputs Bar and selects second memory
		model.SearchTextInput.SetValue("bar")
		model = model.UpdateMatches(model, true)

		// User hits enter
		model, cmd := sazed.SelectCursorMemory(model)

		// QuitWithOutput is sent
		assert.Equal(t, tea.QuitMsg{}, cmd())
		assert.Equal(t, memory2().Command, sazed.QuitOutput)
	})
	t.Run("selects memory fom user with placehold FF off", func(t *testing.T) {
		defer cleanup()
		defer sazed.SetFeatureFlagPlaceholder(false)()

		// Load memories
		memories := []sazed.Memory{memory4()}
		m := update(newTestModel(), sazed.LoadedMemories(memories))

		// User hits enter
		_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		// QuitWithOutput is sent
		assert.Equal(t, tea.QuitMsg{}, cmd())
		assert.Equal(t, memory4().Command, sazed.QuitOutput)
	})
	t.Run("selects memory from user input with placehoder FF on", func(t *testing.T) {
		defer sazed.SetFeatureFlagPlaceholder(true)()

		// Load memories
		memories := []sazed.Memory{memory4()}
		m := update(newTestModel(), sazed.LoadedMemories(memories))

		// User hits enter
		m = update(m, tea.KeyMsg{Type: tea.KeyEnter})

		// Edit mode is set
		assert.Equal(t, sazed.PageEdit, m.CurrentPage)
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
		model = update(model, msg)

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
	t.Run("renders edit view", func(t *testing.T) {
		// Prepare model with view and page
		defer sazed.SetFeatureFlagPlaceholder(true)()
		var cmd tea.Cmd
		var m tea.Model = newTestModel()
		m, cmd = m.Update(sazed.LoadedMemories([]sazed.Memory{memory4()}))
		assert.Nil(t, cmd)
		m, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
		assert.Nil(t, cmd)

		// Render
		rendered := strings.Split(m.View(), "\n")

		// Find key lines to test
		assert.Equal(t, sazed.PageEdit, m.(sazed.Model).CurrentPage)
		assert.Equal(t, "Command: echo ", rendered[0])
	})
}

func Test__GetPlaceholderValues(t *testing.T) {
	t.Cleanup(cleanup)
	t.Run("Returns values from text input", func(t *testing.T) {
		textInputs := make([]textinput.Model, 2)
		textInputs[0] = textinput.New()
		textInputs[0].SetValue("foo")
		textInputs[1] = textinput.New()
		textInputs[1].SetValue("bar")
		m := newTestModel()
		m.EditTextInputs = textInputs
		assert.Equal(t, []string{"foo", "bar"}, m.GetPlaceholderValues())
	})
}

type FakeFuzzy struct {
	mockResult []sazed.Match
}

func (f *FakeFuzzy) GetMatches(memories []sazed.Memory, input string) []sazed.Match {
	return f.mockResult
}

func Test__UpdateMatches(t *testing.T) {
	t.Cleanup(cleanup)

	t.Run("skip if no memories and no input", func(t *testing.T) {
		updateMatches := sazed.UpdateMatches(&FakeFuzzy{[]sazed.Match{}})
		model := newTestModel()
		model.SearchTextInput.SetValue("")
		model.Memories = []sazed.Memory{}
		model = updateMatches(model, false)
		assert.Equal(t, []sazed.Match{}, model.Matches)
	})

	t.Run("don't change matches if unchanged input", func(t *testing.T) {
		oldMatches := []sazed.Match{{Score: 20}}
		newMatches := []sazed.Match{{Score: 30}}
		updateMatches := sazed.UpdateMatches(&FakeFuzzy{newMatches})
		model := newTestModel()
		model.Matches = oldMatches

		// Fist call with "foo", ignore results
		model.SearchTextInput.SetValue("foo")
		_ = updateMatches(model, false)

		// Second call with "foo", save results
		model.SearchTextInput.SetValue("foo")
		model = updateMatches(model, false)

		// Old matches are still there (cache)
		assert.Equal(t, oldMatches, model.Matches)
	})

	t.Run("return update result if input changed", func(t *testing.T) {
		oldMatches := []sazed.Match{{Score: 20}}
		newMatches := []sazed.Match{{Score: 30}}
		updateMatches := sazed.UpdateMatches(&FakeFuzzy{newMatches})
		model := newTestModel()
		model.Matches = oldMatches

		// Fist call with "foo", ignore results
		model.SearchTextInput.SetValue("foo")
		_ = updateMatches(model, false)

		// Second call wth "bar"
		model.SearchTextInput.SetValue("bar")
		model = updateMatches(model, false)

		// New matches are set
		assert.Equal(t, newMatches, model.Matches)
	})

	t.Run("updates matches if cache clean", func(t *testing.T) {
		oldMatches := []sazed.Match{{Score: 20}}
		newMatches := []sazed.Match{{Score: 30}}
		updateMatches := sazed.UpdateMatches(&FakeFuzzy{newMatches})
		model := newTestModel()
		model.Matches = oldMatches

		// Fist call with "foo", ignore results
		model.SearchTextInput.SetValue("foo")
		_ = updateMatches(model, false)

		// Second call with "foo", save results
		model.SearchTextInput.SetValue("foo")
		model = updateMatches(model, true)

		// New matches are there since we cleaned cache
		assert.Equal(t, newMatches, model.Matches)
	})
}

func Test__SelectCursorMemory(t *testing.T) {
	t.Cleanup(cleanup)

	t.Run("quits if memory does not need edit", func(t *testing.T) {
		m := newTestModel()
		m = sazed.LoadMemories(m, []sazed.Memory{memory1()})

		m, cmd := sazed.SelectCursorMemory(m)

		assert.Equal(t, m.SelectedMemory, memory1())
		assert.Equal(t, sazed.QuitWithOutput(memory1().Command)(), cmd())
	})
	t.Run("quits if memory has placeholders but FF is off", func(t *testing.T) {
		defer sazed.SetFeatureFlagPlaceholder(false)()
		m := newTestModel()
		m = sazed.LoadMemories(m, []sazed.Memory{memory5()})

		m, cmd := sazed.SelectCursorMemory(m)

		assert.Equal(t, m.SelectedMemory, memory5())
		assert.Equal(t, sazed.QuitWithOutput(memory5().Command)(), cmd())
	})
	t.Run("goes to edit if memory has placeholders and FF on", func(t *testing.T) {
		defer sazed.SetFeatureFlagPlaceholder(true)()
		m := newTestModel()
		m = sazed.LoadMemories(m, []sazed.Memory{memory5()})

		m, cmd := sazed.SelectCursorMemory(m)

		assert.Nil(t, cmd)
		assert.Equal(t, m.SelectedMemory, memory5())
		assert.Equal(t, m.CurrentPage, sazed.PageEdit)
		assert.Len(t, m.EditTextInputs, 2)
		assert.Equal(t, "value1: ", m.EditTextInputs[0].Prompt)
	})
}

func Test__SetupEditTextInputs(t *testing.T) {
	t.Cleanup(cleanup)

	t.Run("Creates 2 text inputs for 2 placeholders", func(t *testing.T) {
		m := newTestModel()
		m.SelectedMemory = memory5()

		m = sazed.SetupEditTextInputs(m)

		assert.Len(t, m.EditTextInputs, 2)                      // 2 placeholders
		assert.True(t, m.EditTextInputs[0].Focused())           // 1st is focused
		assert.Equal(t, "value2: ", m.EditTextInputs[1].Prompt) // Pompt is there
	})
}

func Test__UpdateEditTextInputs(t *testing.T) {
	t.Run("skip if not on edit page", func(t *testing.T) {
		m := newTestModel()
		m.CurrentPage = sazed.PageSelect
		m.EditTextInputs = append(m.EditTextInputs, textinput.New())
		m.EditTextInputs[0].Focus()
		updated, cmd := m.UpdateEditTextInputs(tea.KeyMsg{})
		assert.Nil(t, cmd)
		assert.Equal(t, m.EditTextInputs, updated)
	})
	t.Run("sets value if input", func(t *testing.T) {
		m := newTestModel()
		m = sazed.LoadMemories(m, []sazed.Memory{memory5()})
		m = m.UpdateMatches(m, true)
		m, _ = sazed.SelectCursorMemory(m)
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f', 'o', 'o'}}
		var cmd tea.Cmd
		m.EditTextInputs, cmd = m.UpdateEditTextInputs(msg)
		m = batchUpdate(m, cmd)
		assert.Equal(t, "foo", m.EditTextInputs[0].Value())
	})
}

func Test__NeedsEdit(t *testing.T) {
	t.Run("FF is on", func(t *testing.T) {
		defer sazed.SetFeatureFlagPlaceholder(true)()
		assert.False(t, sazed.NeedsEdit(memory1()))
		assert.False(t, sazed.NeedsEdit(memory2()))
		assert.False(t, sazed.NeedsEdit(memory3()))
		assert.True(t, sazed.NeedsEdit(memory4()))
		assert.True(t, sazed.NeedsEdit(memory5()))
	})
	t.Run("FF is off", func(t *testing.T) {
		defer sazed.SetFeatureFlagPlaceholder(false)()
		assert.False(t, sazed.NeedsEdit(memory1()))
		assert.False(t, sazed.NeedsEdit(memory2()))
		assert.False(t, sazed.NeedsEdit(memory3()))
		assert.False(t, sazed.NeedsEdit(memory4()))
		assert.False(t, sazed.NeedsEdit(memory5()))
	})
}

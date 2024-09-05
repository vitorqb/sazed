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

func Test__ParseAppOptions(t *testing.T) {
	t.Cleanup(cleanup)
	t.Run("parse all args", func(t *testing.T) {
		// Priority should be given to `args`, not `envOpts`
		args := []string{"--memories-file", "/tmp/foo"}
		envOpts := sazed.AppOptions{MemoriesFile: "/tmp/bar"}
		parsed, err := sazed.ParseAppOptions(args, envOpts)
		assert.Nil(t, err)
		expected := sazed.AppOptions{
			MemoriesFile: "/tmp/foo",
		}
		assert.Equal(t, expected, parsed)
	})
	t.Run("parse from env", func(t *testing.T) {
		args := []string{}
		envOpts := sazed.AppOptions{MemoriesFile: "/tmp/foo"}
		parsed, err := sazed.ParseAppOptions(args, envOpts)
		assert.Nil(t, err)
		expected := sazed.AppOptions{
			MemoriesFile: "/tmp/foo",
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
		model := sazed.Model{}
		msg := sazed.LoadedMemories([]sazed.Memory{
			{Command: "foo", Description: "bar"},
		})
		newModel, cmd := model.Update(msg)
		assert.Nil(t, cmd)
		assert.Equal(t, []sazed.Memory(msg), newModel.(sazed.Model).Memories)
	})
}

func Test__View(t *testing.T) {
	t.Cleanup(cleanup)
	t.Run("renders view with a single memory", func(t *testing.T) {
		model := sazed.InitialModel(sazed.AppOptions{})
		model.Memories = []sazed.Memory{memory1()}
		rendered := strings.Split(model.View(), "\n")
		assert.Equal(t, "Please select a command", rendered[0])
		assert.Contains(t, rendered[3], "cmd1")
		assert.Contains(t, rendered[3], "Memory 1")
	})
	t.Run("renders an input field", func(t *testing.T) {
		model := sazed.InitialModel(sazed.AppOptions{})
		model.TextInput, _ = model.TextInput.Update(tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{'a'},
		})
		rendered := strings.Split(model.View(), "\n")
		assert.Equal(t, "> a\x1b[7m \x1b[0m", rendered[1])
	})
	t.Run("sort memories by fuzzy search", func(t *testing.T) {
		model := sazed.InitialModel(sazed.AppOptions{})
		model.Memories = []sazed.Memory{memory1(), memory2(), memory3()}

		// Simulate user writting foo
		msg := tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{'b', 'a', 'r'},
		}
		newModel, _ := model.Update(msg)

		// Expect ordered
		rendered := strings.Split(newModel.View(), "\n")
		assert.Contains(t, rendered[3], "Bar")
		assert.Contains(t, rendered[4], "not bar")
	})
	t.Run("moves cursor around", func(t *testing.T) {
		model := sazed.InitialModel(sazed.AppOptions{})
		model.Memories = []sazed.Memory{memory1(), memory2(), memory3()}

		// Simulate kew down
		msg := tea.KeyMsg{Type: tea.KeyDown}
		newModel, newMsg := model.Update(msg)
		assert.Nil(t, newMsg)

		// Find the cursor at second memory
		rendered := strings.Split(newModel.View(), "\n")
		assert.Contains(t, rendered[3], "  [cmd1")
		assert.Contains(t, rendered[4], "> [foo")

		// Simulate kew up
		msg = tea.KeyMsg{Type: tea.KeyUp}
		newModel, newMsg = newModel.Update(msg)
		assert.Nil(t, newMsg)

		// Find the cursor at the first memory
		rendered = strings.Split(newModel.View(), "\n")
		assert.Contains(t, rendered[3], "> [cmd1")
		assert.Contains(t, rendered[4], "  [foo")
	})
}

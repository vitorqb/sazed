package main_test

import (
	"os"
	"path"
	"strings"
	"testing"

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

func Test__ParseCliArgs(t *testing.T) {
	t.Cleanup(cleanup)
	t.Run("parse all args", func(t *testing.T) {
		args := []string{"--memories-file", "/tmp/foo"}
		parsed, err := sazed.ParseCliArgs(args)
		assert.Nil(t, err)
		expected := sazed.CLIOptions{
			MemoriesFile: "/tmp/foo",
		}
		assert.Equal(t, expected, parsed)
	})
	t.Run("missing memories file", func(t *testing.T) {
		args := []string{}
		_, err := sazed.ParseCliArgs(args)
		assert.ErrorContains(t, err, "--memories-file")
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
		cliOpts := sazed.CLIOptions{ MemoriesFile: memoriesFile }
	
		msg := sazed.InitLoadMemories(cliOpts)()

		assert.Equal(t, msg, sazed.LoadedMemories([]sazed.Memory{
			{Command:     "foo", Description: "bar"},
		}))
	})
	t.Run("report error if loaded from file", func(t *testing.T) {
		memoriesFile := path.Join(t.TempDir(), "foo")
		cliOpts := sazed.CLIOptions{ MemoriesFile: memoriesFile }

		msg := sazed.InitLoadMemories(cliOpts)()

		assert.ErrorContains(t, msg.(error), "file")
	})
	t.Run("report error if invalid yaml", func(t *testing.T) {
		memoriesFile := path.Join(t.TempDir(), "foo")
		memoriesFileContent := "INV{A}LID{YAML"
		os.WriteFile(memoriesFile, []byte(memoriesFileContent), 0644)
		cliOpts := sazed.CLIOptions{ MemoriesFile: memoriesFile }

		msg := sazed.InitLoadMemories(cliOpts)()

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
		model := sazed.InitialModel(sazed.CLIOptions{})
		model.Memories = []sazed.Memory{memory1()}
		rendered := strings.Split(model.View(), "\n")
		assert.Equal(t, "Please select a command", rendered[0])
		assert.Contains(t, rendered[2], "cmd1")
		assert.Contains(t, rendered[2], "Memory 1")
	})
	t.Run("renders an input field", func(t *testing.T) {
		model := sazed.InitialModel(sazed.CLIOptions{})
		model.TextInput, _ = model.TextInput.Update(tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{'a'},
		})
		rendered := strings.Split(model.View(), "\n")
		assert.Equal(t, "> a\x1b[7m \x1b[0m", rendered[1])
	})
}

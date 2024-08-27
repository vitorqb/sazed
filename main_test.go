package main_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	sazed "github.com/vitorqb/sazed"
)

func Test__ParseCliArgs(t *testing.T) {
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

func Test__LoadMemoriesFromFile(t *testing.T) {
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

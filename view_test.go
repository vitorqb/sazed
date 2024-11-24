package main_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	sazed "github.com/vitorqb/sazed"
)

func TestViewCommandEdit(t *testing.T) {
	t.Run("Renders command on the first line", func(t *testing.T) {
		model := sazed.InitialModel(sazed.AppOptions{})
		model.Matches = []sazed.Match{
			{
				Memory: sazed.Memory{Command: "foo {{bar}} baz"},
			},
		}
		view := sazed.ViewCommandEdit(model)
		lines := strings.Split(view, "\n")
		assert.Equal(t, "Command: foo  baz", lines[0])
	})
	t.Run("Renders command on the first line (multiple placeholders)", func(t *testing.T) {
		model := sazed.InitialModel(sazed.AppOptions{})
		model.Matches = []sazed.Match{
			{
				Memory: sazed.Memory{Command: "foo {{bar}} {{baz}} boz"},
			},
		}
		view := sazed.ViewCommandEdit(model)
		lines := strings.Split(view, "\n")
		assert.Equal(t, "Command: foo   boz", lines[0])
	})
	t.Run("Replaces placeholders for user input", func(t *testing.T) {
		model := sazed.InitialModel(sazed.AppOptions{})
		model.Matches = []sazed.Match{
			{
				Memory: sazed.Memory{Command: "foo {{bar}} baz {{boz}}"},
			},
		}
		model.PlaceholderValues = []string{"--opt1", "--opt2"}
		view := sazed.ViewCommandEdit(model)
		lines := strings.Split(view, "\n")
		assert.Equal(t, "Command: foo --opt1 baz --opt2", lines[0])
	})
}

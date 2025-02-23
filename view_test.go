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
		model.SelectedMemory = sazed.Memory{Command: "foo {{bar}} baz"}
		view := sazed.ViewCommandEdit(model)
		lines := strings.Split(view, "\n")
		assert.Equal(t, "Command: foo {{bar}} baz", lines[0])
	})
	t.Run("Renders command on the first line (multiple placeholders)", func(t *testing.T) {
		model := sazed.InitialModel(sazed.AppOptions{})
		model.SelectedMemory = memory5()
		view := sazed.ViewCommandEdit(model)
		lines := strings.Split(view, "\n")
		assert.Equal(t, "Command: echo {{value1}} {{value2}} end", lines[0])
	})
	t.Run("Replaces placeholders for user input", func(t *testing.T) {
		model := sazed.InitialModel(sazed.AppOptions{})
		model.SelectedMemory = memory5()
		model = sazed.SetupEditTextInputs(model)
		for i, value := range []string{"--opt1", "--opt2"} {
			model.EditTextInputs[i].SetValue(value)
		}
		view := sazed.ViewCommandEdit(model)
		lines := strings.Split(view, "\n")
		assert.Equal(t, "Command: echo --opt1 --opt2 end", lines[0])
	})
	t.Run("Renders input for each placeholder", func(t *testing.T) {
		memory := sazed.Memory{Command: "foo {{bar}} baz {{boz}}"}
		model := sazed.InitialModel(sazed.AppOptions{})
		model.SelectedMemory = memory
		model = sazed.SetupEditTextInputs(model)
		model.EditTextInputs[0].SetValue("--opt1")
		model.EditTextInputs[1].SetValue("--opt2")

		view := sazed.ViewCommandEdit(model)
		lines := strings.Split(view, "\n")
		assert.Equal(t, "Command: foo --opt1 baz --opt2", lines[0])
		assert.Equal(t, "bar: --opt1 ", lines[1])
		assert.Equal(t, "boz: --opt2 ", lines[2])
	})
}

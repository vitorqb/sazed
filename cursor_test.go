package main_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	sazed "github.com/vitorqb/sazed"
)

func TestIncreaseCursor(t *testing.T) {
	memories := []sazed.Memory{
		sazed.Memory{Command: "cmd1", Description: "Memory 1"},
		sazed.Memory{Command: "foo", Description: "Bar"},
		sazed.Memory{Command: "not foo", Description: "not bar"},
	}

	model := sazed.InitialModel(sazed.AppOptions{})
	model.Memories = memories
	model = sazed.IncreaseCursor(model)
	model = sazed.IncreaseCursor(model)

	assert.Equal(t, model.Cursor, 2)

	model = sazed.IncreaseCursor(model)
	model = sazed.IncreaseCursor(model)

	assert.Equal(t, model.Cursor, 2) // Only have 3 memories

	model = sazed.DecreaseCursor(model)
	model = sazed.DecreaseCursor(model)

	assert.Equal(t, model.Cursor, 0)

	model = sazed.DecreaseCursor(model)
	model = sazed.DecreaseCursor(model)

	assert.Equal(t, model.Cursor, 0) // Can't go below 1
}

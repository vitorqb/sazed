package main_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	sazed "github.com/vitorqb/sazed"
)

func TestIncreaseMatchCursor(t *testing.T) {
	memories := []sazed.Memory{
		sazed.Memory{Command: "cmd1", Description: "Memory 1"},
		sazed.Memory{Command: "foo", Description: "Bar"},
		sazed.Memory{Command: "not foo", Description: "not bar"},
	}

	model := sazed.InitialModel(sazed.AppOptions{})
	model.Memories = memories
	model = sazed.IncreaseMatchCursor(model)
	model = sazed.IncreaseMatchCursor(model)

	assert.Equal(t, model.MatchCursor, 2)

	model = sazed.IncreaseMatchCursor(model)
	model = sazed.IncreaseMatchCursor(model)

	assert.Equal(t, model.MatchCursor, 2) // Only have 3 memories

	model = sazed.DecreaseMatchCursor(model)
	model = sazed.DecreaseMatchCursor(model)

	assert.Equal(t, model.MatchCursor, 0)

	model = sazed.DecreaseMatchCursor(model)
	model = sazed.DecreaseMatchCursor(model)

	assert.Equal(t, model.MatchCursor, 0) // Can't go below 1
}

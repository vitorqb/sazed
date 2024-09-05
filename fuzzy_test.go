package main_test

import (
	"testing"
	"github.com/stretchr/testify/assert"
	sazed "github.com/vitorqb/sazed"
)

func Test__Fuzzy__SortByMatch(t *testing.T) {
	t.Run("null value", func(t *testing.T) {
		var memories []sazed.Memory
		sazed.NewFuzzy().SortByMatch(memories, "foo")
		var emptyMemories []sazed.Memory
		assert.Equal(t, emptyMemories, memories)
	})
	t.Run("empty", func(t *testing.T) {
		memories := []sazed.Memory{}
		sazed.NewFuzzy().SortByMatch(memories, "foo")
		assert.Equal(t, []sazed.Memory{}, memories)
	})
	t.Run("sorts by match", func(t *testing.T) {
		memories := []sazed.Memory{
			{Description: "foo"},
			{Description: "bar"},
			{Description: "not foo"},
			{Description: "very long description with foo"},
		}
		expected := []sazed.Memory{
			memories[0],
			memories[2],
			memories[3],
			memories[1],
		}
		sazed.NewFuzzy().SortByMatch(memories, "foo")
		assert.Equal(t, expected, memories)
	})
}

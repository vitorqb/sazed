package main_test

import (
	"github.com/stretchr/testify/assert"
	sazed "github.com/vitorqb/sazed"
	"testing"
)

func TestGetMatches(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		memories := []sazed.Memory{}
		matches := sazed.NewFuzzy().GetMatches(memories, "foo")
		assert.Equal(t, []sazed.Match{}, matches)
	})
	t.Run("two in one out", func(t *testing.T) {
		memories := []sazed.Memory{
			{Command: "foo"},
			{Command: "bar"},
		}
		matches := sazed.NewFuzzy().GetMatches(memories, "foo")
		assert.Equal(t, []sazed.Match{
			{
				Memory:                memories[0],
				Score:                 30,
				CommandMatchedIndexes: []int{0, 1, 2},
			},
		}, matches)
	})
	t.Run("three in two out", func(t *testing.T) {
		memories := []sazed.Memory{
			{Command: "foo"},
			{Command: "bar"},
			{Command: "not foo"},
		}
		matches := sazed.NewFuzzy().GetMatches(memories, "foo")
		assert.Equal(t, []sazed.Match{
			{
				Memory:                memories[0],
				Score:                 30,
				CommandMatchedIndexes: []int{0, 1, 2},
			},
			{
				Memory:                memories[2],
				Score:                 21,
				CommandMatchedIndexes: []int{4, 5, 6},
			},
		}, matches)
	})
	t.Run("match by command and description", func(t *testing.T) {
		memories := []sazed.Memory{
			{Command: "foo", Description: "bar"},
			{Command: "bar", Description: "foo2"},
		}
		matches := sazed.NewFuzzy().GetMatches(memories, "foo")
		assert.Equal(t, []sazed.Match{
			{
				Memory:                memories[0],
				Score:                 30,
				CommandMatchedIndexes: []int{0, 1, 2},
			},
			{
				Memory:                    memories[1],
				Score:                     29,
				DescriptionMatchedIndexes: []int{0, 1, 2},
			},
		}, matches)
	})
	t.Run("if input str is empty all memories are returned", func(t *testing.T) {
		memories := []sazed.Memory{
			{Command: "foo", Description: "bar"},
			{Command: "bar", Description: "foo2"},
		}
		matches := sazed.NewFuzzy().GetMatches(memories, "")
		assert.Equal(t, []sazed.Match{
			{Memory: memories[0]},
			{Memory: memories[1]},
		}, matches)
	})
}

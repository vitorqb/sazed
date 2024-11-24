package main_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	sazed "github.com/vitorqb/sazed"
)

func Test__countPlaceholders(t *testing.T) {
	assert.Equal(t, 0, sazed.CountPlaceholders("foo"))
	assert.Equal(t, 1, sazed.CountPlaceholders("foo {{bar}} baz"))
	assert.Equal(t, 2, sazed.CountPlaceholders("foo {{bar}} {{baz}}"))
	assert.Equal(t, 3, sazed.CountPlaceholders("{{foo}} {{bar}} {{baz}}"))
	assert.Equal(t, 2, sazed.CountPlaceholders("{{foo}} bar {{baz}}"))
	assert.Equal(t, 1, sazed.CountPlaceholders("{{foo}} bar {baz}}"))
	assert.Equal(t, 1, sazed.CountPlaceholders("}} bar {{baz}}"))
	assert.Equal(t, 1, sazed.CountPlaceholders("{{ bar {{baz}}"))
	assert.Equal(t, 1, sazed.CountPlaceholders("{ bar {{baz}}}"))
	assert.Equal(t, 0, sazed.CountPlaceholders("{foo} bar {baz}}"))
}

func Test__getPlaceholdersIndexes(t *testing.T) {
	assert.Equal(t, []sazed.Placeholder{}, sazed.GetPlaceholders("foo"))
	assert.Equal(t, []sazed.Placeholder{{4, 10, "bar"}}, sazed.GetPlaceholders("foo {{bar}} baz"))
	assert.Equal(t, []sazed.Placeholder{{4, 10, "bar"}, {12, 18, "baz"}}, sazed.GetPlaceholders("foo {{bar}} {{baz}}"))
	assert.Equal(t, []sazed.Placeholder{{0, 6, "foo"}, {8, 14, "bar"}, {16, 22, "baz"}}, sazed.GetPlaceholders("{{foo}} {{bar}} {{baz}}"))
	assert.Equal(t, []sazed.Placeholder{{0, 6, "foo"}, {12, 18, "baz"}}, sazed.GetPlaceholders("{{foo}} bar {{baz}}"))
	assert.Equal(t, []sazed.Placeholder{{0, 6, "foo"}}, sazed.GetPlaceholders("{{foo}} bar {baz}}"))
	assert.Equal(t, []sazed.Placeholder{{7, 13, "baz"}}, sazed.GetPlaceholders("}} bar {{baz}}"))
	assert.Equal(t, []sazed.Placeholder{{0, 13, " bar {{baz"}}, sazed.GetPlaceholders("{{ bar {{baz}}"))
	assert.Equal(t, []sazed.Placeholder{{6, 12, "baz"}}, sazed.GetPlaceholders("{ bar {{baz}}}"))
	assert.Equal(t, []sazed.Placeholder{}, sazed.GetPlaceholders("{foo} bar {baz}}"))
}

func Test_NextPlaceholder(t *testing.T) {
	td := []struct {
		original string
		expected sazed.Placeholder
	}{
		{
			original: "echo {{foo}}",
			expected: sazed.Placeholder{5, 11, "foo"},
		},
		{
			original: "echo {{foo}} bar",
			expected: sazed.Placeholder{5, 11, "foo"},
		},
		{
			original: "{{foo}} bar baz",
			expected: sazed.Placeholder{0, 6, "foo"},
		},
		{
			original: "foo bar baz",
			expected: sazed.Placeholder{},
		},
	}
	for i, tc := range td {
		t.Run(fmt.Sprintf("%s [%d]", tc.original, i), func(t *testing.T) {
			placeholder, success := sazed.NextPlaceholder(tc.original)
			assert.Equal(t, tc.expected, placeholder)
			assert.Equal(t, tc.expected != sazed.Placeholder{}, success)
		})
	}
}

func Test__ReplacePlaceholder(t *testing.T) {
	td := []struct {
		original    string
		placeholder sazed.Placeholder
		replacement string
		expected    string
	}{
		{
			original:    "echo {{foo}}",
			placeholder: sazed.Placeholder{5, 11, ""},
			replacement: "baz",
			expected:    "echo baz",
		},
		{
			original:    "echo {{foo}} bar",
			placeholder: sazed.Placeholder{5, 11, ""},
			replacement: "baz",
			expected:    "echo baz bar",
		},
		{
			original:    "{{foo}} bar baz",
			placeholder: sazed.Placeholder{0, 6, ""},
			replacement: "foo",
			expected:    "foo bar baz",
		},
	}
	for i, tc := range td {
		t.Run(fmt.Sprintf("%s [%d]", tc.original, i), func(t *testing.T) {
			assert.Equal(t, tc.expected, sazed.ReplacePlaceholder(tc.original, tc.placeholder, tc.replacement))
		})
	}
}

func Test__Render(t *testing.T) {
	td := []struct {
		original          string
		placeholderValues []string
		expected          string
	}{
		{
			original:          "echo {{foo}}",
			placeholderValues: []string{"bar"},
			expected:          "echo bar",
		},
		{
			original:          "echo {{foo}} bar",
			placeholderValues: []string{"baz"},
			expected:          "echo baz bar",
		},
		{
			original:          "{{foo}} bar baz",
			placeholderValues: []string{},
			expected:          " bar baz",
		},
		{
			original:          "{{foo}} bar {{baz}}",
			placeholderValues: []string{"baz", "foo", "boz"},
			expected:          "baz bar foo",
		},
	}
	for i, tc := range td {
		t.Run(fmt.Sprintf("%s [%d]", tc.original, i), func(t *testing.T) {
			assert.Equal(t, tc.expected, sazed.Render(tc.original, tc.placeholderValues))
		})
	}
}

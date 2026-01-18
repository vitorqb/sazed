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

func Test__GetPlaceholders(t *testing.T) {
	assert.Equal(t, []sazed.Placeholder{}, sazed.GetPlaceholders("foo"))
	assert.Equal(t, []sazed.Placeholder{{"bar", "{{bar}}"}}, sazed.GetPlaceholders("foo {{bar}} baz"))
	assert.Equal(t, []sazed.Placeholder{{"bar", "{{bar}}"}, {"baz", "{{baz}}"}}, sazed.GetPlaceholders("foo {{bar}} {{baz}}"))
	assert.Equal(t, []sazed.Placeholder{{"foo", "{{foo}}"}, {"bar", "{{bar}}"}, {"baz", "{{baz}}"}}, sazed.GetPlaceholders("{{foo}} {{bar}} {{baz}}"))
	assert.Equal(t, []sazed.Placeholder{{"foo", "{{foo}}"}, {"baz", "{{baz}}"}}, sazed.GetPlaceholders("{{foo}} bar {{baz}}"))
	assert.Equal(t, []sazed.Placeholder{{"foo", "{{foo}}"}}, sazed.GetPlaceholders("{{foo}} bar {baz}}"))
	assert.Equal(t, []sazed.Placeholder{{"baz", "{{baz}}"}}, sazed.GetPlaceholders("}} bar {{baz}}"))
	assert.Equal(t, []sazed.Placeholder{{" bar {{baz", "{{ bar {{baz}}"}}, sazed.GetPlaceholders("{{ bar {{baz}}"))
	assert.Equal(t, []sazed.Placeholder{{"baz", "{{baz}}"}}, sazed.GetPlaceholders("{ bar {{baz}}}"))
	assert.Equal(t, []sazed.Placeholder{}, sazed.GetPlaceholders("{foo} bar {baz}}"))
}

func Test__ReplacePlaceholder(t *testing.T) {
	td := []struct {
		original    string
		placeholder sazed.Placeholder
		replacement string
		opts        sazed.RenderOpts
		expected    string
	}{
		{
			original:    "echo {{foo}}",
			placeholder: sazed.Placeholder{"foo", "{{foo}}"},
			replacement: "baz",
			opts:        sazed.RenderOpts{},
			expected:    "echo baz",
		},
		{
			original:    "echo {{foo}} bar",
			placeholder: sazed.Placeholder{"foo", "{{foo}}"},
			replacement: "baz",
			opts:        sazed.RenderOpts{},
			expected:    "echo baz bar",
		},
		{
			original:    "{{foo}} bar baz",
			placeholder: sazed.Placeholder{"foo", "{{foo}}"},
			replacement: "foo",
			opts:        sazed.RenderOpts{},
			expected:    "foo bar baz",
		},
		{
			original:    "foo {{bar}} baz",
			placeholder: sazed.Placeholder{"bar", "{{bar}}"},
			replacement: "",
			opts:        sazed.RenderOpts{Optional: true},
			expected:    "foo baz",
		},
		{
			original:    "foo{{bar}} baz",
			placeholder: sazed.Placeholder{"bar", "{{bar}}"},
			replacement: "",
			opts:        sazed.RenderOpts{Optional: true},
			expected:    "foo baz",
		},
		{
			original:    "foo {{bar}} baz",
			placeholder: sazed.Placeholder{"bar", "{{bar}}"},
			replacement: "",
			opts:        sazed.RenderOpts{Optional: false},
			expected:    "foo {{bar}} baz",
		},
		{
			original:    "foo {{bar}} {{baz}}",
			placeholder: sazed.Placeholder{"bar", "{{bar}}"},
			replacement: "val",
			opts:        sazed.RenderOpts{Prefix: "--foo="},
			expected:    "foo --foo=val {{baz}}",
		},
		{
			original:    "foo {{bar}} {{baz}}",
			placeholder: sazed.Placeholder{"baz", "{{baz}}"},
			replacement: "val",
			opts:        sazed.RenderOpts{Prefix: "--foo="},
			expected:    "foo {{bar}} --foo=val",
		},
		{
			original:    "foo {{bar}} {{baz}}",
			placeholder: sazed.Placeholder{"bar", "{{bar}}"},
			replacement: "val",
			opts:        sazed.RenderOpts{Prefix: "--foo=", Optional: true},
			expected:    "foo --foo=val {{baz}}",
		},
		{
			original:    "foo {{baz}}",
			placeholder: sazed.Placeholder{"baz", "{{baz}}"},
			replacement: "val",
			opts:        sazed.RenderOpts{Prefix: "--foo=", Optional: true},
			expected:    "foo --foo=val",
		},
	}
	for i, tc := range td {
		t.Run(fmt.Sprintf("%s [%d]", tc.original, i), func(t *testing.T) {
			assert.Equal(t, tc.expected, sazed.ReplacePlaceholder(tc.original, tc.placeholder, tc.replacement, tc.opts))
		})
	}
}

func Test__Render(t *testing.T) {
	td := []struct {
		original   string
		userInputs []string
		renderOpts []sazed.RenderOpts
		expected   string
	}{
		{
			original:          "echo {{foo}}",
			userInputs:        []string{"bar"},
			renderOpts:        []sazed.RenderOpts{{}},
			expected:          "echo bar",
		},
		{
			original:          "echo {{foo}} bar",
			userInputs: []string{"baz"},
			renderOpts:        []sazed.RenderOpts{{}},
			expected:          "echo baz bar",
		},
		{
			original:          "{{foo}} bar baz",
			userInputs:        []string{},
			renderOpts:        []sazed.RenderOpts{{}},
			expected:          "{{foo}} bar baz",
		},
		{
			original:          "{{foo}} bar baz",
			userInputs:        []string{},
			renderOpts:        []sazed.RenderOpts{{Optional: true}},
			expected:          " bar baz",
		},
		{
			original:          "{{foo}} bar {{baz}}",
			userInputs: []string{"baz", "foo", "boz"},
			renderOpts:        []sazed.RenderOpts{{}},
			expected:          "baz bar foo",
		},
		{
			original:          "grep {{f-exclude-dir}} foo .",
			userInputs: []string{""},
			renderOpts:        []sazed.RenderOpts{{Optional: true, Prefix: "--exclude-dir="}},
			expected:          "grep foo .",
		},
		{
			original:          "grep {{f-exclude-dir}} foo .",
			userInputs: []string{"baz"},
			renderOpts:        []sazed.RenderOpts{{Optional: true, Prefix: "--exclude-dir="}},
			expected:          "grep --exclude-dir=baz foo .",
		},
		{
			original:          "grep {{what}} .",
			userInputs: []string{""},
			renderOpts:        []sazed.RenderOpts{{Optional: false, Prefix: ""}},
			expected:          "grep {{what}} .",
		},
		{
			original:          "grep {{what}} .",
			userInputs: []string{"foo"},
			renderOpts:        []sazed.RenderOpts{{Optional: false, Prefix: ""}},
			expected:          "grep foo .",
		},
		
	}
	for i, tc := range td {
		t.Run(fmt.Sprintf("%s [%d]", tc.original, i), func(t *testing.T) {
			assert.Equal(t, tc.expected, sazed.Render(tc.original, tc.userInputs, tc.renderOpts))
		})
	}
}

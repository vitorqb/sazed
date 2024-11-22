package main_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	sazed "github.com/vitorqb/sazed"
)

func Test__countPlaceholders(t *testing.T) {
	assert.Equal(t, 0, sazed.CountPlaceholders(sazed.Memory{Command: "foo"}))
	assert.Equal(t, 1, sazed.CountPlaceholders(sazed.Memory{Command: "foo {{bar}} baz"}))
	assert.Equal(t, 2, sazed.CountPlaceholders(sazed.Memory{Command: "foo {{bar}} {{baz}}"}))
	assert.Equal(t, 3, sazed.CountPlaceholders(sazed.Memory{Command: "{{foo}} {{bar}} {{baz}}"}))
	assert.Equal(t, 2, sazed.CountPlaceholders(sazed.Memory{Command: "{{foo}} bar {{baz}}"}))
	assert.Equal(t, 1, sazed.CountPlaceholders(sazed.Memory{Command: "{{foo}} bar {baz}}"}))
	assert.Equal(t, 1, sazed.CountPlaceholders(sazed.Memory{Command: "}} bar {{baz}}"}))
	assert.Equal(t, 1, sazed.CountPlaceholders(sazed.Memory{Command: "{{ bar {{baz}}"}))
	assert.Equal(t, 1, sazed.CountPlaceholders(sazed.Memory{Command: "{ bar {{baz}}}"}))
	assert.Equal(t, 0, sazed.CountPlaceholders(sazed.Memory{Command: "{foo} bar {baz}}"}))
}

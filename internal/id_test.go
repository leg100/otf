package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewID(t *testing.T) {
	got := NewID("run")
	t.Log(got)
	assert.Regexp(t, `run-[0-9a-bA-Z]{22}`, got)
}

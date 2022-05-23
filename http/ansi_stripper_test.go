package http

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnsiStripper(t *testing.T) {
	input := "\x1b[38;5;140m foo\x1b[0m bar\n\x1b[38;5;140m foo\x1b[0m bar\n\x1b[38;5;140m foo\x1b[0m bar"
	want := " foo bar\n foo bar\n foo bar\n"
	stripper := NewAnsiStripper(strings.NewReader(input))
	got, err := io.ReadAll(stripper)
	require.NoError(t, err)
	assert.Equal(t, want, string(got))
}

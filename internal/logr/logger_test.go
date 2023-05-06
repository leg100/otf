package logr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToSlogLevel(t *testing.T) {
	tests := []struct {
		name string
		v    int
		want string
	}{
		{"info", 0, "INFO"},
		{"debug", 1, "DEBUG"},
		{"debug-1", 2, "DEBUG-1"},
		{"debug-2", 3, "DEBUG-2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toSlogLevel(tt.v)
			assert.Equal(t, tt.want, got.String())
		})
	}
}

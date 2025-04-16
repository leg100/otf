package user

import (
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/stretchr/testify/assert"
)

func TestNewUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		error    error
	}{
		{"empty", "", internal.ErrInvalidName},
		{"normal username", "bob", nil},
		{"normal username as well", "anne", nil},
		{"email", "daniel@canadianstartup.com", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewUsername(tt.username)
			assert.ErrorIs(t, err, tt.error)
		})
	}
}

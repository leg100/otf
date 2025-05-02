package workspace

import (
	"testing"

	"github.com/leg100/otf/internal/engine"
	"github.com/stretchr/testify/assert"
)

func TestVersion(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError error
	}{
		{"valid semver", "1.9.3", nil},
		{"track latest version", "latest", nil},
		{"invalid semver", "1,2,0", engine.ErrInvalidVersion},
		{"unsupported terraform version", "0.14.0", ErrUnsupportedTerraformVersion},
		{"make exception for go-tfe integration test version", "0.10.0", nil},
		{"make exception for go-tfe integration test version", "0.11.0", nil},
		{"make exception for go-tfe integration test version", "0.11.1", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Version{}
			got := v.set(tt.input)
			assert.Equal(t, tt.wantError, got)
		})
	}
}

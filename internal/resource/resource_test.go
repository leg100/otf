package resource

import (
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/stretchr/testify/assert"
)

func TestValidateName(t *testing.T) {
	tests := []struct {
		name         string
		resourceName *string
		want         error
	}{
		{"nil", nil, internal.ErrRequiredName},
		{"dot", internal.Ptr("."), internal.ErrInvalidName},
		{"underscore", internal.Ptr("_"), nil},
		{"acme-corp", internal.Ptr("acme-corp"), nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateName(tt.resourceName)
			assert.Equal(t, tt.want, err)
		})
	}
}

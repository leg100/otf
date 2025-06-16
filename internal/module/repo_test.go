package module

import (
	"testing"

	"github.com/leg100/otf/internal/vcs"
	"github.com/stretchr/testify/assert"
)

func TestRepo_Split(t *testing.T) {
	tests := []struct {
		repo         vcs.Repo
		wantName     string
		wantProvider string
		wantError    error
	}{
		{vcs.NewMustRepo("leg100", "terraform-aws-vpc"), "vpc", "aws", nil},
		{vcs.NewMustRepo("leg100", "/anything-aws-vpc"), "vpc", "aws", nil},
		{vcs.NewMustRepo("leg100", "/terraform-gcp-secrets-manager"), "secrets-manager", "gcp", nil},
		{vcs.NewMustRepo("leg100", "/not_a_module_repo"), "", "", ErrInvalidModuleRepo},
	}
	for _, tt := range tests {
		t.Run(tt.repo.String(), func(t *testing.T) {
			gotName, gotProvider, gotError := Repo(tt.repo).Split()
			assert.Equal(t, tt.wantName, gotName)
			assert.Equal(t, tt.wantProvider, gotProvider)
			assert.Equal(t, tt.wantError, gotError)
		})
	}
}

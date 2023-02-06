package workspace

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewWorkspace(t *testing.T) {
	tests := []struct {
		name string
		opts CreateWorkspaceOptions
		want error
	}{
		{
			name: "default",
			opts: CreateWorkspaceOptions{
				Name:         String("my-workspace"),
				Organization: String("my-org"),
			},
		},
		{
			name: "missing name",
			opts: CreateWorkspaceOptions{
				Organization: String("my-org"),
			},
			want: ErrRequiredName,
		},
		{
			name: "missing organization",
			opts: CreateWorkspaceOptions{
				Name: String("my-workspace"),
			},
			want: ErrRequiredOrg,
		},
		{
			name: "invalid name",
			opts: CreateWorkspaceOptions{
				Name:         String("%*&^"),
				Organization: String("my-org"),
			},
			want: ErrInvalidName,
		},
		{
			name: "bad terraform version",
			opts: CreateWorkspaceOptions{
				Name:             String("my-workspace"),
				Organization:     String("my-org"),
				TerraformVersion: String("1,2,0"),
			},
			want: ErrInvalidTerraformVersion,
		},
		{
			name: "unsupported terraform version",
			opts: CreateWorkspaceOptions{
				Name:             String("my-workspace"),
				Organization:     String("my-org"),
				TerraformVersion: String("0.14.0"),
			},
			want: ErrUnsupportedTerraformVersion,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, got := NewWorkspace(tt.opts)
			assert.Equal(t, tt.want, got)
		})
	}
}

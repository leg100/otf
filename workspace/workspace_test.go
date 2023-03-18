package workspace

import (
	"testing"

	"github.com/leg100/otf"
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
				Name:         otf.String("my-workspace"),
				Organization: otf.String("my-org"),
			},
		},
		{
			name: "missing name",
			opts: CreateWorkspaceOptions{
				Organization: otf.String("my-org"),
			},
			want: otf.ErrRequiredName,
		},
		{
			name: "missing organization",
			opts: CreateWorkspaceOptions{
				Name: otf.String("my-workspace"),
			},
			want: otf.ErrRequiredOrg,
		},
		{
			name: "invalid name",
			opts: CreateWorkspaceOptions{
				Name:         otf.String("%*&^"),
				Organization: otf.String("my-org"),
			},
			want: otf.ErrInvalidName,
		},
		{
			name: "bad terraform version",
			opts: CreateWorkspaceOptions{
				Name:             otf.String("my-workspace"),
				Organization:     otf.String("my-org"),
				TerraformVersion: otf.String("1,2,0"),
			},
			want: otf.ErrInvalidTerraformVersion,
		},
		{
			name: "unsupported terraform version",
			opts: CreateWorkspaceOptions{
				Name:             otf.String("my-workspace"),
				Organization:     otf.String("my-org"),
				TerraformVersion: otf.String("0.14.0"),
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

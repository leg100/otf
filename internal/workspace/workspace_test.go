package workspace

import (
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/stretchr/testify/assert"
)

func TestNewWorkspace(t *testing.T) {
	tests := []struct {
		name string
		opts CreateOptions
		want error
	}{
		{
			name: "default",
			opts: CreateOptions{
				Name:         internal.String("my-workspace"),
				Organization: internal.String("my-org"),
			},
		},
		{
			name: "missing name",
			opts: CreateOptions{
				Organization: internal.String("my-org"),
			},
			want: internal.ErrRequiredName,
		},
		{
			name: "missing organization",
			opts: CreateOptions{
				Name: internal.String("my-workspace"),
			},
			want: internal.ErrRequiredOrg,
		},
		{
			name: "invalid name",
			opts: CreateOptions{
				Name:         internal.String("%*&^"),
				Organization: internal.String("my-org"),
			},
			want: internal.ErrInvalidName,
		},
		{
			name: "bad terraform version",
			opts: CreateOptions{
				Name:             internal.String("my-workspace"),
				Organization:     internal.String("my-org"),
				TerraformVersion: internal.String("1,2,0"),
			},
			want: internal.ErrInvalidTerraformVersion,
		},
		{
			name: "unsupported terraform version",
			opts: CreateOptions{
				Name:             internal.String("my-workspace"),
				Organization:     internal.String("my-org"),
				TerraformVersion: internal.String("0.14.0"),
			},
			want: internal.ErrUnsupportedTerraformVersion,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, got := NewWorkspace(tt.opts)
			assert.Equal(t, tt.want, got)
		})
	}
}

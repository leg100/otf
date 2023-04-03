package workspace

import (
	"testing"

	"github.com/leg100/otf"
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
				Name:         otf.String("my-workspace"),
				Organization: otf.String("my-org"),
			},
		},
		{
			name: "missing name",
			opts: CreateOptions{
				Organization: otf.String("my-org"),
			},
			want: otf.ErrRequiredName,
		},
		{
			name: "missing organization",
			opts: CreateOptions{
				Name: otf.String("my-workspace"),
			},
			want: otf.ErrRequiredOrg,
		},
		{
			name: "invalid name",
			opts: CreateOptions{
				Name:         otf.String("%*&^"),
				Organization: otf.String("my-org"),
			},
			want: otf.ErrInvalidName,
		},
		{
			name: "bad terraform version",
			opts: CreateOptions{
				Name:             otf.String("my-workspace"),
				Organization:     otf.String("my-org"),
				TerraformVersion: otf.String("1,2,0"),
			},
			want: otf.ErrInvalidTerraformVersion,
		},
		{
			name: "unsupported terraform version",
			opts: CreateOptions{
				Name:             otf.String("my-workspace"),
				Organization:     otf.String("my-org"),
				TerraformVersion: otf.String("0.14.0"),
			},
			want: otf.ErrUnsupportedTerraformVersion,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, got := NewWorkspace(tt.opts)
			assert.Equal(t, tt.want, got)
		})
	}
}

package module

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetLatest(t *testing.T) {
	tests := []struct {
		name        string
		current     *ModuleVersion   // current latest
		versions    []*ModuleVersion // module's full list of versions
		want        *ModuleVersion   // want this to be new latest version
		wantChanged bool             // want latest version to have changed
	}{
		{
			name: "no versions",
			want: nil,
		},
		{
			name: "one ok version",
			versions: []*ModuleVersion{
				{ID: "want", Status: ModuleVersionStatusOK},
			},
			want: &ModuleVersion{ID: "want", Status: ModuleVersionStatusOK},
		},
		{
			name:    "ignore newer pending version",
			current: &ModuleVersion{ID: "want", Status: ModuleVersionStatusOK},
			versions: []*ModuleVersion{
				{Version: "v1", Status: ModuleVersionStatusOK},
				{Version: "v2", Status: ModuleVersionStatusPending},
			},
			want: &ModuleVersion{Version: "v1", Status: ModuleVersionStatusOK},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mod := &Module{Versions: make(map[string]*ModuleVersion)}
			for _, mv := range tt.versions {
				mod.Versions[mv.Version] = mv
			}
			assert.Equal(t, tt.wantChanged, mod.SetLatest())
			assert.Equal(t, tt.want, mod.Latest)
		})
	}
}

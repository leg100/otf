package ui

import (
	"context"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/connections"
	"github.com/leg100/otf/internal/module"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListModules(t *testing.T) {
	h := newTestWebHandlers(t, withMod(&module.Module{}))
	user := &user.User{ID: resource.NewTfeID(resource.UserKind)}

	q := "/?organization_name=acme-corp"
	r := httptest.NewRequest("GET", q, nil)
	r = r.WithContext(authz.AddSubjectToContext(r.Context(), user))
	w := httptest.NewRecorder()
	h.list(w, r)
	assert.Equal(t, 200, w.Code, w.Body.String())
}

func TestGetModule(t *testing.T) {
	tarball, err := os.ReadFile("./testdata/module.tar.gz")
	require.NoError(t, err)

	tests := []struct {
		name string
		mod  module.Module
	}{
		{
			name: "pending",
			mod: module.Module{
				Status: module.ModuleStatusPending,
			},
		},
		{
			name: "no versions",
			mod: module.Module{
				Status: module.ModuleStatusNoVersionTags,
			},
		},
		{
			name: "setup failed",
			mod: module.Module{
				Status: module.ModuleStatusSetupFailed,
			},
		},
		{
			name: "setup complete",
			mod: module.Module{
				Connection: &connections.Connection{},
				Status:     module.ModuleStatusSetupComplete,
				Versions:   []module.ModuleVersion{{Version: "1.0.0"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &moduleHandlers{
				client: &fakeModuleService{mod: &tt.mod},
			}

			q := "/?module_id=mod-123&version=1.0.0"
			r := httptest.NewRequest("GET", q, nil)
			w := httptest.NewRecorder()
			h.get(w, r)
			assert.Equal(t, 200, w.Code, w.Body.String())
		})
	}
}

type fakeModuleService struct {
	mod     *module.Module
	tarball []byte
	moduleClient
}

func (f *fakeModuleService) GetModuleByID(context.Context, resource.TfeID) (*module.Module, error) {
	return f.mod, nil
}

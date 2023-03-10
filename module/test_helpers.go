package module

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/require"
)

func NewTestModule(org *otf.Organization, opts ...NewTestModuleOption) *otf.Module {
	createOpts := otf.CreateModuleOptions{
		Organization: org.Name,
		Provider:     uuid.NewString(),
		Name:         uuid.NewString(),
	}
	mod := otf.NewModule(createOpts)
	for _, o := range opts {
		o(mod)
	}
	return mod
}

type NewTestModuleOption func(*otf.Module)

func WithModuleStatus(status otf.ModuleStatus) NewTestModuleOption {
	return func(mod *otf.Module) {
		mod.Status = status
	}
}

func WithModuleRepo() NewTestModuleOption {
	return func(mod *otf.Module) {
		mod.Connection = &otf.Connection{}
	}
}

func NewTestModuleVersion(mod *otf.Module, version string, status otf.ModuleVersionStatus) *otf.ModuleVersion {
	createOpts := otf.CreateModuleVersionOptions{
		ModuleID: mod.ID,
		Version:  version,
	}
	modver := otf.NewModuleVersion(createOpts)
	modver.Status = status
	return modver
}

func createTestModule(t *testing.T, db *pgdb, org *otf.Organization) *otf.Module {
	ctx := context.Background()

	module := NewTestModule(org)
	err := db.CreateModule(ctx, module)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.DeleteModule(ctx, module.ID)
	})
	return module
}

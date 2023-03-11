package module

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/require"
)

func NewTestModule(org *organization.Organization, opts ...NewTestModuleOption) *Module {
	createOpts := CreateModuleOptions{
		Organization: org.Name,
		Provider:     uuid.NewString(),
		Name:         uuid.NewString(),
	}
	mod := NewModule(createOpts)
	for _, o := range opts {
		o(mod)
	}
	return mod
}

type NewTestModuleOption func(*Module)

func WithModuleStatus(status ModuleStatus) NewTestModuleOption {
	return func(mod *Module) {
		mod.Status = status
	}
}

func WithModuleRepo() NewTestModuleOption {
	return func(mod *Module) {
		mod.Connection = &otf.Connection{}
	}
}

func NewTestModuleVersion(mod *Module, version string, status ModuleVersionStatus) *ModuleVersion {
	createOpts := CreateModuleVersionOptions{
		ModuleID: mod.ID,
		Version:  version,
	}
	modver := NewModuleVersion(createOpts)
	modver.Status = status
	return modver
}

func createTestModule(t *testing.T, db *pgdb, org *organization.Organization) *Module {
	ctx := context.Background()

	module := NewTestModule(org)
	err := db.CreateModule(ctx, module)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.delete(ctx, module.ID)
	})
	return module
}

package module

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/require"
)

func NewTestModule(org otf.Organization, opts ...NewTestModuleOption) *Module {
	createOpts := CreateModuleOptions{
		Organization: org.Name,
		Provider:     uuid.NewString(),
		Name:         uuid.NewString(),
	}
	mod := newModule(createOpts)
	for _, o := range opts {
		o(mod)
	}
	return mod
}

type NewTestModuleOption func(*Module)

func WithModuleStatus(status ModuleStatus) NewTestModuleOption {
	return func(mod *Module) {
		mod.status = status
	}
}

func WithModuleVersion(version string, status ModuleVersionStatus) NewTestModuleOption {
	return func(mod *Module) {
		mod.addNewVersion(NewTestModuleVersion(mod, version, status))
	}
}

func WithModuleRepo() NewTestModuleOption {
	return func(mod *Module) {
		mod.connection = &connection{}
	}
}

func NewTestModuleVersion(mod *Module, version string, status ModuleVersionStatus) *ModuleVersion {
	createOpts := CreateModuleVersionOptions{
		ModuleID: mod.id,
		Version:  version,
	}
	modver := NewModuleVersion(createOpts)
	modver.status = status
	return modver
}

func createTestModule(t *testing.T, db *pgdb, org *otf.Organization) *Module {
	ctx := context.Background()

	module := NewTestModule(org)
	err := db.CreateModule(ctx, module)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.DeleteModule(ctx, module.ID)
	})
	return module
}

package module

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/inmem"
	"github.com/leg100/otf/organization"
	"github.com/leg100/otf/repo"
	"github.com/stretchr/testify/require"
)

func NewTestService(t *testing.T, db otf.DB) *service {
	service := NewService(Options{
		Logger:       logr.Discard(),
		DB:           db,
		CloudService: inmem.NewTestCloudService(),
	})
	service.organization = otf.NewAllowAllAuthorizer()
	return service
}

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

func CreateTestModule(t *testing.T, db otf.DB, org *organization.Organization) *Module {
	return createTestModule(t, &pgdb{db}, org)
}

type NewTestModuleOption func(*Module)

func WithModuleStatus(status ModuleStatus) NewTestModuleOption {
	return func(mod *Module) {
		mod.Status = status
	}
}

func WithModuleRepo() NewTestModuleOption {
	return func(mod *Module) {
		mod.Connection = &repo.Connection{}
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

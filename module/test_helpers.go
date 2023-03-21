package module

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf/organization"
	"github.com/stretchr/testify/require"
)

func NewTestModule(org *organization.Organization) *Module {
	createOpts := CreateOptions{
		Organization: org.Name,
		Provider:     uuid.NewString(),
		Name:         uuid.NewString(),
	}
	return NewModule(createOpts)
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
	err := db.createModule(ctx, module)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.delete(ctx, module.ID)
	})
	return module
}

package sqlite

import (
	"testing"

	"github.com/go-logr/logr"
	"github.com/jmoiron/sqlx"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/require"
)

// testPath is the database path used for tests
var testPath = ":memory:"

type newTestStateVersionOption func(*otf.StateVersion) error

func newTestDB(t *testing.T) *sqlx.DB {
	db, err := New(logr.Discard(), testPath)
	require.NoError(t, err)

	return db
}

func newTestOrganization(id, name string) *otf.Organization {
	return &otf.Organization{
		ID:    id,
		Name:  "automatize",
		Email: "sysadmin@automatize.co.uk",
	}
}

func newTestWorkspace(id, name string, org *otf.Organization) *otf.Workspace {
	return &otf.Workspace{
		ID:           id,
		Name:         name,
		Organization: org,
	}
}

func newTestConfigurationVersion(id string, ws *otf.Workspace) *otf.ConfigurationVersion {
	return &otf.ConfigurationVersion{
		ID:        id,
		Status:    otf.ConfigurationPending,
		Workspace: ws,
	}
}

func newTestStateVersion(id string, ws *otf.Workspace, opts ...newTestStateVersionOption) *otf.StateVersion {
	sv := &otf.StateVersion{
		ID:        id,
		Workspace: ws,
	}
	for _, o := range opts {
		o(sv)
	}
	return sv
}

func appendOutput(id, name, outputType, value string, sensitive bool) newTestStateVersionOption {
	return func(sv *otf.StateVersion) error {
		sv.Outputs = append(sv.Outputs, &otf.StateVersionOutput{
			ID:        id,
			Name:      name,
			Type:      outputType,
			Value:     value,
			Sensitive: sensitive,
		})
		return nil
	}
}

func newTestRun(id string, ws *otf.Workspace, cv *otf.ConfigurationVersion) *otf.Run {
	return &otf.Run{
		ID:               id,
		Status:           otf.RunPending,
		StatusTimestamps: make(otf.TimestampMap),
		Plan: &otf.Plan{
			ID:               "plan-123",
			StatusTimestamps: make(otf.TimestampMap),
		},
		Apply: &otf.Apply{
			ID:               "apply-123",
			StatusTimestamps: make(otf.TimestampMap),
		},
		Workspace:            ws,
		ConfigurationVersion: cv,
	}
}

func createTestOrganization(t *testing.T, db *sqlx.DB, id, name string) *otf.Organization {
	odb := NewOrganizationDB(db)

	org, err := odb.Create(newTestOrganization(id, name))
	require.NoError(t, err)

	return org
}

func createTestWorkspace(t *testing.T, db *sqlx.DB, id, name string, org *otf.Organization) *otf.Workspace {
	wdb := NewWorkspaceDB(db)

	ws, err := wdb.Create(newTestWorkspace(id, name, org))
	require.NoError(t, err)

	return ws
}

func createTestConfigurationVersion(t *testing.T, db *sqlx.DB, id string, ws *otf.Workspace) *otf.ConfigurationVersion {
	cdb := NewConfigurationVersionDB(db)

	cv, err := cdb.Create(newTestConfigurationVersion(id, ws))
	require.NoError(t, err)

	return cv
}

func createTestStateVersion(t *testing.T, db *sqlx.DB, id string, ws *otf.Workspace, opts ...newTestStateVersionOption) *otf.StateVersion {
	sdb := NewStateVersionDB(db)

	sv, err := sdb.Create(newTestStateVersion(id, ws, opts...))
	require.NoError(t, err)

	return sv
}

func createTestRun(t *testing.T, db *sqlx.DB, id string, ws *otf.Workspace, cv *otf.ConfigurationVersion) *otf.Run {
	rdb := NewRunDB(db)

	run, err := rdb.Create(newTestRun(id, ws, cv))
	require.NoError(t, err)

	return run
}

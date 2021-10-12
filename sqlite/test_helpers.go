package sqlite

import (
	"testing"

	"github.com/go-logr/logr"
	"github.com/jmoiron/sqlx"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/require"
)

func newTestDB(t *testing.T) *sqlx.DB {
	db, err := New(logr.Discard(), ":memory:")
	require.NoError(t, err)

	return db
}

func newTestOrganization(id string) *otf.Organization {
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

func newTestRun(id string, ws *otf.Workspace, cv *otf.ConfigurationVersion) *otf.Run {
	return &otf.Run{
		ID:               id,
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

func createTestOrganization(t *testing.T, db *sqlx.DB, id string) *otf.Organization {
	odb := NewOrganizationDB(db)

	org, err := odb.Create(newTestOrganization(id))
	require.NoError(t, err)

	return org
}

func createTestWorkspace(t *testing.T, db *sqlx.DB, id, name string) *otf.Workspace {
	org := createTestOrganization(t, db, "org-123")

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

func createTestRun(t *testing.T, db *sqlx.DB, id string, ws *otf.Workspace, cv *otf.ConfigurationVersion) *otf.Run {
	rdb := NewRunDB(db)

	run, err := rdb.Create(newTestRun(id, ws, cv))
	require.NoError(t, err)

	return run
}

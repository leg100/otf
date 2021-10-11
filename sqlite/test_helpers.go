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

func newTestRun() *otf.Run {
	return &otf.Run{
		ID:               "run-123",
		StatusTimestamps: make(otf.TimestampMap),
		Plan: &otf.Plan{
			ID:               "plan-123",
			StatusTimestamps: make(otf.TimestampMap),
		},
		Apply: &otf.Apply{
			ID:               "apply-123",
			StatusTimestamps: make(otf.TimestampMap),
		},
		Workspace: &otf.Workspace{
			ID: "ws-123",
		},
		ConfigurationVersion: &otf.ConfigurationVersion{
			ID: "cv-123",
		},
	}
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

func createTestOrganization(db *sqlx.DB, id string) *otf.Organization {
	odb := NewOrganizationDB(db)

	org, err := odb.Create(newTestOrganization(id))
	if err != nil {
		panic("cannot create org")
	}
	return org
}

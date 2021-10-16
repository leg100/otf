package sql

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/require"
)

type newTestStateVersionOption func(*otf.StateVersion) error

func connStr(dbname string) string {
	return fmt.Sprintf("postgres:///%s?host=/var/run/postgresql", dbname)
}

func newTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Connect("postgres", "postgres:///postgres?host=/var/run/postgresql")
	require.NoError(t, err)

	dbname := "db_" + strings.ReplaceAll(uuid.NewString(), "-", "_")

	_, err = db.Exec(fmt.Sprint("CREATE DATABASE ", dbname))
	require.NoError(t, err)

	t.Cleanup(func() {
		if t.Failed() {
			t.Log("database: ", connStr(dbname))
		} else {
			_, err := db.Exec(fmt.Sprint("DROP DATABASE ", dbname))
			if err != nil {
				t.Log(fmt.Errorf("unable to drop database %s: %w", dbname, err))
			}
		}
	})

	connStr := fmt.Sprintf("postgres:///%s?host=/var/run/postgresql&TimeZone=UTC", dbname)
	newdb, err := New(logr.Discard(), connStr)
	require.NoError(t, err)

	t.Cleanup(func() {
		err := newdb.Close()
		require.NoError(t, err)
	})

	return newdb
}

// newTestModel constructs a new model obj with timestamps suitable for unit
// tests interfacing with postgres; tests may want to test for equality with
// timestamps retrieved from postgres, and so the timestamps must be of a
// certain precision and timezone.
func newTestModel() otf.Model {
	now := time.Now().Round(time.Millisecond).UTC()
	return otf.Model{
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func newTestOrganization() *otf.Organization {
	return &otf.Organization{
		ID:    otf.GenerateID("org"),
		Model: newTestModel(),
		Name:  uuid.NewString(),
		Email: "sysadmin@automatize.co.uk",
	}
}

func newTestWorkspace(org *otf.Organization) *otf.Workspace {
	return &otf.Workspace{
		ID:           otf.GenerateID("ws"),
		Model:        newTestModel(),
		Name:         uuid.NewString(),
		Organization: org,
	}
}

func newTestConfigurationVersion(ws *otf.Workspace) *otf.ConfigurationVersion {
	return &otf.ConfigurationVersion{
		ID:               otf.GenerateID("cv"),
		Model:            newTestModel(),
		Status:           otf.ConfigurationPending,
		StatusTimestamps: make(otf.TimestampMap),
		Workspace:        ws,
	}
}

func newTestStateVersion(ws *otf.Workspace, opts ...newTestStateVersionOption) *otf.StateVersion {
	sv := &otf.StateVersion{
		ID:        otf.GenerateID("sv"),
		Model:     newTestModel(),
		Workspace: ws,
	}
	for _, o := range opts {
		o(sv)
	}
	return sv
}

func appendOutput(name, outputType, value string, sensitive bool) newTestStateVersionOption {
	return func(sv *otf.StateVersion) error {
		sv.Outputs = append(sv.Outputs, &otf.StateVersionOutput{
			ID:        otf.GenerateID("svo"),
			Name:      name,
			Type:      outputType,
			Value:     value,
			Sensitive: sensitive,
		})
		return nil
	}
}

func newTestRun(ws *otf.Workspace, cv *otf.ConfigurationVersion) *otf.Run {
	return &otf.Run{
		ID:               otf.GenerateID("run"),
		Model:            newTestModel(),
		Status:           otf.RunPending,
		StatusTimestamps: make(otf.TimestampMap),
		Plan: &otf.Plan{
			ID:               otf.GenerateID("plan"),
			Model:            newTestModel(),
			StatusTimestamps: make(otf.TimestampMap),
		},
		Apply: &otf.Apply{
			ID:               otf.GenerateID("apply"),
			Model:            newTestModel(),
			StatusTimestamps: make(otf.TimestampMap),
		},
		Workspace:            ws,
		ConfigurationVersion: cv,
	}
}

func createTestOrganization(t *testing.T, db *sqlx.DB) *otf.Organization {
	odb := NewOrganizationDB(db)

	org, err := odb.Create(newTestOrganization())
	require.NoError(t, err)

	return org
}

func createTestWorkspace(t *testing.T, db *sqlx.DB, org *otf.Organization) *otf.Workspace {
	wdb := NewWorkspaceDB(db)

	ws, err := wdb.Create(newTestWorkspace(org))
	require.NoError(t, err)

	return ws
}

func createTestConfigurationVersion(t *testing.T, db *sqlx.DB, ws *otf.Workspace) *otf.ConfigurationVersion {
	cdb := NewConfigurationVersionDB(db)

	cv, err := cdb.Create(newTestConfigurationVersion(ws))
	require.NoError(t, err)

	return cv
}

func createTestStateVersion(t *testing.T, db *sqlx.DB, ws *otf.Workspace, opts ...newTestStateVersionOption) *otf.StateVersion {
	sdb := NewStateVersionDB(db)

	sv, err := sdb.Create(newTestStateVersion(ws, opts...))
	require.NoError(t, err)

	return sv
}

func createTestRun(t *testing.T, db *sqlx.DB, ws *otf.Workspace, cv *otf.ConfigurationVersion) *otf.Run {
	rdb := NewRunDB(db)

	run, err := rdb.Create(newTestRun(ws, cv))
	require.NoError(t, err)

	return run
}

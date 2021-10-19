package sql

import (
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/require"
)

const TestDatabaseURL = "OTF_TEST_DATABASE_URL"

type newTestStateVersionOption func(*otf.StateVersion) error

func newTestDB(t *testing.T) *sqlx.DB {
	urlStr := os.Getenv(TestDatabaseURL)
	if urlStr == "" {
		t.Fatalf("%s must be set", TestDatabaseURL)
	}

	u, err := url.Parse(urlStr)
	require.NoError(t, err)

	require.Equal(t, "postgres", u.Scheme)

	// We set both postgres and test fixtures to use TZ so that we can test for
	// timestamp equality between the two. (A go time.Time may use "Local"
	// whereas postgres may set "Europe/London", which would fail an equality
	// test).
	q := u.Query()
	q.Add("TimeZone", "UTC")
	u.RawQuery = q.Encode()

	db, err := New(logr.Discard(), u.String())
	require.NoError(t, err)

	t.Cleanup(func() {
		err := db.Close()
		require.NoError(t, err)
	})

	return db
}

// newTestTimestamps constructs timestamps suitable for unit tests interfacing with
// postgres; tests may want to test for equality with timestamps retrieved from
// postgres, and so the timestamps must be of a certain precision and timezone.
func newTestTimestamps() otf.Timestamps {
	now := time.Now().Round(time.Millisecond).UTC()
	return otf.Timestamps{
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func newTestOrganization() *otf.Organization {
	return &otf.Organization{
		ID:         otf.NewID("org"),
		Timestamps: newTestTimestamps(),
		Name:       uuid.NewString(),
		Email:      "sysadmin@automatize.co.uk",
	}
}

func newTestWorkspace(org *otf.Organization) *otf.Workspace {
	return &otf.Workspace{
		ID:           otf.NewID("ws"),
		Timestamps:   newTestTimestamps(),
		Name:         uuid.NewString(),
		Organization: org,
	}
}

func newTestConfigurationVersion(ws *otf.Workspace) *otf.ConfigurationVersion {
	return &otf.ConfigurationVersion{
		ID:               otf.NewID("cv"),
		Timestamps:       newTestTimestamps(),
		Status:           otf.ConfigurationPending,
		StatusTimestamps: make(otf.TimestampMap),
		Workspace:        ws,
	}
}

func newTestStateVersion(ws *otf.Workspace, opts ...newTestStateVersionOption) *otf.StateVersion {
	sv := &otf.StateVersion{
		ID:         otf.NewID("sv"),
		Timestamps: newTestTimestamps(),
		Workspace:  ws,
	}
	for _, o := range opts {
		o(sv)
	}
	return sv
}

func appendOutput(name, outputType, value string, sensitive bool) newTestStateVersionOption {
	return func(sv *otf.StateVersion) error {
		sv.Outputs = append(sv.Outputs, &otf.StateVersionOutput{
			ID:        otf.NewID("wsout"),
			Name:      name,
			Type:      outputType,
			Value:     value,
			Sensitive: sensitive,
		})
		return nil
	}
}

func newTestRun(ws *otf.Workspace, cv *otf.ConfigurationVersion) *otf.Run {
	id := otf.NewID("run")
	return &otf.Run{
		ID:               id,
		Timestamps:       newTestTimestamps(),
		Status:           otf.RunPending,
		StatusTimestamps: make(otf.TimestampMap),
		Plan: &otf.Plan{
			ID:               otf.NewID("plan"),
			Timestamps:       newTestTimestamps(),
			StatusTimestamps: make(otf.TimestampMap),
			RunID:            id,
		},
		Apply: &otf.Apply{
			ID:               otf.NewID("apply"),
			Timestamps:       newTestTimestamps(),
			StatusTimestamps: make(otf.TimestampMap),
			RunID:            id,
		},
		Workspace:            ws,
		ConfigurationVersion: cv,
	}
}

func createTestOrganization(t *testing.T, db *sqlx.DB) *otf.Organization {
	odb := NewOrganizationDB(db)

	org, err := odb.Create(newTestOrganization())
	require.NoError(t, err)

	t.Cleanup(func() {
		odb.Delete(org.Name)
	})

	return org
}

func createTestWorkspace(t *testing.T, db *sqlx.DB, org *otf.Organization) *otf.Workspace {
	wdb := NewWorkspaceDB(db)

	ws, err := wdb.Create(newTestWorkspace(org))
	require.NoError(t, err)

	t.Cleanup(func() {
		wdb.Delete(otf.WorkspaceSpecifier{ID: otf.String(ws.ID)})
	})

	return ws
}

func createTestConfigurationVersion(t *testing.T, db *sqlx.DB, ws *otf.Workspace) *otf.ConfigurationVersion {
	cdb := NewConfigurationVersionDB(db)

	cv, err := cdb.Create(newTestConfigurationVersion(ws))
	require.NoError(t, err)

	t.Cleanup(func() {
		cdb.Delete(cv.ID)
	})

	return cv
}

func createTestStateVersion(t *testing.T, db *sqlx.DB, ws *otf.Workspace, opts ...newTestStateVersionOption) *otf.StateVersion {
	sdb := NewStateVersionDB(db)

	sv, err := sdb.Create(newTestStateVersion(ws, opts...))
	require.NoError(t, err)

	t.Cleanup(func() {
		sdb.Delete(sv.ID)
	})

	return sv
}

func createTestRun(t *testing.T, db *sqlx.DB, ws *otf.Workspace, cv *otf.ConfigurationVersion) *otf.Run {
	rdb := NewRunDB(db)

	run, err := rdb.Create(newTestRun(ws, cv))
	require.NoError(t, err)

	t.Cleanup(func() {
		rdb.Delete(run.ID)
	})

	return run
}

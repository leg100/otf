package sql

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/require"

	_ "github.com/jackc/pgx/v4"
)

const TestDatabaseURL = "OTF_TEST_DATABASE_URL"

type newTestStateVersionOption func(*otf.StateVersion) error

func newTestDB(t *testing.T, sessionCleanupIntervalOverride ...time.Duration) otf.DB {
	urlStr := os.Getenv(TestDatabaseURL)
	if urlStr == "" {
		t.Fatalf("%s must be set", TestDatabaseURL)
	}

	u, err := url.Parse(urlStr)
	require.NoError(t, err)

	require.Equal(t, "postgres", u.Scheme)

	interval := DefaultSessionCleanupInterval
	if len(sessionCleanupIntervalOverride) > 0 {
		interval = sessionCleanupIntervalOverride[0]
	}

	db, err := New(logr.Discard(), u.String(), nil, interval)
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
		ID:         otf.NewID("cv"),
		Timestamps: newTestTimestamps(),
		Status:     otf.ConfigurationPending,
		Workspace:  ws,
	}
}

func newTestStateVersion(run *otf.Run, opts ...newTestStateVersionOption) *otf.StateVersion {
	sv := &otf.StateVersion{
		ID:         otf.NewID("sv"),
		Timestamps: newTestTimestamps(),
		Run:        run,
	}
	for _, o := range opts {
		o(sv)
	}
	return sv
}

func newTestUser() *otf.User {
	return &otf.User{
		ID:         otf.NewID("user"),
		Timestamps: newTestTimestamps(),
		Username:   fmt.Sprintf("mr-%s", otf.GenerateRandomString(6)),
	}
}

type newTestSessionOption func(*otf.Session)

func withFlash(flash *otf.Flash) newTestSessionOption {
	return func(session *otf.Session) {
		session.SessionData.Flash = flash
	}
}

func overrideExpiry(expiry time.Time) newTestSessionOption {
	return func(session *otf.Session) {
		session.Expiry = expiry
	}
}

func newTestSession(t *testing.T, userID string, opts ...newTestSessionOption) *otf.Session {
	session, err := otf.NewSession(userID, &otf.SessionData{
		Address: "127.0.0.1",
	})
	require.NoError(t, err)

	for _, o := range opts {
		o(session)
	}

	session.Timestamps = newTestTimestamps()

	return session
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
		ID:         id,
		Timestamps: newTestTimestamps(),
		Status:     otf.RunPending,
		Plan: &otf.Plan{
			ID:         otf.NewID("plan"),
			Timestamps: newTestTimestamps(),
			RunID:      id,
		},
		Apply: &otf.Apply{
			ID:         otf.NewID("apply"),
			Timestamps: newTestTimestamps(),
			RunID:      id,
		},
		Workspace:            ws,
		ConfigurationVersion: cv,
	}
}

func createTestOrganization(t *testing.T, db otf.DB) *otf.Organization {
	org, err := db.OrganizationStore().Create(newTestOrganization())
	require.NoError(t, err)

	t.Cleanup(func() {
		db.OrganizationStore().Delete(org.Name)
	})

	return org
}

func createTestWorkspace(t *testing.T, db otf.DB, org *otf.Organization) *otf.Workspace {
	ws, err := db.WorkspaceStore().Create(newTestWorkspace(org))
	require.NoError(t, err)

	t.Cleanup(func() {
		db.WorkspaceStore().Delete(otf.WorkspaceSpec{ID: otf.String(ws.ID)})
	})

	return ws
}

func createTestConfigurationVersion(t *testing.T, db otf.DB, ws *otf.Workspace) *otf.ConfigurationVersion {
	cv, err := db.ConfigurationVersionStore().Create(newTestConfigurationVersion(ws))
	require.NoError(t, err)

	t.Cleanup(func() {
		db.ConfigurationVersionStore().Delete(cv.ID)
	})

	return cv
}

func createTestStateVersion(t *testing.T, db otf.DB, run *otf.Run, opts ...newTestStateVersionOption) *otf.StateVersion {
	sv, err := db.StateVersionStore().Create(newTestStateVersion(run, opts...))
	require.NoError(t, err)

	t.Cleanup(func() {
		db.StateVersionStore().Delete(sv.ID)
	})

	return sv
}

func createTestRun(t *testing.T, db otf.DB, ws *otf.Workspace, cv *otf.ConfigurationVersion) *otf.Run {
	run, err := db.RunStore().Create(newTestRun(ws, cv))
	require.NoError(t, err)

	t.Cleanup(func() {
		db.RunStore().Delete(run.ID)
	})

	return run
}

type createTestUserOpt func(*otf.User)

func withOrganizationMemberships(memberships ...*otf.Organization) createTestUserOpt {
	return func(user *otf.User) {
		user.Organizations = memberships
	}
}

func createTestUser(t *testing.T, db otf.DB, opts ...createTestUserOpt) *otf.User {
	user := newTestUser()

	for _, o := range opts {
		o(user)
	}

	err := db.UserStore().Create(context.Background(), user)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.UserStore().Delete(context.Background(), otf.UserSpec{Username: &user.Username})
	})

	return user
}

func createTestSession(t *testing.T, db otf.DB, userID string, opts ...newTestSessionOption) *otf.Session {
	session := newTestSession(t, userID, opts...)

	err := db.SessionStore().CreateSession(context.Background(), session)
	require.NoError(t, err)

	t.Cleanup(func() {
		err := db.SessionStore().DeleteSession(context.Background(), session.Token)
		require.NoError(t, err)
	})

	return session
}

func createTestToken(t *testing.T, db otf.DB, userID, description string) *otf.Token {
	token, err := otf.NewToken(userID, description)
	require.NoError(t, err)

	err = db.TokenStore().CreateToken(context.Background(), token)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.TokenStore().DeleteToken(context.Background(), token.Token)
	})

	return token
}

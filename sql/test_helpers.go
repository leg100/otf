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

func newTestOrganization(t *testing.T) *otf.Organization {
	org, err := otf.NewOrganization(otf.OrganizationCreateOptions{Name: otf.String(uuid.NewString())})
	require.NoError(t, err)
	return org
}

func newTestWorkspace(t *testing.T, org *otf.Organization) *otf.Workspace {
	orgID := org.ID()
	ws, err := (&otf.WorkspaceFactory{}).NewWorkspace(context.Background(), otf.WorkspaceCreateOptions{
		Name:           uuid.NewString(),
		OrganizationID: &orgID,
	})
	require.NoError(t, err)
	return ws
}

func newTestConfigurationVersion(t *testing.T, ws *otf.Workspace) *otf.ConfigurationVersion {
	cv, err := otf.NewConfigurationVersion(ws.ID(), otf.ConfigurationVersionCreateOptions{})
	require.NoError(t, err)
	return cv
}

type newTestSessionOption func(*otf.Session)

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

	return session
}

func newTestRun(ws *otf.Workspace, cv *otf.ConfigurationVersion) *otf.Run {
	return otf.NewRun(cv, ws, otf.RunCreateOptions{})
}

func createTestOrganization(t *testing.T, db otf.DB) *otf.Organization {
	org := newTestOrganization(t)
	err := db.OrganizationStore().Create(context.Background(), org)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.OrganizationStore().Delete(context.Background(), org.Name())
	})
	return org
}

func createTestWorkspace(t *testing.T, db otf.DB, org *otf.Organization) *otf.Workspace {
	ws := newTestWorkspace(t, org)
	err := db.WorkspaceStore().Create(context.Background(), ws)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.WorkspaceStore().Delete(context.Background(), otf.WorkspaceSpec{ID: otf.String(ws.ID())})
	})
	return ws
}

func createTestConfigurationVersion(t *testing.T, db otf.DB, ws *otf.Workspace) *otf.ConfigurationVersion {
	cv := newTestConfigurationVersion(t, ws)
	err := db.ConfigurationVersionStore().Create(context.Background(), cv)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.ConfigurationVersionStore().Delete(context.Background(), cv.ID())
	})
	return cv
}

func createTestStateVersion(t *testing.T, db otf.DB, ws *otf.Workspace, outputs ...otf.StateOutput) *otf.StateVersion {
	sv := otf.NewTestStateVersion(t, outputs...)
	err := db.StateVersionStore().Create(context.Background(), ws.ID(), sv)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.StateVersionStore().Delete(context.Background(), sv.ID())
	})
	return sv
}

func createTestRun(t *testing.T, db otf.DB, ws *otf.Workspace, cv *otf.ConfigurationVersion) *otf.Run {
	run := newTestRun(ws, cv)
	err := db.RunStore().Create(context.Background(), run)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.RunStore().Delete(context.Background(), run.ID())
	})
	return run
}

func createTestUser(t *testing.T, db otf.DB, opts ...otf.NewUserOption) *otf.User {
	username := fmt.Sprintf("mr-%s", otf.GenerateRandomString(6))
	user := otf.NewUser(username, opts...)

	err := db.UserStore().Create(context.Background(), user)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.UserStore().Delete(context.Background(), otf.UserSpec{Username: otf.String(user.Username())})
	})
	return user
}

func createTestSession(t *testing.T, db otf.DB, userID string, opts ...newTestSessionOption) *otf.Session {
	session := newTestSession(t, userID, opts...)
	ctx := context.Background()

	err := db.SessionStore().CreateSession(ctx, session)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.SessionStore().DeleteSession(ctx, session.Token)
	})
	return session
}

func createTestToken(t *testing.T, db otf.DB, userID, description string) *otf.Token {
	ctx := context.Background()

	token, err := otf.NewToken(userID, description)
	require.NoError(t, err)

	err = db.TokenStore().CreateToken(ctx, token)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.TokenStore().DeleteToken(ctx, token.Token())
	})
	return token
}

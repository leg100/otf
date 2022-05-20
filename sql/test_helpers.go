package sql

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/mitchellh/copystructure"
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

func newTestOrganization() *otf.Organization {
	return otf.NewTestOrganization()
}

func newTestWorkspace(org *otf.Organization) *otf.Workspace {
	return otf.NewTestWorkspace(org)
}

func newShallowNestedWorkspace(ws *otf.Workspace) *otf.Workspace {
	cp, _ := copystructure.Copy(ws)
	shallowWorkspace := cp.(*otf.Workspace)
	shallowWorkspace.Organization = &otf.Organization{ID: shallowWorkspace.Organization.ID}
	return shallowWorkspace
}

func newTestConfigurationVersion(ws *otf.Workspace) *otf.ConfigurationVersion {
	return otf.NewConfigurationVersionFromDefaults(otf.NewShallowNestedWorkspace(ws))
}

func newTestStateVersion(opts ...newTestStateVersionOption) *otf.StateVersion {
	sv := &otf.StateVersion{
		ID:    otf.NewID("sv"),
		State: []byte("stuff"),
	}
	for _, o := range opts {
		o(sv)
	}
	return sv
}

func newTestUser() *otf.User {
	return &otf.User{
		ID:       otf.NewID("user"),
		Username: fmt.Sprintf("mr-%s", otf.GenerateRandomString(6)),
	}
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

func appendOutput(name, outputType, value string, sensitive bool) newTestStateVersionOption {
	return func(sv *otf.StateVersion) error {
		sv.Outputs = append(sv.Outputs, &otf.StateVersionOutput{
			ID:             otf.NewID("wsout"),
			Name:           name,
			Type:           outputType,
			Value:          value,
			Sensitive:      sensitive,
			StateVersionID: sv.ID,
		})
		return nil
	}
}

func newTestRun(ws *otf.Workspace, cv *otf.ConfigurationVersion) *otf.Run {
	ws = newShallowNestedWorkspace(ws)
	cv.ShallowNest()

	return otf.NewRunFromDefaults(cv, ws)
}

func createTestOrganization(t *testing.T, db otf.DB) *otf.Organization {
	org, err := db.OrganizationStore().Create(newTestOrganization())
	require.NoError(t, err)

	t.Cleanup(func() {
		db.OrganizationStore().Delete(org.Name())
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

func createTestStateVersion(t *testing.T, db otf.DB, ws *otf.Workspace, opts ...newTestStateVersionOption) *otf.StateVersion {
	sv := newTestStateVersion(opts...)
	err := db.StateVersionStore().Create(ws.ID, sv)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.StateVersionStore().Delete(sv.ID)
	})

	return sv
}

func createTestRun(t *testing.T, db otf.DB, ws *otf.Workspace, cv *otf.ConfigurationVersion) *otf.Run {
	run := newTestRun(ws, cv)
	err := db.RunStore().Create(run)
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
		db.TokenStore().DeleteToken(ctx, token.Token)
	})

	return token
}

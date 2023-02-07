package sql

import (
	"context"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/inmem"
	"github.com/leg100/otf/rbac"
	"github.com/stretchr/testify/require"

	_ "github.com/jackc/pgx/v4"
)

const TestDatabaseURL = "OTF_TEST_DATABASE_URL"

func NewTestDB(t *testing.T, overrides ...newTestDBOption) *DB {
	urlStr := os.Getenv(TestDatabaseURL)
	if urlStr == "" {
		t.Fatalf("%s must be set", TestDatabaseURL)
	}

	u, err := url.Parse(urlStr)
	require.NoError(t, err)

	require.Equal(t, "postgres", u.Scheme)

	opts := Options{
		Logger:       logr.Discard(),
		ConnString:   u.String(),
		Cache:        nil,
		CloudService: inmem.NewTestCloudService(),
	}

	for _, or := range overrides {
		or(&opts)
	}

	db, err := New(context.Background(), opts)
	require.NoError(t, err)

	t.Cleanup(func() { db.Close() })

	return db
}

type newTestDBOption func(*Options)

func overrideCleanupInterval(d time.Duration) newTestDBOption {
	return func(o *Options) {
		o.CleanupInterval = d
	}
}

func createTestWorkspacePermission(t *testing.T, db otf.DB, ws *otf.Workspace, team *otf.Team, role rbac.Role) *otf.WorkspacePermission {
	ctx := context.Background()
	err := db.SetWorkspacePermission(ctx, ws.ID(), team.Name(), role)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.UnsetWorkspacePermission(ctx, ws.ID(), team.Name())
	})
	return &otf.WorkspacePermission{Team: team, Role: role}
}

func createTestRun(t *testing.T, db otf.DB, ws *otf.Workspace, cv *otf.ConfigurationVersion) *otf.Run {
	ctx := context.Background()
	run := otf.NewRun(cv, ws, otf.RunCreateOptions{})
	err := db.CreateRun(ctx, run)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.DeleteRun(ctx, run.ID())
	})
	return run
}

func createTestSession(t *testing.T, db otf.DB, userID string, opts ...otf.NewSessionOption) *otf.Session {
	session := otf.NewTestSession(t, userID, opts...)
	ctx := context.Background()

	err := db.CreateSession(ctx, session)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.DeleteSession(ctx, session.Token())
	})
	return session
}

func createTestToken(t *testing.T, db otf.DB, userID, description string) *otf.Token {
	ctx := context.Background()

	token, err := otf.NewToken(userID, description)
	require.NoError(t, err)

	err = db.CreateToken(ctx, token)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.DeleteToken(ctx, token.Token())
	})
	return token
}

func createTestModule(t *testing.T, db *DB, org *otf.Organization) *otf.Module {
	ctx := context.Background()

	module := otf.NewTestModule(org)
	err := db.CreateModule(ctx, module)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.DeleteModule(ctx, module.ID())
	})
	return module
}

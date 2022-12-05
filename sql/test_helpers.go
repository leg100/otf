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
	"github.com/leg100/otf/github"
	"github.com/leg100/otf/inmem"
	"github.com/leg100/otf/sql/pggen"
	"github.com/stretchr/testify/require"

	_ "github.com/jackc/pgx/v4"
)

const TestDatabaseURL = "OTF_TEST_DATABASE_URL"

func newTestDB(t *testing.T, sessionCleanupIntervalOverride ...time.Duration) *DB {
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

	db, err := New(context.Background(), Options{
		Logger:          logr.Discard(),
		Path:            u.String(),
		Cache:           nil,
		CleanupInterval: interval,
		CloudService:    inmem.NewTestCloudService(),
	})
	require.NoError(t, err)

	t.Cleanup(func() { db.Close() })

	return db
}

func createTestWorkspacePermission(t *testing.T, db otf.DB, ws *otf.Workspace, team *otf.Team, role otf.WorkspaceRole) *otf.WorkspacePermission {
	ctx := context.Background()
	err := db.SetWorkspacePermission(ctx, ws.SpecName(), team.Name(), role)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.UnsetWorkspacePermission(ctx, ws.SpecName(), team.Name())
	})
	return &otf.WorkspacePermission{Team: team, Role: role}
}

func createTestOrganization(t *testing.T, db otf.DB) *otf.Organization {
	org := otf.NewTestOrganization(t)
	err := db.CreateOrganization(context.Background(), org)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.DeleteOrganization(context.Background(), org.Name())
	})
	return org
}

func createTestTeam(t *testing.T, db otf.DB, org *otf.Organization) *otf.Team {
	team := otf.NewTestTeam(t, org)
	err := db.CreateTeam(context.Background(), team)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.DeleteTeam(context.Background(), team.Name(), org.Name())
	})
	return team
}

func createTestWorkspace(t *testing.T, db otf.DB, org *otf.Organization) *otf.Workspace {
	ws := otf.NewTestWorkspace(t, org, otf.WorkspaceCreateOptions{})
	err := db.CreateWorkspace(context.Background(), ws)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.DeleteWorkspace(context.Background(), otf.WorkspaceSpec{ID: otf.String(ws.ID())})
	})
	return ws
}

func createTestConfigurationVersion(t *testing.T, db otf.DB, ws *otf.Workspace, opts otf.ConfigurationVersionCreateOptions) *otf.ConfigurationVersion {
	cv := otf.NewTestConfigurationVersion(t, ws, opts)
	err := db.CreateConfigurationVersion(context.Background(), cv)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.DeleteConfigurationVersion(context.Background(), cv.ID())
	})
	return cv
}

func createTestStateVersion(t *testing.T, db otf.DB, ws *otf.Workspace, outputs ...otf.StateOutput) *otf.StateVersion {
	sv := otf.NewTestStateVersion(t, outputs...)
	err := db.CreateStateVersion(context.Background(), ws.ID(), sv)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.DeleteStateVersion(context.Background(), sv.ID())
	})
	return sv
}

func createTestRun(t *testing.T, db otf.DB, ws *otf.Workspace, cv *otf.ConfigurationVersion) *otf.Run {
	run := otf.NewRun(cv, ws, otf.RunCreateOptions{})
	err := db.CreateRun(context.Background(), run)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.DeleteRun(context.Background(), run.ID())
	})
	return run
}

func createTestUser(t *testing.T, db otf.DB, opts ...otf.NewUserOption) *otf.User {
	username := fmt.Sprintf("mr-%s", otf.GenerateRandomString(6))
	user := otf.NewUser(username, opts...)

	err := db.CreateUser(context.Background(), user)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.DeleteUser(context.Background(), otf.UserSpec{Username: otf.String(user.Username())})
	})
	return user
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

func newTestVCSProvider(t *testing.T, org *otf.Organization) *otf.VCSProvider {
	factory := &otf.VCSProviderFactory{inmem.NewTestCloudService()}
	provider, err := factory.NewVCSProvider(otf.VCSProviderCreateOptions{
		OrganizationName: org.Name(),
		// unit tests require a legitimate cloud name to avoid invalid foreign
		// key error upon insert/update
		Cloud: "github",
		Name:  uuid.NewString(),
		Token: uuid.NewString(),
	})
	require.NoError(t, err)
	return provider
}

func createTestVCSProvider(t *testing.T, db otf.DB, organization *otf.Organization) *otf.VCSProvider {
	provider := newTestVCSProvider(t, organization)
	ctx := context.Background()

	err := db.CreateVCSProvider(ctx, provider)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.DeleteVCSProvider(ctx, provider.ID())
	})
	return provider
}

func createTestWorkspaceRepo(t *testing.T, db *DB, ws *otf.Workspace, provider *otf.VCSProvider, hook *otf.Webhook) *otf.WorkspaceRepo {
	ctx := context.Background()

	repo := otf.WorkspaceRepo{
		ProviderID: provider.ID(),
		Branch:     "master",
		Webhook:    hook,
	}
	ws, err := db.CreateWorkspaceRepo(ctx, ws.SpecID(), repo)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.DeleteWorkspaceRepo(ctx, ws.SpecID())
	})
	return ws.Repo()
}

func createTestWebhook(t *testing.T, db *DB) *otf.Webhook {
	ctx := context.Background()
	repo := otf.NewTestRepo()
	hook := otf.NewTestWebhook(repo, github.Defaults())

	_, err := db.InsertWebhook(ctx, pggen.InsertWebhookParams{
		WebhookID:  UUID(hook.WebhookID),
		VCSID:      String(hook.VCSID),
		Secret:     String(hook.Secret),
		Identifier: String(hook.Identifier),
		HTTPURL:    String(hook.HTTPURL),
		Cloud:      String(hook.CloudName()),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		db.DeleteWebhook(ctx, hook.WebhookID)
	})
	return hook
}

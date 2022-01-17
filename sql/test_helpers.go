package sql

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/require"
)

const TestDatabaseURL = "OTF_TEST_DATABASE_URL"

type newTestStateVersionOption func(*otf.StateVersion) error

func newTestDB(t *testing.T) otf.DB {
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

	db, err := New(logr.Discard(), u.String(), nil)
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

func newTestUser() *otf.User {
	return &otf.User{
		ID:         otf.NewID("user"),
		Timestamps: newTestTimestamps(),
		Username:   fmt.Sprintf("mr-%s", otf.GenerateRandomString(6)),
	}
}

func newTestSession(userID string) *otf.Session {
	token, _ := generateToken()

	return &otf.Session{
		Token:  token,
		Expiry: time.Now().Add(time.Second * 10),
		UserID: userID,
	}
}

// generateToken is taken from alexedwards/scs - here for generating a session
// token for testing purposes.
func generateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
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
			PlanFile:         []byte{},
			PlanJSON:         []byte{},
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
		db.WorkspaceStore().Delete(otf.WorkspaceSpecifier{ID: otf.String(ws.ID)})
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
	sv, err := db.StateVersionStore().Create(newTestStateVersion(ws, opts...))
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

func createTestUser(t *testing.T, db otf.DB) *otf.User {
	user := newTestUser()

	err := db.UserStore().Create(context.Background(), user)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.UserStore().Delete(context.Background(), user.ID)
	})

	return user
}

func createTestSession(t *testing.T, db otf.DB, userID string) *otf.Session {
	session := newTestSession(userID)

	insertSQL := `INSERT INTO sessions (data, token, expiry, user_id) VALUES (:data, :token, :expiry, :user_id)`
	sql, args, err := db.Handle().BindNamed(insertSQL, session)
	require.NoError(t, err)

	_, err = db.Handle().Exec(sql, args...)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.Handle().Exec("DELETE FROM sessions WHERE token = ?", session.Token)
	})

	return session
}

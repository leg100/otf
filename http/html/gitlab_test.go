package html

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xanzy/go-gitlab"
)

func TestGitlab_GetUser(t *testing.T) {
	http.HandleFunc("/api/v4/user", func(w http.ResponseWriter, r *http.Request) {
		out, err := json.Marshal(&gitlab.User{ID: 123, Username: "fake-user"})
		require.NoError(t, err)
		w.Header().Add("Content-Type", "application/json")
		w.Write(out)
	})
	http.HandleFunc("/api/v4/groups", func(w http.ResponseWriter, r *http.Request) {
		out, err := json.Marshal([]*gitlab.Group{
			{ID: 789, Path: "fake-group"},
		})
		require.NoError(t, err)
		w.Header().Add("Content-Type", "application/json")
		w.Write(out)
	})
	http.HandleFunc("/api/v4/groups/789/members/123", func(w http.ResponseWriter, r *http.Request) {
		out, err := json.Marshal(&gitlab.GroupMember{AccessLevel: gitlab.MaintainerPermissions})
		require.NoError(t, err)
		w.Header().Add("Content-Type", "application/json")
		w.Write(out)
	})

	srv := httptest.NewServer(nil)
	t.Cleanup(srv.Close)

	client, err := gitlab.NewOAuthClient("fake-oauth-token", gitlab.WithBaseURL(srv.URL))
	require.NoError(t, err)

	provider := &gitlabProvider{client: client}

	user, err := provider.GetUser(context.Background())
	require.NoError(t, err)

	assert.Equal(t, "fake-user", user.Username())
	if assert.Equal(t, 1, len(user.Organizations())) {
		assert.Equal(t, "fake-group", user.Organizations()[0].Name())
	}
	if assert.Equal(t, 1, len(user.Teams())) {
		assert.Equal(t, "maintainers", user.Teams()[0].Name())
	}
}

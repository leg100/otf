package html

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-github/v41/github"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestGithub_GetUser(t *testing.T) {
	ctx := context.Background()

	http.HandleFunc("/api/v3/user", func(w http.ResponseWriter, r *http.Request) {
		out, err := json.Marshal(&github.User{Login: otf.String("fake-user")})
		require.NoError(t, err)
		w.Header().Add("Content-Type", "application/json")
		w.Write(out)
	})
	http.HandleFunc("/api/v3/user/orgs", func(w http.ResponseWriter, r *http.Request) {
		out, err := json.Marshal([]*github.Organization{{Login: otf.String("fake-org")}})
		require.NoError(t, err)
		w.Header().Add("Content-Type", "application/json")
		w.Write(out)
	})
	http.HandleFunc("/api/v3/user/teams", func(w http.ResponseWriter, r *http.Request) {
		out, err := json.Marshal([]*github.Team{
			{
				Name: otf.String("fake-team"),
				Organization: &github.Organization{
					Login: otf.String("fake-org"),
				},
			},
		})
		require.NoError(t, err)
		w.Header().Add("Content-Type", "application/json")
		w.Write(out)
	})

	srv := httptest.NewTLSServer(nil)
	t.Cleanup(srv.Close)
	u, err := url.Parse(srv.URL)
	require.NoError(t, err)

	cloud := &GithubCloud{
		&GithubConfig{
			cloudConfig{
				hostname:            u.Host,
				skipTLSVerification: true,
			},
		},
	}
	client, err := cloud.NewDirectoryClient(ctx, DirectoryClientOptions{
		Token: &oauth2.Token{AccessToken: "fake-token"},
	})
	require.NoError(t, err)

	user, err := client.GetUser(ctx)
	require.NoError(t, err)

	assert.Equal(t, "fake-user", user.Username())
	if assert.Equal(t, 1, len(user.Organizations)) {
		assert.Equal(t, "fake-org", user.Organizations[0].Name())
	}
	if assert.Equal(t, 1, len(user.Teams)) {
		assert.Equal(t, "fake-team", user.Teams[0].Name())
	}
}

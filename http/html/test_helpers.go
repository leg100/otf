package html

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-github/v41/github"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func NewTestGithubServer(t *testing.T, user *otf.User) *httptest.Server {
	http.HandleFunc("/login/oauth/authorize", func(w http.ResponseWriter, r *http.Request) {
		q := url.Values{}
		q.Add("state", r.URL.Query().Get("state"))
		q.Add("code", otf.GenerateRandomString(10))

		referrer, err := url.Parse(r.Referer())
		require.NoError(t, err)

		callback := url.URL{
			Scheme:   referrer.Scheme,
			Host:     referrer.Host,
			Path:     "/oauth/github/callback",
			RawQuery: q.Encode(),
		}

		http.Redirect(w, r, callback.String(), http.StatusFound)
	})
	http.HandleFunc("/login/oauth/access_token", func(w http.ResponseWriter, r *http.Request) {
		out, err := json.Marshal(&oauth2.Token{AccessToken: "stub_token"})
		require.NoError(t, err)
		w.Header().Add("Content-Type", "application/json")
		w.Write(out)
	})
	http.HandleFunc("/api/v3/user", func(w http.ResponseWriter, r *http.Request) {
		out, err := json.Marshal(&github.User{Login: otf.String(user.Username())})
		require.NoError(t, err)
		w.Header().Add("Content-Type", "application/json")
		w.Write(out)
	})
	http.HandleFunc("/api/v3/user/orgs", func(w http.ResponseWriter, r *http.Request) {
		var orgs []*github.Organization
		for _, org := range user.Organizations() {
			orgs = append(orgs, &github.Organization{Login: otf.String(org.Name())})
		}
		out, err := json.Marshal(orgs)
		require.NoError(t, err)
		w.Header().Add("Content-Type", "application/json")
		w.Write(out)
	})
	for _, org := range user.Organizations() {
		http.HandleFunc("/api/v3/user/memberships/orgs/"+org.Name(), func(w http.ResponseWriter, r *http.Request) {
			out, err := json.Marshal(&github.Membership{
				Role: otf.String("admin"),
			})
			require.NoError(t, err)
			w.Header().Add("Content-Type", "application/json")
			w.Write(out)
		})
	}
	http.HandleFunc("/api/v3/user/teams", func(w http.ResponseWriter, r *http.Request) {
		var teams []*github.Team
		for _, team := range user.Teams() {
			teams = append(teams, &github.Team{
				Name: otf.String(team.Name()),
				Organization: &github.Organization{
					Login: otf.String(team.OrganizationName()),
				},
			})
		}
		out, err := json.Marshal(teams)
		require.NoError(t, err)
		w.Header().Add("Content-Type", "application/json")
		w.Write(out)
	})

	srv := httptest.NewTLSServer(nil)
	t.Cleanup(srv.Close)
	return srv
}

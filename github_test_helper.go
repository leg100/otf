package otf

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-github/v41/github"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

type testGithubServerDB struct {
	user    *User
	repo    *Repo
	tarball []byte
}

func NewTestGithubServer(t *testing.T, opts ...TestGithubServerOption) *httptest.Server {
	db := &testGithubServerDB{}
	for _, o := range opts {
		o(db)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/login/oauth/authorize", func(w http.ResponseWriter, r *http.Request) {
		q := url.Values{}
		q.Add("state", r.URL.Query().Get("state"))
		q.Add("code", GenerateRandomString(10))

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
	mux.HandleFunc("/login/oauth/access_token", func(w http.ResponseWriter, r *http.Request) {
		out, err := json.Marshal(&oauth2.Token{AccessToken: "stub_token"})
		require.NoError(t, err)
		w.Header().Add("Content-Type", "application/json")
		w.Write(out)
	})
	if db.user != nil {
		mux.HandleFunc("/api/v3/user", func(w http.ResponseWriter, r *http.Request) {
			out, err := json.Marshal(&github.User{Login: String(db.user.Username())})
			require.NoError(t, err)
			w.Header().Add("Content-Type", "application/json")
			w.Write(out)
		})
		mux.HandleFunc("/api/v3/user/orgs", func(w http.ResponseWriter, r *http.Request) {
			var orgs []*github.Organization
			for _, org := range db.user.Organizations() {
				orgs = append(orgs, &github.Organization{Login: String(org.Name())})
			}
			out, err := json.Marshal(orgs)
			require.NoError(t, err)
			w.Header().Add("Content-Type", "application/json")
			w.Write(out)
		})
		for _, org := range db.user.Organizations() {
			mux.HandleFunc("/api/v3/user/memberships/orgs/"+org.Name(), func(w http.ResponseWriter, r *http.Request) {
				out, err := json.Marshal(&github.Membership{
					Role: String("member"),
				})
				require.NoError(t, err)
				w.Header().Add("Content-Type", "application/json")
				w.Write(out)
			})
		}
		mux.HandleFunc("/api/v3/user/teams", func(w http.ResponseWriter, r *http.Request) {
			var teams []*github.Team
			for _, team := range db.user.Teams() {
				teams = append(teams, &github.Team{
					Name: String(team.Name()),
					Organization: &github.Organization{
						Login: String(team.OrganizationName()),
					},
				})
			}
			out, err := json.Marshal(teams)
			require.NoError(t, err)
			w.Header().Add("Content-Type", "application/json")
			w.Write(out)
		})
	}
	mux.HandleFunc("/api/v3/user/repos", func(w http.ResponseWriter, r *http.Request) {
		repos := []*github.Repository{
			{
				FullName:      String(db.repo.Identifier),
				URL:           String(db.repo.HTTPURL),
				DefaultBranch: String(db.repo.Branch),
			},
		}
		out, err := json.Marshal(repos)
		require.NoError(t, err)
		w.Header().Add("Content-Type", "application/json")
		w.Write(out)
	})
	if db.repo != nil {
		mux.HandleFunc("/api/v3/repos/"+db.repo.Identifier, func(w http.ResponseWriter, r *http.Request) {
			repo := &github.Repository{
				FullName:      String(db.repo.Identifier),
				URL:           String(db.repo.HTTPURL),
				DefaultBranch: String(db.repo.Branch),
			}
			out, err := json.Marshal(repo)
			require.NoError(t, err)
			w.Header().Add("Content-Type", "application/json")
			w.Write(out)
		})
		// https://docs.github.com/en/rest/repos/contents#download-a-repository-archive-tar
		mux.HandleFunc("/api/v3/repos/"+db.repo.Identifier+"/tarball/"+db.repo.Branch, func(w http.ResponseWriter, r *http.Request) {
			link := url.URL{Scheme: "https", Host: r.Host, Path: "/mytarball"}
			http.Redirect(w, r, link.String(), http.StatusFound)
		})
	}

	if db.tarball != nil {
		mux.HandleFunc("/mytarball", func(w http.ResponseWriter, r *http.Request) {
			w.Write(db.tarball)
		})
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("github server received request for non-existent path: %s", r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
	})

	srv := httptest.NewTLSServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

type TestGithubServerOption func(*testGithubServerDB)

func WithGithubUser(user *User) TestGithubServerOption {
	return func(db *testGithubServerDB) {
		db.user = user
	}
}

func WithGithubRepo(repo *Repo) TestGithubServerOption {
	return func(db *testGithubServerDB) {
		db.repo = repo
	}
}

func WithGithubArchive(tarball []byte) TestGithubServerOption {
	return func(db *testGithubServerDB) {
		db.tarball = tarball
	}
}

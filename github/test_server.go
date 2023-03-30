package github

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-github/v41/github"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

type (
	TestServer struct {
		HookEndpoint *string  // populated upon creation
		HookSecret   *string  // populated upon creation
		HookEvents   []string // populated upon creation

		statusCallback // callback invoked whenever a commit status is received

		*httptest.Server
		*testdb
	}

	TestServerOption func(*TestServer)

	testdb struct {
		user    *cloud.User
		repo    *string
		tarball []byte
		refs    []string
	}

	statusCallback func(*github.StatusEvent)
)

func NewTestServer(t *testing.T, opts ...TestServerOption) (*TestServer, cloud.Config) {
	srv := TestServer{
		testdb: &testdb{},
	}
	for _, o := range opts {
		o(&srv)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/login/oauth/authorize", func(w http.ResponseWriter, r *http.Request) {
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
	mux.HandleFunc("/login/oauth/access_token", func(w http.ResponseWriter, r *http.Request) {
		out, err := json.Marshal(&oauth2.Token{AccessToken: "stub_token"})
		require.NoError(t, err)
		w.Header().Add("Content-Type", "application/json")
		w.Write(out)
	})
	if srv.user != nil {
		mux.HandleFunc("/api/v3/user", func(w http.ResponseWriter, r *http.Request) {
			out, err := json.Marshal(&github.User{Login: otf.String(srv.user.Name)})
			require.NoError(t, err)
			w.Header().Add("Content-Type", "application/json")
			w.Write(out)
		})
		mux.HandleFunc("/api/v3/user/orgs", func(w http.ResponseWriter, r *http.Request) {
			var orgs []*github.Organization
			for _, org := range srv.user.Organizations() {
				orgs = append(orgs, &github.Organization{Login: otf.String(org)})
			}
			out, err := json.Marshal(orgs)
			require.NoError(t, err)
			w.Header().Add("Content-Type", "application/json")
			w.Write(out)
		})
		for _, org := range srv.user.Organizations() {
			mux.HandleFunc("/api/v3/user/memberships/orgs/"+org, func(w http.ResponseWriter, r *http.Request) {
				out, err := json.Marshal(&github.Membership{
					Role: otf.String("member"),
				})
				require.NoError(t, err)
				w.Header().Add("Content-Type", "application/json")
				w.Write(out)
			})
		}
		mux.HandleFunc("/api/v3/user/teams", func(w http.ResponseWriter, r *http.Request) {
			var teams []*github.Team
			for _, team := range srv.user.Teams {
				teams = append(teams, &github.Team{
					Slug: otf.String(team.Name),
					Organization: &github.Organization{
						Login: otf.String(team.Organization),
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
		repos := []*github.Repository{{FullName: srv.repo}}
		out, err := json.Marshal(repos)
		require.NoError(t, err)
		w.Header().Add("Content-Type", "application/json")
		w.Write(out)
	})
	if srv.repo != nil {
		mux.HandleFunc("/api/v3/repos/"+*srv.repo+"/git/matching-refs/", func(w http.ResponseWriter, r *http.Request) {
			var refs []*github.Reference
			for _, ref := range srv.refs {
				refs = append(refs, &github.Reference{Ref: otf.String(ref)})
			}
			out, err := json.Marshal(refs)
			require.NoError(t, err)
			w.Header().Add("Content-Type", "application/json")
			w.Write(out)
		})
		mux.HandleFunc("/api/v3/repos/"+*srv.repo, func(w http.ResponseWriter, r *http.Request) {
			repo := &github.Repository{FullName: srv.repo}
			out, err := json.Marshal(repo)
			require.NoError(t, err)
			w.Header().Add("Content-Type", "application/json")
			w.Write(out)
		})
		mux.HandleFunc("/api/v3/repos/"+*srv.repo+"/tarball/", func(w http.ResponseWriter, r *http.Request) {
			link := url.URL{Scheme: "https", Host: r.Host, Path: "/mytarball"}
			http.Redirect(w, r, link.String(), http.StatusFound)
		})
		// docs.github.com/en/rest/webhooks/repos#create-a-repository-webhook
		mux.HandleFunc("/api/v3/repos/"+*srv.repo+"/hooks", func(w http.ResponseWriter, r *http.Request) {
			// retrieve the webhook url
			type options struct {
				Events []string `json:"events"`
				Config struct {
					URL    string `json:"url"`
					Secret string `json:"secret"`
				} `json:"config"`
			}
			var opts options
			if err := json.NewDecoder(r.Body).Decode(&opts); err != nil {
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
				return
			}
			// persist hook to the 'db'
			srv.HookEndpoint = &opts.Config.URL
			srv.HookEvents = opts.Events
			srv.HookSecret = &opts.Config.Secret

			hook := github.Hook{
				ID: otf.Int64(123),
			}
			out, err := json.Marshal(hook)
			require.NoError(t, err)
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			w.Write(out)
		})
		// https://docs.github.com/en/free-pro-team@latest/rest/reference/repos/#delete-a-repository-webhook
		mux.HandleFunc("/api/v3/repos/"+*srv.repo+"/hooks/123", func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case "GET":
				hook := github.Hook{
					ID:     otf.Int64(123),
					Events: srv.HookEvents,
					URL:    srv.HookEndpoint,
				}
				out, err := json.Marshal(hook)
				require.NoError(t, err)
				w.Header().Add("Content-Type", "application/json")
				w.Write(out)
			case "DELETE":
				// delete hook from 'db'
				srv.HookEndpoint = nil
				srv.HookEvents = nil
				srv.HookSecret = nil

				w.WriteHeader(http.StatusNoContent)
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		})
		// https://docs.github.com/en/rest/commits/statuses?apiVersion=2022-11-28#create-a-commit-status
		mux.HandleFunc("/api/v3/repos/"+*srv.repo+"/statuses/", func(w http.ResponseWriter, r *http.Request) {
			var commit github.StatusEvent
			if err := json.NewDecoder(r.Body).Decode(&commit); err != nil {
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
				return
			}

			// pass commit status to callback if one is registered
			if srv.statusCallback != nil {
				srv.statusCallback(&commit)
			}

			w.WriteHeader(http.StatusCreated)
		})
	}

	if srv.tarball != nil {
		mux.HandleFunc("/mytarball", func(w http.ResponseWriter, r *http.Request) {
			w.Write(srv.tarball)
		})
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("github server received request for non-existent path: %s", r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
	})

	srv.Server = httptest.NewTLSServer(mux)
	t.Cleanup(srv.Close)

	u, err := url.Parse(srv.URL)
	require.NoError(t, err)

	cfg := cloud.Config{
		Name:                "github",
		Hostname:            u.Host,
		Cloud:               &Cloud{},
		SkipTLSVerification: true,
	}
	return &srv, cfg
}

func WithUser(user *cloud.User) TestServerOption {
	return func(srv *TestServer) {
		srv.user = user
	}
}

func WithRepo(repo string) TestServerOption {
	return func(srv *TestServer) {
		srv.repo = &repo
	}
}

func WithRefs(refs ...string) TestServerOption {
	return func(srv *TestServer) {
		srv.refs = refs
	}
}

func WithArchive(tarball []byte) TestServerOption {
	return func(srv *TestServer) {
		srv.tarball = tarball
	}
}

func WithStatusCallback(callback statusCallback) TestServerOption {
	return func(srv *TestServer) {
		srv.statusCallback = callback
	}
}

func (s *TestServer) HasWebhook() bool {
	return s.HookEndpoint != nil && s.HookSecret != nil
}

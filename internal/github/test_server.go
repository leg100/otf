package github

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/google/go-github/v65/github"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/user"
	"github.com/leg100/otf/internal/vcs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

const (
	PushEvent   GithubEvent = "push"
	PullRequest GithubEvent = "pull_request"

	WebhookCreated webhookAction = iota
	WebhookUpdated
	WebhookDeleted
)

type (
	TestServer struct {
		// status updates received from otfd
		statuses chan *github.StatusEvent

		// webhook created/updated/deleted events channel
		WebhookEvents chan webhookEvent

		*httptest.Server
		*testdb
		mux        *http.ServeMux
		disableTLS bool
	}

	TestServerOption func(*TestServer)

	testdb struct {
		username      *user.Username
		repo          *vcs.Repo
		commit        *string
		defaultBranch *string
		tarball       []byte
		refs          []string
		webhook       *hook

		// pull request stub
		pullNumber string
		pullFiles  []string

		// url of server, only populated once server starts
		url *string
	}

	hook struct {
		secret string
		*github.Hook
	}

	// The name of the event sent in the X-Github-Event header
	GithubEvent string

	webhookAction int

	webhookEvent struct {
		Action webhookAction
		Hook   *hook
	}
)

func NewTestServer(t *testing.T, opts ...TestServerOption) (*TestServer, *url.URL) {
	srv := TestServer{
		testdb:        &testdb{},
		statuses:      make(chan *github.StatusEvent, 999),
		WebhookEvents: make(chan webhookEvent, 999),
		mux:           http.NewServeMux(),
	}
	for _, o := range opts {
		o(&srv)
	}

	srv.mux.HandleFunc("/login/oauth/authorize", func(w http.ResponseWriter, r *http.Request) {
		q := url.Values{}
		q.Add("state", r.URL.Query().Get("state"))
		q.Add("code", internal.GenerateRandomString(10))

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
	srv.mux.HandleFunc("/login/oauth/access_token", func(w http.ResponseWriter, r *http.Request) {
		out, err := json.Marshal(&oauth2.Token{AccessToken: "stub_token"})
		require.NoError(t, err)
		w.Header().Add("Content-Type", "application/json")
		w.Write(out)
	})
	if srv.username != nil {
		srv.mux.HandleFunc("/api/v3/user", func(w http.ResponseWriter, r *http.Request) {
			out, err := json.Marshal(&github.User{Login: new(srv.username.String())})
			require.NoError(t, err)
			w.Header().Add("Content-Type", "application/json")
			w.Write(out)
		})
	}
	srv.mux.HandleFunc("/api/v3/user/repos", func(w http.ResponseWriter, r *http.Request) {
		repos := []*github.Repository{
			{
				Owner:         &github.User{Login: new(srv.repo.Owner())},
				Name:          new(srv.repo.Name()),
				DefaultBranch: srv.defaultBranch,
			},
		}
		out, err := json.Marshal(repos)
		require.NoError(t, err)
		w.Header().Add("Content-Type", "application/json")
		w.Write(out)
	})
	if srv.repo != nil {
		srv.mux.HandleFunc("/api/v3/repos/"+srv.repo.String()+"/git/matching-refs/", func(w http.ResponseWriter, r *http.Request) {
			var refs []*github.Reference
			for _, ref := range srv.refs {
				refs = append(refs, &github.Reference{Ref: new(ref)})
			}
			out, err := json.Marshal(refs)
			require.NoError(t, err)
			w.Header().Add("Content-Type", "application/json")
			w.Write(out)
		})
		srv.mux.HandleFunc("/api/v3/repos/"+srv.repo.String(), func(w http.ResponseWriter, r *http.Request) {
			repo := &github.Repository{
				Owner:         &github.User{Login: new(srv.repo.Owner())},
				Name:          new(srv.repo.Name()),
				DefaultBranch: srv.defaultBranch,
			}
			out, err := json.Marshal(repo)
			require.NoError(t, err)
			w.Header().Add("Content-Type", "application/json")
			w.Write(out)
		})
		srv.mux.HandleFunc("/api/v3/repos/"+srv.repo.String()+"/tarball/", func(w http.ResponseWriter, r *http.Request) {
			link := url.URL{Scheme: "https", Host: r.Host, Path: "/mytarball"}
			http.Redirect(w, r, link.String(), http.StatusFound)
		})
		srv.mux.HandleFunc("/api/v3/repos/"+srv.repo.String()+"/tarball", func(w http.ResponseWriter, r *http.Request) {
			link := url.URL{Scheme: "https", Host: r.Host, Path: "/mytarball"}
			http.Redirect(w, r, link.String(), http.StatusFound)
		})
		// https://docs.github.com/en/rest/webhooks/repos?apiVersion=2022-11-28#create-a-repository-webhook
		srv.mux.HandleFunc("/api/v3/repos/"+srv.repo.String()+"/hooks", func(w http.ResponseWriter, r *http.Request) {
			var opts struct {
				Events []string `json:"events"`
				Config struct {
					URL    string `json:"url"`
					Secret string `json:"secret"`
				} `json:"config"`
			}
			if err := json.NewDecoder(r.Body).Decode(&opts); err != nil {
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
				return
			}
			// persist hook to the 'db'
			srv.testdb.webhook = &hook{
				Hook: &github.Hook{
					ID:     new(int64(123)),
					Events: opts.Events,
					Config: &github.HookConfig{
						URL: &opts.Config.URL,
					},
				},
				secret: opts.Config.Secret,
			}

			// notify tests
			srv.WebhookEvents <- webhookEvent{
				Action: WebhookCreated,
				Hook:   srv.testdb.webhook,
			}

			out, err := json.Marshal(srv.testdb.webhook)
			require.NoError(t, err)
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			w.Write(out)
		})
		// https://docs.github.com/en/rest/webhooks/repos?apiVersion=2022-11-28#get-a-repository-webhook
		// https://docs.github.com/en/rest/webhooks/repos?apiVersion=2022-11-28#update-a-repository-webhook
		// https://docs.github.com/en/rest/webhooks/repos?apiVersion=2022-11-28#delete-a-repository-webhook
		srv.mux.HandleFunc("/api/v3/repos/"+srv.repo.String()+"/hooks/123", func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case "PATCH":
				var opts struct {
					Events []string `json:"events"`
					Config struct {
						URL    string `json:"url"`
						Secret string `json:"secret"`
					} `json:"config"`
				}
				if err := json.NewDecoder(r.Body).Decode(&opts); err != nil {
					http.Error(w, err.Error(), http.StatusUnprocessableEntity)
					return
				}
				// persist hook to the 'db'
				srv.testdb.webhook = &hook{
					Hook: &github.Hook{
						ID:     new(int64(123)),
						Events: opts.Events,
						Config: &github.HookConfig{
							URL: &opts.Config.URL,
						},
					},
					secret: opts.Config.Secret,
				}
				// notify tests
				srv.WebhookEvents <- webhookEvent{
					Action: WebhookUpdated,
					Hook:   srv.testdb.webhook,
				}
				fallthrough
			case "GET":
				out, err := json.Marshal(srv.testdb.webhook)
				require.NoError(t, err)
				w.Header().Add("Content-Type", "application/json")
				w.Write(out)
			case "DELETE":
				// notify tests
				srv.WebhookEvents <- webhookEvent{
					Action: WebhookDeleted,
					Hook:   srv.testdb.webhook,
				}

				// delete hook from 'db'
				srv.testdb.webhook = nil

				w.WriteHeader(http.StatusNoContent)
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		})
		// https://docs.github.com/en/rest/commits/statuses?apiVersion=2022-11-28#create-a-commit-status
		srv.mux.HandleFunc("/api/v3/repos/"+srv.repo.String()+"/statuses/", func(w http.ResponseWriter, r *http.Request) {
			var commit github.StatusEvent
			if err := json.NewDecoder(r.Body).Decode(&commit); err != nil {
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
				return
			}
			srv.statuses <- &commit
			w.WriteHeader(http.StatusCreated)
		})
		// https://docs.github.com/en/rest/pulls/pulls?apiVersion=2022-11-28#list-pull-requests-files
		srv.mux.HandleFunc("/api/v3/repos/"+srv.repo.String()+"/pulls/"+srv.pullNumber+"/files", func(w http.ResponseWriter, r *http.Request) {
			var commits []*github.CommitFile
			for _, f := range srv.pullFiles {
				commits = append(commits, &github.CommitFile{
					Filename: new(f),
				})
			}
			out, err := json.Marshal(commits)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
				return
			}
			w.Header().Add("Content-Type", "application/json")
			w.Write(out)
		})
		if srv.commit != nil {
			// https://docs.github.com/en/rest/commits/commits?apiVersion=2022-11-28#get-a-commit
			srv.mux.HandleFunc("/api/v3/repos/"+srv.repo.String()+"/commits/"+*srv.commit, func(w http.ResponseWriter, r *http.Request) {
				out, err := json.Marshal(&github.Commit{
					SHA: new(*srv.commit),
					URL: new(*srv.url + "/" + srv.repo.String()),
					Author: &github.CommitAuthor{
						Login: new("leg100"),
					},
				})
				if err != nil {
					http.Error(w, err.Error(), http.StatusUnprocessableEntity)
					return
				}
				w.Header().Add("Content-Type", "application/json")
				w.Write(out)
			})
		}
	}

	if srv.tarball != nil {
		srv.mux.HandleFunc("/mytarball", func(w http.ResponseWriter, r *http.Request) {
			w.Write(srv.tarball)
		})
	}

	srv.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t.Logf("github server received request for non-existent path: %s", r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
	})

	srv.Server = httptest.NewUnstartedServer(srv.mux)
	srv.Server.Config.ErrorLog = log.New(io.Discard, "", 0)
	if srv.disableTLS {
		srv.Server.Start()
	} else {
		srv.Server.StartTLS()
	}
	t.Cleanup(srv.Close)
	srv.url = &srv.URL

	u, err := url.Parse(srv.URL)
	require.NoError(t, err)
	return &srv, u
}

func WithUsername(username user.Username) TestServerOption {
	return func(srv *TestServer) {
		srv.username = &username
	}
}

func WithRepo(repo vcs.Repo) TestServerOption {
	return func(srv *TestServer) {
		srv.repo = &repo
	}
}

func WithHook(hook hook) TestServerOption {
	return func(srv *TestServer) {
		srv.webhook = &hook
	}
}

func WithCommit(commit string) TestServerOption {
	return func(srv *TestServer) {
		srv.commit = &commit
	}
}

func WithDefaultBranch(branch string) TestServerOption {
	return func(srv *TestServer) {
		srv.defaultBranch = &branch
	}
}

func WithPullRequest(pullNumber string, changedPaths ...string) TestServerOption {
	return func(srv *TestServer) {
		srv.pullNumber = pullNumber
		srv.pullFiles = changedPaths
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

func WithHandler(path string, h http.HandlerFunc) TestServerOption {
	return func(srv *TestServer) {
		srv.mux.HandleFunc(path, h)
	}
}

func WithDisableTLS() TestServerOption {
	return func(srv *TestServer) {
		srv.disableTLS = true
	}
}

func (s *TestServer) HasWebhook() bool {
	return s.testdb.webhook != nil
}

// SendEvent sends an event to the registered webhook.
func (s *TestServer) SendEvent(t *testing.T, event GithubEvent, payload []byte) {
	t.Helper()

	require.True(t, s.HasWebhook())
	SendEventRequest(t, event, *s.testdb.webhook.Config.URL, s.testdb.webhook.secret, payload)
}

// GetStatus retrieves a commit status event off the queue, timing out after 10
// seconds if nothing is on the queue.
func (s *TestServer) GetStatus(t *testing.T, ctx context.Context) *github.StatusEvent {
	t.Helper()

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	select {
	case status := <-s.statuses:
		return status
	case <-ctx.Done():
		t.Fatalf("github server: waiting to receive commit status: %s", ctx.Err().Error())
	}
	return nil
}

// SendEventRequest sends a GitHub event via a http request to the url, signed with the secret,
func SendEventRequest(t *testing.T, event GithubEvent, url, secret string, payload []byte) {
	t.Helper()

	// generate signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	sig := mac.Sum(nil)

	req, err := http.NewRequest("POST", url, bytes.NewReader(payload))
	require.NoError(t, err)
	req.Header.Add("Content-type", "application/json")
	req.Header.Add("X-GitHub-Event", string(event))
	req.Header.Add("X-Hub-Signature-256", "sha256="+hex.EncodeToString(sig))

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	if !assert.Equal(t, http.StatusOK, res.StatusCode) {
		response, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		t.Fatal(string(response))
	}
}

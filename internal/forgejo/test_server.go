package forgejo

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/user"
	"github.com/leg100/otf/internal/vcs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type (
	TestServer struct {
		version string
		// webhook created/updated/deleted events channel
		WebhookEvents chan webhookEvent
		statuses      chan *forgejo.Status

		*httptest.Server
		*testdb
		mux        *http.ServeMux
		disableTLS bool
	}

	TestServerOption func(*TestServer)

	testref struct {
		ref    string
		object string
	}
	testdb struct {
		username      *user.Username
		repo          *vcs.Repo
		commit        *string
		defaultBranch string
		tarball       []byte
		refs          []testref
		webhook       *hook

		// pull request stub
		pullNumber string
		pullFiles  []string

		// url of server, only populated once server starts
		url *string
	}

	hook struct {
		secret string
		*forgejo.Hook
	}

	// The name of the event sent in the X-Github-Event header
	GithubEvent string

	// webhookAction int

	webhookEvent struct {
		Action vcs.Action
		Hook   *hook
	}
)

// newTestServerClient creates a github server for testing purposes and
// returns a client configured to access the server.
func newTestServerClientPair(t *testing.T, opts ...TestServerOption) (*Client, *TestServer) {
	s, u := NewTestServer(t, opts...)

	client, err := NewTokenClient(vcs.NewTokenClientOptions{
		BaseURL:             &internal.WebURL{URL: *u},
		SkipTLSVerification: true,
	})
	require.NoError(t, err)

	return client.(*Client), s
}

func newTestServerClient(t *testing.T, opts ...TestServerOption) *Client {
	c, _ := newTestServerClientPair(t, opts...)
	return c
}

func NewTestServer(t *testing.T, opts ...TestServerOption) (*TestServer, *url.URL) {
	srv := TestServer{
		version:       "11.0.1+gitea-1.22.0",
		testdb:        &testdb{},
		WebhookEvents: make(chan webhookEvent, 999),
		statuses:      make(chan *forgejo.Status, 999),
		mux:           http.NewServeMux(),
	}
	for _, o := range opts {
		o(&srv)
	}

	// general endpoints

	srv.mux.HandleFunc("/api/v1/version", func(w http.ResponseWriter, r *http.Request) {
		out, err := json.Marshal(map[string]string{"version": srv.version})
		require.NoError(t, err)
		w.Header().Add("Content-Type", "application/json")
		_, _ = w.Write(out)
	})

	// repo endpoints

	if srv.repo != nil {
		// SearchRepos
		srv.mux.HandleFunc("/api/v1/user/repos", func(w http.ResponseWriter, r *http.Request) {
			repos := []*forgejo.Repository{{
				Owner:       &forgejo.User{UserName: srv.repo.Owner()},
				Name:        srv.repo.Name(),
				Permissions: &forgejo.Permission{Admin: true},
				Updated:     time.Now(),
			}}
			out, err := json.Marshal(repos)
			require.NoError(t, err)
			w.Header().Add("Content-Type", "application/json")
			_, _ = w.Write(out)
		})

		// GetRepo
		srv.mux.HandleFunc("/api/v1/repos/"+srv.repo.String(), func(w http.ResponseWriter, r *http.Request) {
			repo := &forgejo.Repository{
				Owner:         &forgejo.User{UserName: srv.repo.Owner()},
				Name:          srv.repo.Name(),
				DefaultBranch: srv.defaultBranch,
			}
			out, err := json.Marshal(repo)
			require.NoError(t, err)
			w.Header().Add("Content-Type", "application/json")
			_, _ = w.Write(out)
		})

		// CreateRepoHook
		srv.mux.HandleFunc("/api/v1/repos/"+srv.repo.String()+"/hooks", func(w http.ResponseWriter, r *http.Request) {
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
			srv.webhook = &hook{
				Hook: &forgejo.Hook{
					ID:     123,
					Events: opts.Events,
					Config: map[string]string{
						"content_type": "json",
						"url":          opts.Config.URL,
						"secret":       opts.Config.Secret,
					},
				},
				secret: opts.Config.Secret,
			}

			// notify tests
			srv.WebhookEvents <- webhookEvent{
				Action: vcs.ActionCreated,
				Hook:   srv.webhook,
			}

			out, err := json.Marshal(srv.webhook)
			require.NoError(t, err)
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write(out)
		})

		// EditRepoHook,
		// GetRepoHook,
		// DeleteRepoHook
		srv.mux.HandleFunc("/api/v1/repos/"+srv.repo.String()+"/hooks/123", func(w http.ResponseWriter, r *http.Request) {
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
				srv.webhook = &hook{
					Hook: &forgejo.Hook{
						ID:     123,
						Events: opts.Events,
						Config: map[string]string{
							"content_type": "json",
							"url":          opts.Config.URL,
							"secret":       opts.Config.Secret,
						},
					},
					secret: opts.Config.Secret,
				}
				// notify tests
				srv.WebhookEvents <- webhookEvent{
					Action: vcs.ActionUpdated,
					Hook:   srv.webhook,
				}
				fallthrough
			case "GET":
				if srv.webhook == nil {
					w.WriteHeader(http.StatusNotFound)
				} else {
					out, err := json.Marshal(srv.webhook)
					require.NoError(t, err)
					w.Header().Add("Content-Type", "application/json")
					_, _ = w.Write(out)
				}
			case "DELETE":
				// notify tests
				srv.WebhookEvents <- webhookEvent{
					Action: vcs.ActionDeleted,
					Hook:   srv.webhook,
				}

				// delete hook from 'db'
				srv.webhook = nil

				w.WriteHeader(http.StatusNoContent)
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		})

		// GetRepoRefs
		srv.mux.HandleFunc("/api/v1/repos/"+srv.repo.String()+"/git/refs/", func(w http.ResponseWriter, r *http.Request) {
			var refs []*forgejo.Reference
			for _, ref := range srv.refs {
				refs = append(refs, &forgejo.Reference{
					Ref: ref.ref,
					Object: &forgejo.GitObject{
						SHA: ref.object,
					},
				})
			}
			out, err := json.Marshal(refs)
			require.NoError(t, err)
			w.Header().Add("Content-Type", "application/json")
			_, _ = w.Write(out)
		})

		// GetRepoTags
		srv.mux.HandleFunc("/api/v1/repos/"+srv.repo.String()+"/tags", func(w http.ResponseWriter, r *http.Request) {
			var tags []*forgejo.Tag
			for _, ref := range srv.refs {
				if strings.HasPrefix(ref.ref, "refs/tags/") {
					refname := strings.TrimPrefix(ref.ref, "refs/tags/")
					tags = append(tags, &forgejo.Tag{
						Name: refname,
						Commit: &forgejo.CommitMeta{
							SHA: ref.object,
						},
					})
				}
			}
			out, err := json.Marshal(tags)
			require.NoError(t, err)
			w.Header().Add("Content-Type", "application/json")
			_, _ = w.Write(out)
		})

		// GetArchive
		if srv.tarball != nil {
			srv.mux.HandleFunc("/api/v1/repos/"+srv.repo.String()+"/archive/"+srv.defaultBranch+".tar.gz", func(w http.ResponseWriter, r *http.Request) {
				link := url.URL{Scheme: "https", Host: r.Host, Path: "/mytarball"}
				http.Redirect(w, r, link.String(), http.StatusFound)
			})
			srv.mux.HandleFunc("/mytarball", func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write(srv.tarball)
			})
		}

		// CreateStatus
		if srv.commit != nil {
			srv.mux.HandleFunc("/api/v1/repos/"+srv.repo.String()+"/statuses/"+*srv.commit, func(w http.ResponseWriter, r *http.Request) {
				var opt forgejo.CreateStatusOption
				if err := json.NewDecoder(r.Body).Decode(&opt); err != nil {
					http.Error(w, err.Error(), http.StatusUnprocessableEntity)
					return
				}
				status := forgejo.Status{
					ID:          123,
					State:       opt.State,
					TargetURL:   opt.TargetURL,
					Description: opt.Description,
					Context:     opt.Context,
					Created:     time.Now(),
					Updated:     time.Now(),
				}
				srv.statuses <- &status
				status.ID = 123
				out, err := json.Marshal(status)
				require.NoError(t, err)
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write(out)
			})
		}

		// ListRepoTags

		// ListPullRequestFiles
		srv.mux.HandleFunc("/api/v1/repos/"+srv.repo.String()+"/pulls/"+srv.pullNumber+"/files", func(w http.ResponseWriter, r *http.Request) {
			var commits []*forgejo.ChangedFile
			for _, f := range srv.pullFiles {
				commits = append(commits, &forgejo.ChangedFile{
					Filename: f,
				})
			}
			out, err := json.Marshal(commits)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnprocessableEntity)
				return
			}
			w.Header().Add("Content-Type", "application/json")
			_, _ = w.Write(out)
		})

		// GetSingleCommit
		if srv.commit != nil {
			srv.mux.HandleFunc("/api/v1/repos/"+srv.repo.String()+"/git/commits/"+*srv.commit, func(w http.ResponseWriter, r *http.Request) {
				out, err := json.Marshal(&forgejo.Commit{
					CommitMeta: &forgejo.CommitMeta{
						SHA: *srv.commit,
						URL: *srv.url + "/" + srv.repo.String(),
					},
					Author: &forgejo.User{
						UserName: srv.username.String(),
					},
				})
				if err != nil {
					http.Error(w, err.Error(), http.StatusUnprocessableEntity)
					return
				}
				w.Header().Add("Content-Type", "application/json")
				_, _ = w.Write(out)
			})
		}
	}

	// user endpoints
	// GetMyUserInfo
	// ListMyTeams
	srv.mux.HandleFunc("/api/v1/user/teams", func(w http.ResponseWriter, r *http.Request) {
		out, err := json.Marshal([]*forgejo.Team{})
		require.NoError(t, err)
		w.Header().Add("Content-Type", "application/json")
		_, _ = w.Write(out)
	})

	if srv.disableTLS {
		srv.Server = httptest.NewServer(srv.mux)
	} else {
		srv.Server = httptest.NewTLSServer(srv.mux)
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
		srv.defaultBranch = branch
	}
}

func WithPullRequest(pullNumber string, changedPaths ...string) TestServerOption {
	return func(srv *TestServer) {
		srv.pullNumber = pullNumber
		srv.pullFiles = changedPaths
	}
}

func WithRefs(refs ...testref) TestServerOption {
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
	return s.webhook != nil
}

// SendEvent sends an event to the registered webhook.
func (s *TestServer) SendEvent(t *testing.T, event string, payload []byte) {
	t.Helper()

	require.True(t, s.HasWebhook())
	SendEventRequest(t, event, s.webhook.Config["url"], s.webhook.secret, payload)
}

// GetStatus retrieves a commit status event off the queue, timing out after 10
// seconds if nothing is on the queue.
func (s *TestServer) GetStatus(t *testing.T, ctx context.Context) *forgejo.Status {
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

// SendEventRequest sends a forgejo event via a http request to the url, signed with the secret,
func SendEventRequest(t *testing.T, event string, url, secret string, payload []byte) {
	t.Helper()

	// generate signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	sig := mac.Sum(nil)

	req, err := http.NewRequest("POST", url, bytes.NewReader(payload))
	require.NoError(t, err)
	req.Header.Add("Content-type", "application/json")
	req.Header.Add("X-Forgejo-Event", string(event))
	req.Header.Add("X-Gitea-Event", string(event))
	req.Header.Add("X-Forgejo-Signature", hex.EncodeToString(sig))
	req.Header.Add("X-Gitea-Signature", hex.EncodeToString(sig))

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	if !assert.Equal(t, http.StatusOK, res.StatusCode) {
		response, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		t.Fatal(string(response))
	}
}

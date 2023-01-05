package gitlab

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/stretchr/testify/require"
	"github.com/xanzy/go-gitlab"
	"golang.org/x/oauth2"
)

type testServerDB struct {
	user *gitlab.User
	// top-level group memberships
	groups []*gitlab.Group
	// group ID -> access level
	access map[int]gitlab.AccessLevelValue

	project *gitlab.Project
	tarball []byte
}

func NewTestServer(t *testing.T, opts ...TestGitlabServerOption) *httptest.Server {
	db := &testServerDB{}
	for _, o := range opts {
		o(db)
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
	if db.user != nil {
		mux.HandleFunc("/api/v4/user", func(w http.ResponseWriter, r *http.Request) {
			out, err := json.Marshal(db.user)
			require.NoError(t, err)
			w.Header().Add("Content-Type", "application/json")
			w.Write(out)
		})
		mux.HandleFunc("/api/v4/groups", func(w http.ResponseWriter, r *http.Request) {
			out, err := json.Marshal(db.groups)
			require.NoError(t, err)
			w.Header().Add("Content-Type", "application/json")
			w.Write(out)
		})
		for groupID, level := range db.access {
			path := fmt.Sprintf("/api/v4/groups/%d/members/1", groupID)
			mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
				out, err := json.Marshal(&gitlab.GroupMember{AccessLevel: level})
				require.NoError(t, err)
				w.Header().Add("Content-Type", "application/json")
				w.Write(out)
			})
		}
	}
	if db.project != nil {
		path := "/api/v4/projects/" + db.project.PathWithNamespace
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			out, err := json.Marshal(db.project)
			require.NoError(t, err)
			w.Header().Add("Content-Type", "application/json")
			w.Write(out)
		})
		mux.HandleFunc("/api/v4/projects", func(w http.ResponseWriter, r *http.Request) {
			out, err := json.Marshal([]*gitlab.Project{db.project})
			require.NoError(t, err)
			w.Header().Add("Content-Type", "application/json")
			w.Write(out)
		})
	}
	if db.project != nil && db.tarball != nil {
		path := "/api/v4/projects/" + db.project.PathWithNamespace + "/repository/archive.tar.gz"
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			w.Write(db.tarball)
		})
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)

		msg := fmt.Sprintf("gitlab server received request for non-existent path: %s", r.URL.Path)
		out, err := json.Marshal(map[string]string{"error": msg})
		require.NoError(t, err)
		w.Write(out)
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

type TestGitlabServerOption func(*testServerDB)

func WithGitlabUser(user *cloud.User) TestGitlabServerOption {
	return func(db *testServerDB) {
		db.user = &gitlab.User{Username: user.Name, ID: 1}
		db.access = make(map[int]gitlab.AccessLevelValue)

		for i, org := range user.Organizations {
			db.groups = append(db.groups, &gitlab.Group{
				ID:   i,
				Path: org,
			})
			// find team belonging to organization and map team name to gitlab
			// access level
			for i, team := range user.Teams {
				if team.Organization == org {
					switch team.Name {
					case "maintainers":
						db.access[i] = gitlab.MaintainerPermissions
					}
				}
			}
		}
	}
}

func WithGitlabRepo(repo *otf.Repo) TestGitlabServerOption {
	return func(db *testServerDB) {
		db.project = &gitlab.Project{
			PathWithNamespace: repo.Identifier,
			DefaultBranch:     repo.Branch,
		}
	}
}

func WithGitlabTarball(tarball []byte) TestGitlabServerOption {
	return func(db *testServerDB) {
		db.tarball = tarball
	}
}

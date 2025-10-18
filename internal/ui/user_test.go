package ui

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/team"
	"github.com/leg100/otf/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserHandlers(t *testing.T) {
	t.Run("new token", func(t *testing.T) {
		h := &userHandlers{}
		r := httptest.NewRequest("GET", "/?", nil)
		w := httptest.NewRecorder()

		h.newUserToken(w, r)

		assert.Equal(t, 200, w.Code, w.Body.String())
	})

	t.Run("create token", func(t *testing.T) {
		h := userHandlers{users: &fakeUserService{}}
		r := httptest.NewRequest("GET", "/?", nil)
		r = r.WithContext(authz.AddSubjectToContext(context.Background(), user.NewTestUser(t)))
		w := httptest.NewRecorder()

		h.createUserToken(w, r)

		if assert.Equal(t, 302, w.Code) {
			redirect, _ := w.Result().Location()
			assert.Equal(t, paths.Tokens(), redirect.Path)
		}
	})

	t.Run("list tokens", func(t *testing.T) {
		h := userHandlers{
			users: &fakeUserService{
				ut: &user.UserToken{},
			},
		}
		r := httptest.NewRequest("GET", "/?", nil)
		r = r.WithContext(authz.AddSubjectToContext(context.Background(), user.NewTestUser(t)))
		w := httptest.NewRecorder()

		h.userTokens(w, r)

		assert.Equal(t, 200, w.Code, w.Body.String())
	})

	t.Run("delete token", func(t *testing.T) {
		h := userHandlers{
			users: &fakeUserService{},
		}
		r := httptest.NewRequest("POST", "/?id=token-123", nil)
		r = r.WithContext(authz.AddSubjectToContext(context.Background(), user.NewTestUser(t)))
		w := httptest.NewRecorder()

		h.deleteUserToken(w, r)

		if assert.Equal(t, 302, w.Code) {
			redirect, _ := w.Result().Location()
			assert.Equal(t, paths.Tokens(), redirect.Path)
		}
	})
}

// TestUser_TeamGetHandler tests the getTeam handler. The getTeam page renders
// permissions only if the authenticated user is an owner, so the test sets that
// up first.
func TestUser_TeamGetHandler(t *testing.T) {
	org1 := organization.NewTestName(t)
	owners := &team.Team{Name: "owners", Organization: org1}
	owner, err := user.NewUser(uuid.NewString(), user.WithTeams(owners))
	require.NoError(t, err)
	h := &userHandlers{
		authorizer: authz.NewAllowAllAuthorizer(),
		teams:      &fakeTeamService{team: owners},
		users:      &fakeUserService{user: owner},
	}

	q := "/?team_id=team-123"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()
	h.getTeam(w, r)
	if !assert.Equal(t, 200, w.Code) {
		t.Log(t, w.Body.String())
	}
}

func TestAdminLoginHandler(t *testing.T) {
	h := &userHandlers{
		siteToken: "secrettoken",
		tokens:    &fakeTokensService{},
	}

	tests := []struct {
		name         string
		token        string
		wantRedirect string
	}{
		{
			name:         "valid token",
			token:        "secrettoken",
			wantRedirect: "/app/profile",
		},
		{
			name:         "invalid token",
			token:        "badtoken",
			wantRedirect: "/admin/login",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := strings.NewReader(url.Values{
				"token": {tt.token},
			}.Encode())

			r := httptest.NewRequest("POST", "/admin/login", form)
			r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			r.Header.Add("Referer", "http://otf.server/admin/login")
			w := httptest.NewRecorder()
			h.adminLogin(w, r)

			if assert.Equal(t, 302, w.Code) {
				redirect, err := w.Result().Location()
				require.NoError(t, err)
				assert.Equal(t, tt.wantRedirect, redirect.Path)
			}
		})
	}
}

func TestUser_diffUsers(t *testing.T) {
	alice := user.NewTestUser(t)
	bob := user.NewTestUser(t)
	a := []*user.User{bob}
	b := []*user.User{alice, bob}
	assert.Equal(t, []*user.User{alice}, diffUsers(a, b))
}

type fakeTokensService struct{}

func (f *fakeTokensService) StartSession(w http.ResponseWriter, r *http.Request, userID resource.TfeID) error {
	http.Redirect(w, r, paths.Profile(), http.StatusFound)
	return nil
}

type fakeUserService struct {
	user  *user.User
	token []byte
	ut    *user.UserToken

	*user.Service
}

func (f *fakeUserService) Create(context.Context, string, ...user.NewUserOption) (*user.User, error) {
	return f.user, nil
}

func (f *fakeUserService) List(ctx context.Context) ([]*user.User, error) {
	return []*user.User{f.user}, nil
}

func (f *fakeUserService) ListTeamUsers(ctx context.Context, teamID resource.TfeID) ([]*user.User, error) {
	return []*user.User{f.user}, nil
}

func (f *fakeUserService) Delete(context.Context, user.Username) error {
	return nil
}

func (f *fakeUserService) AddTeamMembership(context.Context, resource.TfeID, []user.Username) error {
	return nil
}

func (f *fakeUserService) RemoveTeamMembership(context.Context, resource.TfeID, []user.Username) error {
	return nil
}

func (f *fakeUserService) CreateToken(context.Context, user.CreateUserTokenOptions) (*user.UserToken, []byte, error) {
	return nil, f.token, nil
}

func (f *fakeUserService) ListTokens(context.Context) ([]*user.UserToken, error) {
	return []*user.UserToken{f.ut}, nil
}

func (f *fakeUserService) DeleteToken(context.Context, resource.TfeID) error {
	return nil
}

package ui

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/ui/paths"
	"github.com/leg100/otf/internal/user"
	"github.com/stretchr/testify/assert"
)

func TestUserHandlers(t *testing.T) {
	t.Run("new token", func(t *testing.T) {
		h := &Handlers{}
		r := httptest.NewRequest("GET", "/?", nil)
		w := httptest.NewRecorder()

		h.newUserToken(w, r)

		assert.Equal(t, 200, w.Code, w.Body.String())
	})

	t.Run("create token", func(t *testing.T) {
		h := Handlers{Users: &fakeUserService{}}
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
		h := Handlers{
			Users: &fakeUserService{
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
		h := Handlers{
			Users: &fakeUserService{},
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
	UserService
	user  *user.User
	token []byte
	ut    *user.UserToken
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

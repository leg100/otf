package ui

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/ui/helpers"
	"github.com/leg100/otf/internal/path"
	"github.com/leg100/otf/internal/ui/session"
	"github.com/leg100/otf/internal/user"
)

type Handlers struct {
	Client Client
}

type Client interface {
	Create(ctx context.Context, username string, opts ...user.NewUserOption) (*user.User, error)
	List(ctx context.Context) ([]*user.User, error)
	ListOrganizationUsers(ctx context.Context, organization organization.Name) ([]*user.User, error)
	ListTeamUsers(ctx context.Context, teamID resource.TfeID) ([]*user.User, error)
	GetUser(ctx context.Context, spec user.UserSpec) (*user.User, error)
	Delete(ctx context.Context, username user.Username) error
	AddTeamMembership(ctx context.Context, teamID resource.TfeID, usernames []user.Username) error
	RemoveTeamMembership(ctx context.Context, teamID resource.TfeID, usernames []user.Username) error

	CreateToken(ctx context.Context, opts user.CreateUserTokenOptions) (*user.UserToken, []byte, error)
	ListTokens(ctx context.Context) ([]*user.UserToken, error)
	DeleteToken(ctx context.Context, tokenID resource.TfeID) error
}

// AddHandlers registers user UI handlers with the router
func (h *Handlers) AddHandlers(r *mux.Router) {
	r.HandleFunc("/logout", h.logout).Methods("POST")
	r.HandleFunc("/organizations/{name}/users", h.listOrganizationUsers).Methods("GET")
	r.HandleFunc("/profile", h.profileHandler).Methods("GET")
	r.HandleFunc("/admin", h.site).Methods("GET")

	// user tokens
	r.HandleFunc("/current-user/tokens", h.userTokens).Methods("GET")
	r.HandleFunc("/current-user/tokens/delete", h.deleteUserToken).Methods("POST")
	r.HandleFunc("/current-user/tokens/new", h.newUserToken).Methods("GET")
	r.HandleFunc("/current-user/tokens/create", h.createUserToken).Methods("POST")

	// team membership
	r.HandleFunc("/teams/{team_id}/add-member", h.addTeamMember).Methods("POST")
	r.HandleFunc("/teams/{team_id}/remove-member", h.removeTeamMember).Methods("POST")

	// terraform login opens a browser to this hardcoded URL
	r.HandleFunc("/settings/tokens", h.userTokens).Methods("GET")
}

func (h *Handlers) logout(w http.ResponseWriter, r *http.Request) {
	helpers.SetCookie(w, session.SessionCookie, "", &time.Time{})
	http.Redirect(w, r, "/login", http.StatusFound)
}

func (h *Handlers) listOrganizationUsers(w http.ResponseWriter, r *http.Request) {
	var params user.ListOptions
	if err := decode.All(&params, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	users, err := h.Client.ListOrganizationUsers(r.Context(), params.Organization)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	props := userListProps{
		organization: params.Organization,
		users:        resource.NewPage(users, params.PageOptions, nil),
	}
	helpers.RenderPage(
		userList(props),
		"users",
		w,
		r,
		helpers.WithOrganization(params.Organization),
		helpers.WithBreadcrumbs(helpers.Breadcrumb{Name: "Users"}),
	)
}

func (h *Handlers) profileHandler(w http.ResponseWriter, r *http.Request) {
	helpers.RenderPage(
		profile(),
		"profile",
		w,
		r,
		helpers.WithBreadcrumbs(helpers.Breadcrumb{Name: "Profile"}),
		helpers.WithContentActions(profileActions()),
	)
}

func (h *Handlers) site(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, path.List(resource.OrganizationKind, nil), http.StatusFound)
}

// team membership handlers

func (h *Handlers) addTeamMember(w http.ResponseWriter, r *http.Request) {
	var params struct {
		TeamID   resource.TfeID `schema:"team_id,required"`
		Username *user.Username `schema:"username,required"`
	}
	if err := decode.All(&params, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	err := h.Client.AddTeamMembership(r.Context(), params.TeamID, []user.Username{*params.Username})
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	helpers.FlashSuccess(w, fmt.Sprintf("added team member: %s", *params.Username))
	http.Redirect(w, r, path.Get(params.TeamID), http.StatusFound)
}

func (h *Handlers) removeTeamMember(w http.ResponseWriter, r *http.Request) {
	var params struct {
		TeamID   resource.TfeID `schema:"team_id,required"`
		Username user.Username  `schema:"username,required"`
	}
	if err := decode.All(&params, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	err := h.Client.RemoveTeamMembership(r.Context(), params.TeamID, []user.Username{params.Username})
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	helpers.FlashSuccess(w, fmt.Sprintf("removed team member: %s", params.Username))
	http.Redirect(w, r, path.Get(params.TeamID), http.StatusFound)
}

//
// User tokens
//

func (h *Handlers) newUserToken(w http.ResponseWriter, r *http.Request) {
	helpers.RenderPage(
		newToken(),
		"new user token",
		w,
		r,
		helpers.WithBreadcrumbs(
			helpers.Breadcrumb{Name: "user tokens", Link: path.Tokens()},
			helpers.Breadcrumb{Name: "new"},
		),
	)
}

func (h *Handlers) createUserToken(w http.ResponseWriter, r *http.Request) {
	var opts user.CreateUserTokenOptions
	if err := decode.Form(&opts, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	_, token, err := h.Client.CreateToken(r.Context(), opts)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	if err := helpers.TokenFlashMessage(w, token); err != nil {
		helpers.Error(r, w, err.Error())
		return
	}
	http.Redirect(w, r, path.Tokens(), http.StatusFound)
}

func (h *Handlers) userTokens(w http.ResponseWriter, r *http.Request) {
	tokens, err := h.Client.ListTokens(r.Context())
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	// re-order tokens by creation date, newest first
	sort.Slice(tokens, func(i, j int) bool {
		return tokens[i].CreatedAt.After(tokens[j].CreatedAt)
	})

	helpers.RenderPage(
		tokenList(tokens),
		"User tokens",
		w,
		r,
		helpers.WithBreadcrumbs(helpers.Breadcrumb{Name: "User tokens"}),
		helpers.WithContentActions(tokenListActions()),
	)
}

func (h *Handlers) deleteUserToken(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("id", r)
	if err != nil {
		helpers.Error(r, w, "missing id", helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	if err := h.Client.DeleteToken(r.Context(), id); err != nil {
		helpers.Error(r, w, err.Error())
		return
	}
	helpers.FlashSuccess(w, "Deleted token")
	http.Redirect(w, r, path.Tokens(), http.StatusFound)
}

// diffUsers returns the users from b that are not in a.
func diffUsers(a, b []*user.User) (c []*user.User) {
	m := make(map[user.Username]struct{}, len(a))
	for _, user := range a {
		m[user.Username] = struct{}{}
	}
	for _, user := range b {
		if _, ok := m[user.Username]; !ok {
			c = append(c, user)
		}
	}
	return
}

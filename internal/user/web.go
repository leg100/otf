package user

import (
	"context"
	"net/http"
	"sort"
	"time"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/resource"
	otfteam "github.com/leg100/otf/internal/team"
	"github.com/leg100/otf/internal/tokens"
)

// webHandlers provides handlers for the web UI
type webHandlers struct {
	html.Renderer

	users     usersClient
	teams     teamsClient
	tokens    tokensClient
	siteToken string
}

type usersClient interface {
	Create(ctx context.Context, username string, opts ...NewUserOption) (*User, error)
	List(ctx context.Context) ([]*User, error)
	ListOrganizationUsers(ctx context.Context, organization string) ([]*User, error)
	ListTeamUsers(ctx context.Context, teamID resource.ID) ([]*User, error)
	Delete(ctx context.Context, username string) error
	AddTeamMembership(ctx context.Context, teamID resource.ID, usernames []string) error
	RemoveTeamMembership(ctx context.Context, teamID resource.ID, usernames []string) error

	CreateToken(ctx context.Context, opts CreateUserTokenOptions) (*UserToken, []byte, error)
	ListTokens(ctx context.Context) ([]*UserToken, error)
	DeleteToken(ctx context.Context, tokenID resource.ID) error
}

type tokensClient interface {
	StartSession(w http.ResponseWriter, r *http.Request, opts tokens.StartSessionOptions) error
}

type teamsClient interface {
	GetByID(ctx context.Context, teamID resource.ID) (*otfteam.Team, error)
}

func (h *webHandlers) addHandlers(r *mux.Router) {
	// Unauthenticated routes
	r.HandleFunc("/admin/login", h.adminLoginPromptHandler).Methods("GET")
	r.HandleFunc("/admin/login", h.adminLogin).Methods("POST")

	// Authenticated routes
	r = html.UIRouter(r)

	r.HandleFunc("/logout", h.logout).Methods("POST")

	r.HandleFunc("/organizations/{name}/users", h.listOrganizationUsers).Methods("GET")
	r.HandleFunc("/profile", h.profileHandler).Methods("GET")
	r.HandleFunc("/admin", h.site).Methods("GET")

	// user tokens
	r.HandleFunc("/profile/tokens", h.userTokens).Methods("GET")
	r.HandleFunc("/profile/tokens/delete", h.deleteUserToken).Methods("POST")
	r.HandleFunc("/profile/tokens/new", h.newUserToken).Methods("GET")
	r.HandleFunc("/profile/tokens/create", h.createUserToken).Methods("POST")

	// team membership
	r.HandleFunc("/teams/{team_id}/add-member", h.addTeamMember).Methods("POST")
	r.HandleFunc("/teams/{team_id}/remove-member", h.removeTeamMember).Methods("POST")
	// NOTE: to avoid an import cycle the getTeam handler is located here rather
	// than in the team package where it ought to belong.the
	r.HandleFunc("/teams/{team_id}", h.getTeam).Methods("GET")

	// terraform login opens a browser to this hardcoded URL
	r.HandleFunc("/settings/tokens", h.userTokens).Methods("GET")
}

func (h *webHandlers) logout(w http.ResponseWriter, r *http.Request) {
	html.SetCookie(w, tokens.SessionCookie, "", &time.Time{})
	http.Redirect(w, r, "/login", http.StatusFound)
}

func (h *webHandlers) listOrganizationUsers(w http.ResponseWriter, r *http.Request) {
	name, err := decode.Param("name", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	users, err := h.users.ListOrganizationUsers(r.Context(), name)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("users_list.tmpl", w, struct {
		html.SitePage
		Users []*User
	}{
		SitePage: html.NewSitePage(r, "users"),
		Users:    users,
	})
}

func (h *webHandlers) profileHandler(w http.ResponseWriter, r *http.Request) {
	user, err := authz.SubjectFromContext(r.Context())
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.Render("profile.tmpl", w, struct {
		html.SitePage
		User authz.Subject
	}{
		SitePage: html.NewSitePage(r, "profile"),
		User:     user,
	})
}

// adminLoginPromptHandler presents a prompt for logging in as site admin
func (h *webHandlers) adminLoginPromptHandler(w http.ResponseWriter, r *http.Request) {
	h.Render("site_admin_login.tmpl", w, html.NewSitePage(r, "site admin login"))
}

// adminLogin logs in a site admin
func (h *webHandlers) adminLogin(w http.ResponseWriter, r *http.Request) {
	token, err := decode.Param("token", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if token != h.siteToken {
		html.FlashError(w, "incorrect token")
		http.Redirect(w, r, paths.AdminLogin(), http.StatusFound)
		return
	}

	err = h.tokens.StartSession(w, r, tokens.StartSessionOptions{
		UserID: SiteAdminID,
	})
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *webHandlers) site(w http.ResponseWriter, r *http.Request) {
	user, err := authz.SubjectFromContext(r.Context())
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.Render("site.tmpl", w, struct {
		html.SitePage
		User authz.Subject
	}{
		SitePage: html.NewSitePage(r, "site"),
		User:     user,
	})
}

// team membership handlers

func (h *webHandlers) addTeamMember(w http.ResponseWriter, r *http.Request) {
	var params struct {
		TeamID   resource.ID `schema:"team_id,required"`
		Username *string     `schema:"username,required"`
	}
	if err := decode.All(&params, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err := h.users.AddTeamMembership(r.Context(), params.TeamID, []string{*params.Username})
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "added team member: "+*params.Username)
	http.Redirect(w, r, paths.Team(params.TeamID), http.StatusFound)
}

func (h *webHandlers) removeTeamMember(w http.ResponseWriter, r *http.Request) {
	var params struct {
		TeamID   resource.ID `schema:"team_id,required"`
		Username string      `schema:"username,required"`
	}
	if err := decode.All(&params, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err := h.users.RemoveTeamMembership(r.Context(), params.TeamID, []string{params.Username})
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "removed team member: "+params.Username)
	http.Redirect(w, r, paths.Team(params.TeamID), http.StatusFound)
}

func (h *webHandlers) getTeam(w http.ResponseWriter, r *http.Request) {
	teamID, err := decode.ID("team_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	team, err := h.teams.GetByID(r.Context(), teamID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// get usernames of team members
	members, err := h.users.ListTeamUsers(r.Context(), teamID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	usernames := make([]string, len(members))
	for i, m := range members {
		usernames[i] = m.Username
	}

	// Retrieve full list of users for populating a select form from which new
	// team members can be chosen. Only do this if the subject has perms to
	// retrieve the list.
	user, err := authz.SubjectFromContext(r.Context())
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// get usernames of non-members
	var nonMemberUsernames []string
	if user.CanAccessSite(rbac.ListUsersAction) {
		users, err := h.users.List(r.Context())
		if err != nil {
			h.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		nonMembers := diffUsers(members, users)
		nonMemberUsernames = make([]string, len(nonMembers))
		for i, m := range nonMembers {
			nonMemberUsernames[i] = m.Username
		}
	}

	h.Render("team_get.tmpl", w, struct {
		organization.OrganizationPage
		Team              *otfteam.Team
		Members           []*User
		AddMemberDropdown html.DropdownUI
		CanUpdateTeam     bool
		CanDeleteTeam     bool
		CanAddMember      bool
		CanRemoveMember   bool
		CanDelete         bool
		IsOwner           bool
	}{
		OrganizationPage: organization.NewPage(r, team.ID, team.Organization),
		Team:             team,
		Members:          members,
		CanUpdateTeam:    user.CanAccessOrganization(rbac.UpdateTeamAction, team.Organization),
		CanDeleteTeam:    user.CanAccessOrganization(rbac.DeleteTeamAction, team.Organization),
		CanAddMember:     user.CanAccessOrganization(rbac.AddTeamMembershipAction, team.Organization),
		CanRemoveMember:  user.CanAccessOrganization(rbac.RemoveTeamMembershipAction, team.Organization),
		CanDelete:        user.CanAccessOrganization(rbac.DeleteTeamAction, team.Organization),
		IsOwner:          user.IsOwner(team.Organization),
		AddMemberDropdown: html.DropdownUI{
			Name:        "username",
			Available:   nonMemberUsernames,
			Existing:    usernames,
			Action:      paths.AddMemberTeam(team.ID),
			Placeholder: "Add user",
			Width:       html.WideDropDown,
		},
	})
}

//
// User tokens
//

func (h *webHandlers) newUserToken(w http.ResponseWriter, r *http.Request) {
	h.Render("token_new.tmpl", w, html.NewSitePage(r, "new user token"))
}

func (h *webHandlers) createUserToken(w http.ResponseWriter, r *http.Request) {
	var opts CreateUserTokenOptions
	if err := decode.Form(&opts, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	_, token, err := h.users.CreateToken(r.Context(), opts)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tokens.TokenFlashMessage(h, w, token); err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, paths.Tokens(), http.StatusFound)
}

func (h *webHandlers) userTokens(w http.ResponseWriter, r *http.Request) {
	tokens, err := h.users.ListTokens(r.Context())
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// re-order tokens by creation date, newest first
	sort.Slice(tokens, func(i, j int) bool {
		return tokens[i].CreatedAt.After(tokens[j].CreatedAt)
	})

	h.Render("user_token_list.tmpl", w, struct {
		html.SitePage
		// list template expects pagination object
		*resource.Pagination
		Items []*UserToken
	}{
		SitePage:   html.NewSitePage(r, "user tokens"),
		Pagination: &resource.Pagination{},
		Items:      tokens,
	})
}

func (h *webHandlers) deleteUserToken(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		h.Error(w, "missing id", http.StatusUnprocessableEntity)
		return
	}
	if err := h.users.DeleteToken(r.Context(), id); err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	html.FlashSuccess(w, "Deleted token")
	http.Redirect(w, r, paths.Tokens(), http.StatusFound)
}

// diffUsers returns the users from b that are not in a.
func diffUsers(a, b []*User) (c []*User) {
	m := make(map[string]struct{}, len(a))
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

package auth

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/rbac"
)

func (h *webHandlers) addTeamHandlers(r *mux.Router) {
	r.HandleFunc("/organizations/{organization_name}/teams", h.listTeams).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/teams/new", h.newTeam).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/teams/create", h.createTeam).Methods("POST")
	r.HandleFunc("/teams/{team_id}", h.getTeam).Methods("GET")
	r.HandleFunc("/teams/{team_id}/update", h.updateTeam).Methods("POST")
	r.HandleFunc("/teams/{team_id}/delete", h.deleteTeam).Methods("POST")
	r.HandleFunc("/teams/{team_id}/add-member", h.addTeamMember).Methods("POST")
	r.HandleFunc("/teams/{team_id}/remove-member", h.removeTeamMember).Methods("POST")
}

func (h *webHandlers) newTeam(w http.ResponseWriter, r *http.Request) {
	org, err := decode.Param("organization_name", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	h.Render("team_new.tmpl", w, struct {
		organization.OrganizationPage
	}{
		OrganizationPage: organization.NewPage(r, "new team", org),
	})
}

func (h *webHandlers) createTeam(w http.ResponseWriter, r *http.Request) {
	var opts CreateTeamOptions
	if err := decode.All(&opts, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	team, err := h.svc.CreateTeam(r.Context(), opts)
	if err == internal.ErrResourceAlreadyExists {
		html.FlashError(w, "team already exists: "+opts.Name)
		http.Redirect(w, r, paths.NewTeam(opts.Organization), http.StatusFound)
		return
	}
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "created team: "+team.Name)
	http.Redirect(w, r, paths.Team(team.ID), http.StatusFound)
}

func (h *webHandlers) getTeam(w http.ResponseWriter, r *http.Request) {
	teamID, err := decode.Param("team_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	team, err := h.svc.GetTeamByID(r.Context(), teamID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	members, err := h.svc.ListTeamMembers(r.Context(), teamID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Retrieve full list of users for populating a select form from which new
	// team members can be chosen. Only do this if the subject has perms to
	// retrieve the list.
	subject, err := internal.SubjectFromContext(r.Context())
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var users []*User
	if subject.CanAccessSite(rbac.ListUsersAction) {
		users, err = h.svc.ListUsers(r.Context())
		if err != nil {
			h.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	h.Render("team_get.tmpl", w, struct {
		organization.OrganizationPage
		Team                       *Team
		Members                    []*User
		NonMembers                 []*User
		AddTeamMembershipAction    rbac.Action
		RemoveTeamMembershipAction rbac.Action
		DeleteTeamAction           rbac.Action
	}{
		OrganizationPage:           organization.NewPage(r, team.ID, team.Organization),
		Team:                       team,
		NonMembers:                 diffUsers(members, users),
		Members:                    members,
		AddTeamMembershipAction:    rbac.AddTeamMembershipAction,
		RemoveTeamMembershipAction: rbac.RemoveTeamMembershipAction,
		DeleteTeamAction:           rbac.DeleteTeamAction,
	})
}

func (h *webHandlers) updateTeam(w http.ResponseWriter, r *http.Request) {
	var params struct {
		TeamID string `schema:"team_id,required"`
		UpdateTeamOptions
	}
	if err := decode.All(&params, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	team, err := h.svc.UpdateTeam(r.Context(), params.TeamID, params.UpdateTeamOptions)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "team permissions updated")
	http.Redirect(w, r, paths.Team(team.ID), http.StatusFound)
}

func (h *webHandlers) listTeams(w http.ResponseWriter, r *http.Request) {
	org, err := decode.Param("organization_name", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	teams, err := h.svc.ListTeams(r.Context(), org)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("team_list.tmpl", w, struct {
		organization.OrganizationPage
		Teams            []*Team
		CreateTeamAction rbac.Action
		DeleteTeamAction rbac.Action
	}{
		OrganizationPage: organization.NewPage(r, "teams", org),
		Teams:            teams,
		CreateTeamAction: rbac.CreateTeamAction,
		DeleteTeamAction: rbac.DeleteTeamAction,
	})
}

func (h *webHandlers) deleteTeam(w http.ResponseWriter, r *http.Request) {
	teamID, err := decode.Param("team_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	team, err := h.svc.GetTeamByID(r.Context(), teamID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = h.svc.DeleteTeam(r.Context(), teamID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "deleted team: "+team.Name)
	http.Redirect(w, r, paths.Teams(team.Organization), http.StatusFound)
}

func (h *webHandlers) addTeamMember(w http.ResponseWriter, r *http.Request) {
	var params TeamMembershipOptions
	if err := decode.All(&params, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if err := h.svc.AddTeamMembership(r.Context(), params); err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "added team member: "+params.Username)
	http.Redirect(w, r, paths.Team(params.TeamID), http.StatusFound)
}

func (h *webHandlers) removeTeamMember(w http.ResponseWriter, r *http.Request) {
	var params TeamMembershipOptions
	if err := decode.All(&params, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if err := h.svc.RemoveTeamMembership(r.Context(), params); err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "removed team member: "+params.Username)
	http.Redirect(w, r, paths.Team(params.TeamID), http.StatusFound)
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

package team

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/rbac"
)

// authorizer authorizes access to a team
type authorizer struct {
	logr.Logger
}

func (a *authorizer) CanAccess(ctx context.Context, action rbac.Action, teamID string) (authz.Subject, error) {
	subj, err := authz.SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if authz.SkipAuthz(ctx) {
		return subj, nil
	}
	if subj.CanAccessTeam(action, teamID) {
		return subj, nil
	}
	a.Error(nil, "unauthorized action", "team_id", teamID, "action", action.String(), "subject", subj)
	return nil, internal.ErrAccessNotPermitted
}

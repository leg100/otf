package auth

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/rbac"
)

// TeamAuthorizer authorizes access to a team
type TeamAuthorizer struct {
	logr.Logger
}

func (a *TeamAuthorizer) CanAccess(ctx context.Context, action rbac.Action, teamID string) (internal.Subject, error) {
	subj, err := internal.SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if subj.CanAccessTeam(action, teamID) {
		return subj, nil
	}
	a.Error(nil, "unauthorized action", "team_id", teamID, "action", action.String(), "subject", subj)
	return nil, internal.ErrAccessNotPermitted
}

package policy

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/runner"
	"github.com/leg100/otf/internal/team"
)

// authorizer authorizes access to a workspace according to individual workspace
// policy permissions
type authorizer struct {
	logr.Logger

	db *pgdb
}

func (a *authorizer) CanAccess(ctx context.Context, action rbac.Action, workspaceID resource.ID) (bool, error) {
	subj, err := authz.SubjectFromContext(ctx)
	if err != nil {
		return false, err
	}
	policy, err := a.db.GetWorkspacePolicy(ctx, workspaceID)
	if err != nil {
		return false, err
	}
	switch caller := subj.(type) {
	case *team.Team:
		// Team can only access workspace if a specific permission has been
		// assigned to the team.
		for _, perm := range policy.Permissions {
			if caller.ID == perm.TeamID {
				return perm.Role.IsAllowed(action), nil
			}
		}
	case *runner.Job:
		// NOTE: *runner.Job itself is responsible for authenticating access to
		// actions on the *same* workspace; whereas this authorizer handles
		// cases where the job is attempting to access resources on a
		// *different* workspace.
		if caller.Organization != policy.Organization {
			// Job cannot access workspace in different organization
			return false, nil
		}
		switch action {
		case rbac.GetStateVersionAction, rbac.GetWorkspaceAction, rbac.DownloadStateAction:
			if policy.GlobalRemoteState {
				// Job is allowed to retrieve the state of this workspace
				// because the workspace has allowed global remote state
				// sharing.
				return true, nil
			}
		}
	}
	// a.Error(nil, "unauthorized action", "workspace_id", workspaceID, "organization", policy.Organization, "action", action.String(), "subject", subj)
	return false, nil
}

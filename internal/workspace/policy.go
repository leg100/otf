package workspace

import (
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/resource"
)

type Policy struct {
	Permissions []Permission
	// Whether workspace permits its state to be consumed by all workspaces in
	// the organization.
	globalRemoteState bool
}

// Permission binds a role to a team.
type Permission struct {
	TeamID resource.TfeID
	Role   authz.Role
}

// Check whether the subject is allowed to carry out the given action on the
// workspace belonging to the policy.
func (p Policy) Check(subject resource.ID, action authz.Action) bool {
	switch subject.Kind() {
	case resource.TeamKind:
		// Team can only access workspace if a specific permission has been
		// assigned to the team.
		for _, perm := range p.Permissions {
			if subject == perm.TeamID {
				return perm.Role.IsAllowed(action)
			}
		}
	case resource.JobKind:
		// Job is allowed to retrieve the state of this workspace if the
		// workspace has allowed global remote state sharing.
		switch action {
		case authz.GetStateVersionAction, authz.GetWorkspaceAction, authz.DownloadStateAction:
			return p.globalRemoteState
		}
	}
	return false
}

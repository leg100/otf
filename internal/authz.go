package internal

import (
	"context"
	"fmt"

	"github.com/leg100/otf/internal/rbac"
)

// unexported key type prevents collisions
type subjectCtxKeyType string

const subjectCtxKey subjectCtxKeyType = "subject"

// Subject is an entity that carries out actions on resources.
type Subject interface {
	CanAccessSite(action rbac.Action) bool
	CanAccessTeam(action rbac.Action, name string) bool
	CanAccessOrganization(action rbac.Action, name string) bool
	CanAccessWorkspace(action rbac.Action, policy WorkspacePolicy) bool

	IsOwner(organization string) bool
	IsSiteAdmin() bool

	// Organizations returns subject's organization memberships
	Organizations() []string

	String() string
}

// WorkspacePolicy binds workspace permissions to a workspace
type WorkspacePolicy struct {
	Organization string
	WorkspaceID  string
	Permissions  []WorkspacePermission

	// Whether workspace permits its state to be consumed by all workspaces in
	// the organization.
	GlobalRemoteState bool
}

// WorkspacePermission binds a role to a team.
type WorkspacePermission struct {
	Team   string // team name
	TeamID string
	Role   rbac.Role
}

// AddSubjectToContext adds a subject to a context
func AddSubjectToContext(ctx context.Context, subj Subject) context.Context {
	return context.WithValue(ctx, subjectCtxKey, subj)
}

// SubjectFromContext retrieves a subject from a context
func SubjectFromContext(ctx context.Context) (Subject, error) {
	subj, ok := ctx.Value(subjectCtxKey).(Subject)
	if !ok {
		return nil, fmt.Errorf("no subject in context")
	}
	return subj, nil
}

// Superuser is a subject with unlimited privileges.
type Superuser struct {
	Username string
}

func (*Superuser) CanAccessSite(action rbac.Action) bool                { return true }
func (*Superuser) CanAccessTeam(rbac.Action, string) bool               { return true }
func (*Superuser) CanAccessOrganization(rbac.Action, string) bool       { return true }
func (*Superuser) CanAccessWorkspace(rbac.Action, WorkspacePolicy) bool { return true }
func (s *Superuser) Organizations() []string                            { return nil }
func (s *Superuser) String() string                                     { return s.Username }
func (s *Superuser) ID() string                                         { return s.Username }
func (s *Superuser) IsSiteAdmin() bool                                  { return true }
func (s *Superuser) IsOwner(string) bool                                { return true }

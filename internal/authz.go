package internal

import (
	"context"
	"fmt"

	"github.com/leg100/otf/internal/rbac"
)

// unexported key types prevents collisions
type (
	subjectCtxKeyType   string
	skipAuthzCtxKeyType string
)

const (
	subjectCtxKey   subjectCtxKeyType   = "subject"
	skipAuthzCtxKey skipAuthzCtxKeyType = "skip_authz"
)

// Subject is an entity that carries out actions on resources.
type Subject interface {
	CanAccessSite(action rbac.Action) bool
	CanAccessTeam(action rbac.Action, id string) bool
	CanAccessOrganization(action rbac.Action, name string) bool
	CanAccessWorkspace(action rbac.Action, policy WorkspacePolicy) bool

	IsOwner(organization string) bool
	IsSiteAdmin() bool

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

// AddSkipAuthz adds to the context an instruction to skip authorization.
// Authorizers should obey this instruction using SkipAuthz
func AddSkipAuthz(ctx context.Context) context.Context {
	return context.WithValue(ctx, skipAuthzCtxKey, "")
}

// SkipAuthz determines whether the context contains an instruction to skip
// authorization.
func SkipAuthz(ctx context.Context) bool {
	if v := ctx.Value(skipAuthzCtxKey); v != nil {
		return true
	}
	return false
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

// Nobody is a subject with no privileges.
type Nobody struct {
	Username string
}

func (*Nobody) CanAccessSite(action rbac.Action) bool                { return false }
func (*Nobody) CanAccessTeam(rbac.Action, string) bool               { return false }
func (*Nobody) CanAccessOrganization(rbac.Action, string) bool       { return false }
func (*Nobody) CanAccessWorkspace(rbac.Action, WorkspacePolicy) bool { return false }
func (s *Nobody) Organizations() []string                            { return nil }
func (s *Nobody) String() string                                     { return s.Username }
func (s *Nobody) ID() string                                         { return s.Username }
func (s *Nobody) IsSiteAdmin() bool                                  { return false }
func (s *Nobody) IsOwner(string) bool                                { return false }

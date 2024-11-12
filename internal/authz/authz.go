// Package authz handles all things authorization
package authz

import (
	"context"
	"fmt"

	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/resource"
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
	CanAccess(action rbac.Action, req *AccessRequest) bool

	IsOwner(organization string) bool
	IsSiteAdmin() bool

	Organizations() []string

	String() string
}

// WorkspacePolicy binds workspace permissions to a workspace
type WorkspacePolicy struct {
	Organization string
	WorkspaceID  resource.ID
	Permissions  []WorkspacePermission

	// Whether workspace permits its state to be consumed by all workspaces in
	// the organization.
	GlobalRemoteState bool
}

// WorkspacePermission binds a role to a team.
type WorkspacePermission struct {
	TeamID resource.ID
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

func (*Superuser) CanAccess(rbac.Action, *AccessRequest) bool { return true }
func (s *Superuser) Organizations() []string                  { return nil }
func (s *Superuser) String() string                           { return s.Username }
func (s *Superuser) GetID() resource.ID                       { return resource.NewID(resource.UserKind) }
func (s *Superuser) IsSiteAdmin() bool                        { return true }
func (s *Superuser) IsOwner(string) bool                      { return true }

// Nobody is a subject with no privileges.
type Nobody struct {
	Username string
}

func (*Nobody) CanAccess(rbac.Action, *AccessRequest) bool { return true }
func (s *Nobody) Organizations() []string                  { return nil }
func (s *Nobody) String() string                           { return s.Username }
func (s *Nobody) ID() string                               { return s.Username }
func (s *Nobody) IsSiteAdmin() bool                        { return false }
func (s *Nobody) IsOwner(string) bool                      { return false }

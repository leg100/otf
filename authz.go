package otf

import (
	"context"
	"fmt"
)

// unexported key type prevents collisions
type subjectCtxKeyType string

const subjectCtxKey subjectCtxKeyType = "subject"

// Subject is an entity that carries out actions on resources.
type Subject interface {
	CanAccessSite(action Action) bool
	CanAccessOrganization(action Action, name string) bool
	CanAccessWorkspace(action Action, policy *WorkspacePolicy) bool

	Identity
}

// WorkspacePolicy binds workspace permissions to a workspace
type WorkspacePolicy struct {
	OrganizationName string
	WorkspaceID      string
	Permissions      []*WorkspacePermission
}

// WorkspacePermission binds a role to a team.
type WorkspacePermission struct {
	Team *Team
	Role Role
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

// UserFromContext retrieves a user from a context
func UserFromContext(ctx context.Context) (*User, error) {
	subj, err := SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	user, ok := subj.(*User)
	if !ok {
		return nil, fmt.Errorf("subject found in context but it is not a user")
	}
	return user, nil
}

// AgentFromContext retrieves an agent(-token) from a context
func AgentFromContext(ctx context.Context) (*AgentToken, error) {
	subj, err := SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	agent, ok := subj.(*AgentToken)
	if !ok {
		return nil, fmt.Errorf("subject found in context but it is not an agent")
	}
	return agent, nil
}

// LockFromContext retrieves a workspace lock from a context
func LockFromContext(ctx context.Context) (WorkspaceLockState, error) {
	subj, err := SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	lock, ok := subj.(WorkspaceLockState)
	if !ok {
		return nil, fmt.Errorf("no lock subject in context")
	}
	return lock, nil
}

// Superuser is a subject with unlimited privileges.
type Superuser struct {
	Username string
}

func (*Superuser) CanAccessSite(action Action) bool                 { return true }
func (*Superuser) CanAccessOrganization(Action, string) bool        { return true }
func (*Superuser) CanAccessWorkspace(Action, *WorkspacePolicy) bool { return true }
func (s *Superuser) String() string                                 { return s.Username }
func (s *Superuser) ID() string                                     { return s.Username }

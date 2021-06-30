package ots

import (
	"errors"
	"fmt"

	tfe "github.com/leg100/go-tfe"
)

const (
	DefaultAllowDestroyPlan    = true
	DefaultFileTriggersEnabled = true
	DefaultTerraformVersion    = "0.15.4"
)

var (
	ErrWorkspaceAlreadyLocked   = errors.New("workspace already locked")
	ErrWorkspaceAlreadyUnlocked = errors.New("workspace already unlocked")
)

type WorkspaceService interface {
	CreateWorkspace(org string, opts *tfe.WorkspaceCreateOptions) (*tfe.Workspace, error)
	GetWorkspace(name, org string) (*tfe.Workspace, error)
	GetWorkspaceByID(id string) (*tfe.Workspace, error)
	ListWorkspaces(org string, opts tfe.WorkspaceListOptions) (*tfe.WorkspaceList, error)
	UpdateWorkspace(name, org string, opts *tfe.WorkspaceUpdateOptions) (*tfe.Workspace, error)
	UpdateWorkspaceByID(id string, opts *tfe.WorkspaceUpdateOptions) (*tfe.Workspace, error)
	DeleteWorkspace(name, org string) error
	DeleteWorkspaceByID(id string) error
	LockWorkspace(id string, opts tfe.WorkspaceLockOptions) (*tfe.Workspace, error)
	UnlockWorkspace(id string) (*tfe.Workspace, error)
}

func NewWorkspaceID() string {
	return fmt.Sprintf("ws-%s", GenerateRandomString(16))
}

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

type WorkspaceList struct {
	*Pagination
	Items []*tfe.Workspace
}

// WorkspaceListOptions represents the options for listing workspaces.
type WorkspaceListOptions struct {
	ListOptions

	// A search string (partial workspace name) used to filter the results.
	Search *string `schema:"search[name]"`

	// A list of relations to include. See available resources https://www.terraform.io/docs/cloud/api/workspaces.html#available-related-resources
	Include *string `schema:"include"`
}

// WorkspaceLockOptions represents the options for locking a workspace.
type WorkspaceLockOptions struct {
	// Specifies the reason for locking the workspace.
	Reason *string `jsonapi:"attr,reason,omitempty"`
}

type WorkspaceService interface {
	CreateWorkspace(org string, opts *tfe.WorkspaceCreateOptions) (*tfe.Workspace, error)
	GetWorkspace(name, org string) (*tfe.Workspace, error)
	GetWorkspaceByID(id string) (*tfe.Workspace, error)
	ListWorkspaces(org string, opts WorkspaceListOptions) (*WorkspaceList, error)
	UpdateWorkspace(name, org string, opts *tfe.WorkspaceUpdateOptions) (*tfe.Workspace, error)
	UpdateWorkspaceByID(id string, opts *tfe.WorkspaceUpdateOptions) (*tfe.Workspace, error)
	DeleteWorkspace(name, org string) error
	DeleteWorkspaceByID(id string) error
	LockWorkspace(id string, opts WorkspaceLockOptions) (*tfe.Workspace, error)
	UnlockWorkspace(id string) (*tfe.Workspace, error)
}

func NewWorkspaceID() string {
	return fmt.Sprintf("ws-%s", GenerateRandomString(16))
}

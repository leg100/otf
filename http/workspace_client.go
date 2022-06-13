package http

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/dto"
)

// Compile-time proof of interface implementation.
var _ otf.WorkspaceService = (*workspaces)(nil)

// workspaces implements WorkspaceService.
type workspaces struct {
	client *client

	// TODO: implement all of otf.WorkspaceService's methods
	otf.WorkspaceService
}

// Create is used to create a new workspace.
func (s *workspaces) Create(ctx context.Context, options otf.WorkspaceCreateOptions) (*otf.Workspace, error) {
	if err := options.Valid(); err != nil {
		return nil, err
	}

	u := fmt.Sprintf("organizations/%s/workspaces", url.QueryEscape(options.OrganizationName))
	req, err := s.client.newRequest("POST", u, &options)
	if err != nil {
		return nil, err
	}

	w := &dto.Workspace{}
	err = s.client.do(ctx, req, w)
	if err != nil {
		return nil, err
	}

	return otf.UnmarshalWorkspaceJSONAPI(w), nil
}

// Retrieve a workspace either by its ID, or by organization and workspace name.
func (s *workspaces) Get(ctx context.Context, spec otf.WorkspaceSpec) (*otf.Workspace, error) {
	path, err := getWorkspacePath(spec)
	if err != nil {
		return nil, err
	}

	req, err := s.client.newRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	w := &dto.Workspace{}
	err = s.client.do(ctx, req, w)
	if err != nil {
		return nil, err
	}

	// durations come over in ms
	w.ApplyDurationAverage *= time.Millisecond
	w.PlanDurationAverage *= time.Millisecond

	return otf.UnmarshalWorkspaceJSONAPI(w), nil
}

// List all the workspaces within an organization.
func (s *workspaces) List(ctx context.Context, options otf.WorkspaceListOptions) (*otf.WorkspaceList, error) {
	u := fmt.Sprintf("organizations/%s/workspaces", url.QueryEscape(*options.OrganizationName))
	req, err := s.client.newRequest("GET", u, &options)
	if err != nil {
		return nil, err
	}

	wl := &dto.WorkspaceList{}
	err = s.client.do(ctx, req, wl)
	if err != nil {
		return nil, err
	}

	return otf.UnmarshalWorkspaceListJSONAPI(wl), nil
}

// Update settings of an existing workspace.
func (s *workspaces) Update(ctx context.Context, spec otf.WorkspaceSpec, options otf.WorkspaceUpdateOptions) (*otf.Workspace, error) {
	path, err := getWorkspacePath(spec)
	if err != nil {
		return nil, err
	}

	req, err := s.client.newRequest("PATCH", path, &options)
	if err != nil {
		return nil, err
	}

	w := &dto.Workspace{}
	err = s.client.do(ctx, req, w)
	if err != nil {
		return nil, err
	}

	return otf.UnmarshalWorkspaceJSONAPI(w), nil
}

// Delete a workspace by its name.
func (s *workspaces) Delete(ctx context.Context, spec otf.WorkspaceSpec) error {
	path, err := getWorkspacePath(spec)
	if err != nil {
		return err
	}

	req, err := s.client.newRequest("DELETE", path, nil)
	if err != nil {
		return err
	}

	return s.client.do(ctx, req, nil)
}

// Lock a workspace by its ID.
func (s *workspaces) Lock(ctx context.Context, spec otf.WorkspaceSpec, opts otf.WorkspaceLockOptions) (*otf.Workspace, error) {
	var path string
	if spec.ID != nil {
		path = fmt.Sprintf("workspaces/%s/actions/lock", url.QueryEscape(*spec.ID))
	} else if spec.OrganizationName != nil && spec.Name != nil {
		path = fmt.Sprintf("organizations/%s/workspaces/%s/actions/lock", url.QueryEscape(*spec.OrganizationName), url.QueryEscape(*spec.Name))
	} else {
		return nil, otf.ErrInvalidWorkspaceSpec
	}
	req, err := s.client.newRequest("POST", path, &opts)
	if err != nil {
		return nil, err
	}

	w := &dto.Workspace{}
	err = s.client.do(ctx, req, w)
	if err != nil {
		return nil, err
	}

	return otf.UnmarshalWorkspaceJSONAPI(w), nil
}

// Unlock a workspace by its ID.
func (s *workspaces) Unlock(ctx context.Context, spec otf.WorkspaceSpec, _ otf.WorkspaceUnlockOptions) (*otf.Workspace, error) {
	var path string
	if spec.ID != nil {
		path = fmt.Sprintf("workspaces/%s/actions/unlock", url.QueryEscape(*spec.ID))
	} else if spec.OrganizationName != nil && spec.Name != nil {
		path = fmt.Sprintf("organizations/%s/workspaces/%s/actions/unlock", url.QueryEscape(*spec.OrganizationName), url.QueryEscape(*spec.Name))
	} else {
		return nil, otf.ErrInvalidWorkspaceSpec
	}
	req, err := s.client.newRequest("POST", path, nil)
	if err != nil {
		return nil, err
	}

	w := &dto.Workspace{}
	err = s.client.do(ctx, req, w)
	if err != nil {
		return nil, err
	}

	return otf.UnmarshalWorkspaceJSONAPI(w), nil
}

// getWorkspacePath generates a URL path for a workspace according to whether
// the spec specifies an ID, or an organization and workspace name.
func getWorkspacePath(spec otf.WorkspaceSpec) (string, error) {
	if spec.ID != nil {
		return fmt.Sprintf("workspaces/%s", url.QueryEscape(*spec.ID)), nil
	}

	if spec.Name != nil && spec.OrganizationName != nil {
		return fmt.Sprintf(
			"organizations/%s/workspaces/%s",
			url.QueryEscape(*spec.OrganizationName),
			url.QueryEscape(*spec.Name),
		), nil
	}

	return "", fmt.Errorf("invalid workspace spec")
}

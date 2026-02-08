package ui

import (
	"context"

	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/user"
	"github.com/leg100/otf/internal/workspace"
)

// workspaceCache is a caching client for retrieving workspaces
type workspaceCache struct {
	cache  map[resource.TfeID]*workspace.Workspace
	getter workspaceCacheService
}

type workspaceCacheService interface {
	Get(ctx context.Context, workspaceID resource.TfeID) (*workspace.Workspace, error)
}

func newWorkspaceCache(getter workspaceCacheService) *workspaceCache {
	return &workspaceCache{
		cache:  make(map[resource.TfeID]*workspace.Workspace),
		getter: getter,
	}
}

func (c *workspaceCache) Get(ctx context.Context, workspaceID resource.TfeID) (*workspace.Workspace, error) {
	if ws, ok := c.cache[workspaceID]; ok {
		return ws, nil
	}
	ws, err := c.getter.Get(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	c.cache[workspaceID] = ws
	return ws, nil
}

// userCache is a caching client for retrieving users
type userCache struct {
	cache  map[user.Username]*user.User
	getter userCacheService
}

type userCacheService interface {
	GetUser(ctx context.Context, spec user.UserSpec) (*user.User, error)
}

func newUserCache(getter userCacheService) *userCache {
	return &userCache{
		cache:  make(map[user.Username]*user.User),
		getter: getter,
	}
}

func (c *userCache) GetUser(ctx context.Context, spec user.UserSpec) (*user.User, error) {
	if spec.Username != nil {
		if user, ok := c.cache[*spec.Username]; ok {
			return user, nil
		}
	}
	user, err := c.getter.GetUser(ctx, spec)
	if err != nil {
		return nil, err
	}
	c.cache[user.Username] = user
	return user, nil
}

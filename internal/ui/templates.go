package ui

import (
	"github.com/leg100/otf/internal/authz"
)

type templates struct {
	*authz.Authorizer

	workspaces WorkspaceService
	users      UserService
	configs    ConfigVersionService
}

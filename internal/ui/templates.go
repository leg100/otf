package ui

import (
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/user"
	"github.com/leg100/otf/internal/workspace"
)

type templates struct {
	workspaces *workspace.Service
	users      *user.Service
	configs    *configversion.Service
}

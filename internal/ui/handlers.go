package ui

import (
	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/user"
	"github.com/leg100/otf/internal/workspace"
)

// Handlers implements internal.Handlers and registers all UI handlers
type Handlers struct {
	Logger     logr.Logger
	Runs       *run.Service
	Workspaces *workspace.Service
	Users      *user.Service
}

// AddHandlers registers all UI handlers with the router
func (h *Handlers) AddHandlers(r *mux.Router) {
	AddRunHandlers(r, h.Logger, h.Runs, h.Workspaces, h.Users, h.Runs)
}

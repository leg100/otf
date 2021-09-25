package http

import (
	"github.com/leg100/go-tfe"
)

const (
	DefaultAddress = "localhost:8080"
)

type Client interface {
	Organizations() tfe.Organizations
	Workspaces() tfe.Workspaces
}

type client struct {
	*tfe.Client
}

func (c *client) Organizations() tfe.Organizations {
	return c.Client.Organizations
}

func (c *client) Workspaces() tfe.Workspaces {
	return c.Client.Workspaces
}

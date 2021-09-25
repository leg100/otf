package main

import (
	"github.com/leg100/go-tfe"
)

const (
	DefaultAddress = "localhost:8080"
)

type ClientConfig interface {
	NewClient() (Client, error)
}

type clientConfig struct {
	tfe.Config
}

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

func (c *clientConfig) NewClient() (Client, error) {
	var err error

	c.Address, err = sanitizeAddress(c.Address)
	if err != nil {
		return nil, err
	}

	creds, err := NewCredentialsStore()
	if err != nil {
		return nil, err
	}

	// If --token isn't set then load from DB
	if c.Token == "" {
		c.Token, err = creds.Load(c.Address)
		if err != nil {
			return nil, err
		}
	}

	tfeClient, err := tfe.NewClient(&c.Config)
	if err != nil {
		return nil, err
	}

	return &client{Client: tfeClient}, nil
}

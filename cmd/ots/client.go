package main

import (
	"fmt"
	"net/url"

	"github.com/hashicorp/go-tfe"
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
	if err := c.sanitizeAddress(); err != nil {
		return nil, err
	}

	creds, err := NewCredentialsStore(&SystemDirectories{})
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

// Ensure address is in format https://<host>:<port>
func (c *clientConfig) sanitizeAddress() error {
	u, err := url.ParseRequestURI(c.Address)
	if err != nil || u.Host == "" {
		u, er := url.ParseRequestURI("https://" + c.Address)
		if er != nil {
			return fmt.Errorf("could not parse hostname: %w", err)
		}
		c.Address = u.String()
		return nil
	}

	u.Scheme = "https"
	c.Address = u.String()

	return nil
}

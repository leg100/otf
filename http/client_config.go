package http

import "github.com/leg100/go-tfe"

var _ ClientFactory = (*ClientConfig)(nil)

// ClientConfig is an implementation of a ClientFactory, using the TFE
// configuration and credentials to construct new clients.
type ClientConfig struct {
	tfe.Config
}

func (c *ClientConfig) NewClient() (Client, error) {
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

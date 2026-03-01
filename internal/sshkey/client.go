package sshkey

import (
	"context"
	"fmt"

	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/resource"
)

// Client is an HTTP client for the SSH key API, used by agent runners.
type Client struct {
	*otfhttp.Client
}

func (c *Client) GetSSHKey(ctx context.Context, id resource.TfeID) (*SSHKey, error) {
	path := fmt.Sprintf("ssh-keys/%s", id)
	req, err := c.NewRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	var key TFESSHKey
	if err := c.Do(ctx, req, &key); err != nil {
		return nil, err
	}
	return &SSHKey{
		ID:         key.ID,
		Name:       key.Name,
		PrivateKey: key.Value,
	}, nil
}

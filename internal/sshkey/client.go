package sshkey

import (
	"bytes"
	"context"
	"fmt"

	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/resource"
)

// Alias client to permit embedding it with other clients in a struct
// without a name clash.
type SSHKeyClient = Client

// Client is an HTTP client for the SSH key API, used by agent runners.
type Client struct {
	*otfhttp.Client
}

func (c *Client) GetSSHKeyPrivateKey(ctx context.Context, id resource.TfeID) ([]byte, error) {
	path := fmt.Sprintf("ssh-keys/%s", id)
	req, err := c.NewRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	buf := bytes.Buffer{}
	if err := c.Do(ctx, req, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

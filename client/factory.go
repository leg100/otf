package client

import "github.com/leg100/otf/http"

type ClientFactory interface {
	NewClient(http.Config) (*client, error)
}

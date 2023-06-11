package http

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/leg100/otf/internal/apigen"
)

type Client struct {
	*apigen.Client

	baseURL *url.URL
}

func NewClient(config Config) (*Client, error) {
	addr, err := SanitizeAddress(config.Address)
	if err != nil {
		return nil, err
	}
	baseURL, err := url.Parse(addr)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %v", err)
	}
	baseURL.Path = config.BasePath
	if !strings.HasSuffix(baseURL.Path, "/") {
		baseURL.Path += "/"
	}
	// This value must be provided by the user.
	if config.Token == "" {
		return nil, fmt.Errorf("missing API token")
	}
	// Configure transports
	pt := cleanhttp.DefaultPooledTransport()
	pt.TLSClientConfig = &tls.Config{InsecureSkipVerify: config.Insecure}
	bt := HeadersTransport{rt: pt, token: config.Token}
	httpClient := &http.Client{Transport: &bt}

	// Construct the (auto-generated) client.
	client, err := apigen.NewClient(baseURL.String(), apigen.WithClient(httpClient))
	if err != nil {
		return nil, err
	}
	return &Client{Client: client, baseURL: baseURL}, nil
}

// Hostname returns the server host:port.
func (c *Client) Hostname() string {
	return c.baseURL.Host
}

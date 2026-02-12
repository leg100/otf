// Package github provides github related code
package github

import (
	"net/url"

	"github.com/leg100/otf/internal"
)

func DefaultBaseURL() *internal.WebURL {
	return &internal.WebURL{
		URL: url.URL{
			Scheme: "https",
			Host:   "github.com",
		},
	}
}

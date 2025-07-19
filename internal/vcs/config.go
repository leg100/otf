package vcs

import "net/url"

type Config struct {
	Token        *string
	Installation *Installation
	// The base URL of the API of the provider. Optional.
	APIURL *url.URL
}

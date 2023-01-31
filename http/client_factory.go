package http

import "github.com/leg100/otf"

// ClientFactory implementations of capable of constructing new OTF clients
type ClientFactory interface {
	NewClient() (otf.Client, error)
}

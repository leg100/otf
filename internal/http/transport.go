package http

import (
	"crypto/tls"
	"net/http"
)

var DefaultTransport http.RoundTripper = http.DefaultTransport
var InsecureTransport http.RoundTripper

func init() {
	// Assign InsecureTransport package variable.
	clone := http.DefaultTransport.(*http.Transport).Clone()
	clone.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
	InsecureTransport = clone
}

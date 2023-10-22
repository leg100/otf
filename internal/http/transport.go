package http

import (
	"crypto/tls"
	"net/http"
)

var InsecureTransport http.RoundTripper

func init() {
	// Assign InsecureTransport package variable.
	clone := http.DefaultTransport.(*http.Transport).Clone()
	clone.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
	InsecureTransport = clone
}

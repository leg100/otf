package http

import (
	"crypto/tls"
	"net/http"
)

var insecureTransport http.RoundTripper

func init() {
	// http.DefaultTransport is a pkg variable, so we need to clone it to
	// avoid disabling TLS verification globally.
	clone := http.DefaultTransport.(*http.Transport).Clone()
	clone.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
	insecureTransport = clone
}

// DefaultTransport wraps the stdlib http.DefaultTransport, returning either
// http.DefaultTransport, or if skipTLSVerification is true, a clone of
// http.DefaultTransport configured to skip TLS verification.
func DefaultTransport(skipTLSVerification bool) http.RoundTripper {
	if skipTLSVerification {
		return insecureTransport
	}
	return http.DefaultTransport
}

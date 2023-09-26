package http

import (
	"crypto/tls"
	"net/http"
)

// DefaultTransport wraps the stdlib http.DefaultTransport, optionally disabling
// TLS verification.
func DefaultTransport(skipTLSVerification bool) http.RoundTripper {
	if skipTLSVerification {
		// http.DefaultTransport is a pkg variable, so we need to clone it to
		// avoid disabling TLS verification globally.
		cloned := http.DefaultTransport.(*http.Transport).Clone()
		cloned.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
		return cloned
	}
	return http.DefaultTransport
}

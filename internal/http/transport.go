package http

import (
	"crypto/tls"
	"net/http"
)

func DefaultTransport(skipTLSVerification bool) http.RoundTripper {
	dt := http.DefaultTransport
	if skipTLSVerification {
		dt.(*http.Transport).TLSClientConfig = &tls.Config{
			InsecureSkipVerify: skipTLSVerification,
		}
	}
	return dt
}

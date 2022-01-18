/*
Package util shouldn't exist, but allows us to escape an import cycle for
* routines shared by both http and http/html packages.
*/

package util

import (
	"net/http"
	"net/url"
)

var (
	// Whether or not oTF is running with SSL enabled.
	//
	// TODO: replace with something that doesn't involve package variables!
	SSL bool
)

// Absolute returns an absolute URL for the given path. It uses the http request
// to determine the correct hostname and scheme to use. Handles situations where
// oTF is sitting behind a reverse proxy, using the X-Forwarded-* headers the
// proxy sets.
func Absolute(r *http.Request, path string) string {
	u := url.URL{
		Host: r.Host,
		Path: path,
	}

	if SSL {
		u.Scheme = "https"
	} else {
		u.Scheme = "http"
	}

	if host := r.Header.Get("X-Forwarded-Host"); host != "" {
		u.Host = host
	}
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		u.Scheme = proto
	}

	return u.String()
}

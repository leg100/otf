/*
Package http provides an HTTP interface allowing HTTP clients to interact with OTF.
*/
package http

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gorilla/schema"
	"github.com/leg100/jsonapi"
)

// Query schema Encoder, caches structs, and safe for sharing
var encoder = schema.NewEncoder()

// Absolute returns an absolute URL for the given path. It uses the http request
// to determine the correct hostname and scheme to use. Handles situations where
// otf is sitting behind a reverse proxy, using the X-Forwarded-* headers the
// proxy sets.
func Absolute(r *http.Request, path string) string {
	u := url.URL{
		Host: ExternalHost(r),
		Path: path,
	}

	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		u.Scheme = proto
	} else if r.TLS != nil {
		u.Scheme = "https"
	} else {
		u.Scheme = "http"
	}

	return u.String()
}

// ExternalHost uses the incoming HTTP request to determine the host:port on
// which this server can be reached externally by clients and the internet.
func ExternalHost(r *http.Request) string {
	if host := r.Header.Get("X-Forwarded-Host"); host != "" {
		return host
	}
	return r.Host
}

// SanitizeHostname ensures hostname is in the format <host>:<port>
func SanitizeHostname(hostname string) (string, error) {
	u, err := url.ParseRequestURI(hostname)
	if err != nil || u.Host == "" {
		u, er := url.ParseRequestURI("https://" + hostname)
		if er != nil {
			return "", fmt.Errorf("could not parse hostname: %w", err)
		}
		return u.Host, nil
	}
	return u.Host, nil
}

// SanitizeAddress ensures address is in format https://<host>:<port>
func SanitizeAddress(address string) (string, error) {
	u, err := url.ParseRequestURI(address)
	if err != nil || u.Host == "" {
		u, er := url.ParseRequestURI("https://" + address)
		if er != nil {
			return "", fmt.Errorf("could not parse hostname: %w", err)
		}
		return u.String(), nil
	}
	u.Scheme = "https"
	return u.String(), nil
}

// GetClientIP gets the client's IP address
func GetClientIP(r *http.Request) (string, error) {
	// reverse proxy adds client IP to an HTTP header, and each successive proxy
	// adds a client IP, so we want the leftmost IP.
	if hdr := r.Header.Get("X-Forwarded-For"); hdr != "" {
		first, _, _ := strings.Cut(hdr, ",")
		addr := strings.TrimSpace(first)
		if isIP := net.ParseIP(addr); isIP != nil {
			return addr, nil
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	return host, err
}

// writeError writes an HTTP response with a JSON-API marshalled error obj.
func writeError(w http.ResponseWriter, code int, err error) {
	w.Header().Set("Content-type", jsonapi.MediaType)
	w.WriteHeader(code)
	jsonapi.MarshalErrors(w, []*jsonapi.ErrorObject{
		{
			Status: strconv.Itoa(code),
			Title:  http.StatusText(code),
			Detail: err.Error(),
		},
	})
}

// withCode is a helper func for writing an HTTP status code to a response
// stream.  For use with WriteResponse.
func withCode(code int) func(w http.ResponseWriter) {
	return func(w http.ResponseWriter) {
		w.WriteHeader(code)
	}
}

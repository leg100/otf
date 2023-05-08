package loginserver

import (
	"net/http"
	"net/url"
)

// https://datatracker.ietf.org/doc/html/rfc6749#section-4.1.2.1
const (
	ErrInvalidRequest          string = "invalid_request"
	ErrInvalidGrant            string = "invalid_grant"
	ErrInvalidClient           string = "invalid_client"
	ErrUnsupportedGrantType    string = "unsupported_grant_type"
	ErrUnsupportedResponseType string = "unsupported_response_type"
	ErrAccessDenied            string = "access_denied"
	ErrServerError             string = "server_error"
)

type redirectError struct {
	redirect *url.URL
	state    string
}

func (e redirectError) error(w http.ResponseWriter, r *http.Request, error, description string) {
	q := e.redirect.Query()
	q.Add("error", error)
	q.Add("error_description", error)
	if e.state != "" {
		q.Add("state", e.state)
	}
	e.redirect.RawQuery = q.Encode()

	http.Redirect(w, r, e.redirect.String(), http.StatusFound)
}

package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gorilla/schema"
	"github.com/leg100/jsonapi"
	"github.com/leg100/otf"
)

var (
	// Query schema encoder, caches structs, and safe for sharing
	encoder = schema.NewEncoder()

	// Query schema decoder: caches structs, and safe for sharing.
	decoder = schema.NewDecoder()
)

// DecodeQuery unmarshals a query string (k1=v1&k2=v2...) into a struct.
func DecodeQuery(opts interface{}, query url.Values) error {
	if err := decoder.Decode(opts, query); err != nil {
		return fmt.Errorf("unable to decode query string: %w", err)
	}
	return nil
}

// WithCode is a helper func for writing an HTTP status code to a response
// stream.  For use with WriteResponse.
func WithCode(code int) func(w http.ResponseWriter) {
	return func(w http.ResponseWriter) {
		w.WriteHeader(code)
	}
}

// WriteResponse writes an HTTP response with a JSON-API marshalled payload.
func WriteResponse(w http.ResponseWriter, r *http.Request, obj interface{}, opts ...func(http.ResponseWriter)) {
	w.Header().Set("Content-type", jsonapi.MediaType)

	for _, o := range opts {
		o(w)
	}

	// Only sideline relationships for responses to GET requests
	var err error
	if r.Method == "GET" {
		err = MarshalPayload(w, r, obj)
	} else {
		err = jsonapi.MarshalPayloadWithoutIncluded(w, obj)
	}

	if err != nil {
		WriteError(w, http.StatusInternalServerError, err)
	}
}

// WriteError writes an HTTP response with a JSON-API marshalled error obj.
func WriteError(w http.ResponseWriter, code int, err error) {
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

// checkResponseCode can be used to check the status code of an HTTP request.
func checkResponseCode(r *http.Response) error {
	if r.StatusCode >= 200 && r.StatusCode <= 299 {
		return nil
	}

	switch r.StatusCode {
	case 401:
		return otf.ErrUnauthorized
	case 404:
		return otf.ErrResourceNotFound
	case 409:
		switch {
		case strings.HasSuffix(r.Request.URL.Path, "actions/lock"):
			return otf.ErrWorkspaceLocked
		case strings.HasSuffix(r.Request.URL.Path, "actions/unlock"):
			return otf.ErrWorkspaceNotLocked
		case strings.HasSuffix(r.Request.URL.Path, "actions/force-unlock"):
			return otf.ErrWorkspaceNotLocked
		}
	}

	// Decode the error payload.
	errPayload := &jsonapi.ErrorsPayload{}
	err := json.NewDecoder(r.Body).Decode(errPayload)
	if err != nil || len(errPayload.Errors) == 0 {
		return fmt.Errorf(r.Status)
	}

	// Parse and format the errors.
	var errs []string
	for _, e := range errPayload.Errors {
		if e.Detail == "" {
			errs = append(errs, e.Title)
		} else {
			errs = append(errs, fmt.Sprintf("%s\n\n%s", e.Title, e.Detail))
		}
	}

	return fmt.Errorf(strings.Join(errs, "\n"))
}

func parsePagination(body io.Reader) (*otf.Pagination, error) {
	var raw struct {
		Meta struct {
			Pagination otf.Pagination `jsonapi:"pagination"`
		} `jsonapi:"meta"`
	}

	// JSON decode the raw response.
	if err := json.NewDecoder(body).Decode(&raw); err != nil {
		return &otf.Pagination{}, err
	}

	return &raw.Meta.Pagination, nil
}

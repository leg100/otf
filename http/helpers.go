package http

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/schema"
	"github.com/leg100/jsonapi"
)

// Query schema decoder: caches structs, and safe for sharing.
var decoder = schema.NewDecoder()

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

// Ensure hostname is in the format <host>:<port>
func sanitizeHostname(hostname string) (string, error) {
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

// Ensure address is in format https://<host>:<port>
func sanitizeAddress(address string) (string, error) {
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

// Package decode contains decoders for various HTTP artefacts
package decode

import (
	"errors"
	"maps"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/resource"
)

// Query schema decoder: caches structs, and safe for sharing.
var decoder *schema.Decoder

func init() {
	decoder = schema.NewDecoder()
	// Don't error if there are keys in the source map that are not present in
	// the destination struct.
	decoder.IgnoreUnknownKeys(true)
	// Don't skip decoding empty strings from source map to the destination
	// struct.
	decoder.ZeroEmpty(true)
}

// Form decodes an HTTP request's POST form contents into dst.
func Form(dst any, r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return err
	}
	if err := Decode(dst, r.PostForm); err != nil {
		return err
	}
	return nil
}

// Query unmarshals a query string (k1=v1&k2=v2...) into dst.
func Query(dst any, query url.Values) error {
	if err := Decode(dst, query); err != nil {
		return err
	}
	return nil
}

// Route decodes a mux route parameters (e.g. /foo/{bar}) into dst.
func Route(dst any, r *http.Request) error {
	// decoder only takes map[string][]string, not map[string]string
	vars := convertStrMapToStrSliceMap(mux.Vars(r))
	if err := Decode(dst, vars); err != nil {
		return err
	}
	return nil
}

// All populates the struct pointed to by dst with query params, req body
// params, cookie values, and request path variables, with the following
// precedence:
// 1. cookies
// 2. path variables
// 3. body params
// 4. query params
func All(dst any, r *http.Request) error {
	// Parses both query and req body if POST/PUT/PATCH
	if err := r.ParseForm(); err != nil {
		return err
	}
	vars := make(map[string][]string, len(r.Form))
	maps.Copy(vars, r.Form)
	// Merge in request path variables
	for k, v := range mux.Vars(r) {
		vars[k] = []string{v}
	}
	// Merge in cookie values
	for _, cookie := range r.Cookies() {
		vars[cookie.Name] = []string{cookie.Value}
	}
	if err := Decode(dst, vars); err != nil {
		return err
	}
	return nil
}

// Param retrieves a single parameter by name from the request, first checking the body
// (if POST/PUT/PATCH) and the query, falling back to looking for a path variable.
func Param(name string, r *http.Request) (string, error) {
	// Parses both query and req body
	if err := r.ParseForm(); err != nil {
		return "", err
	}
	if v := r.Form.Get(name); v != "" {
		return v, nil
	}
	if v, ok := mux.Vars(r)[name]; ok {
		return v, nil
	}
	return "", &internal.ErrMissingParameter{Parameter: name}
}

// ID retrieves a single parameter by name from the request and parses into a
// resource ID.
func ID(name string, r *http.Request) (resource.TfeID, error) {
	s, err := Param(name, r)
	if err != nil {
		return resource.TfeID{}, err
	}
	return resource.ParseTfeID(s)
}

func Decode(dst any, src map[string][]string) error {
	if err := decoder.Decode(dst, src); err != nil {
		var emptyField schema.EmptyFieldError
		if errors.As(err, &emptyField) {
			return &internal.ErrMissingParameter{Parameter: emptyField.Key}
		}
		return err
	}
	return nil
}

func convertStrMapToStrSliceMap(m map[string]string) map[string][]string {
	mm := make(map[string][]string, len(m))
	for k, v := range m {
		mm[k] = []string{v}
	}
	return mm
}

package html

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
)

var (
	// Query schema decoder: caches structs, and safe for sharing.
	decoder = schema.NewDecoder()
)

// decode collectively decodes route params, query params and form params into
// obj (a pointer to a struct)
func decode(r *http.Request, obj interface{}) error {
	if err := decodeForm(r, obj); err != nil {
		return err
	}

	if err := decodeQuery(r, obj); err != nil {
		return err
	}

	if err := decodeRouteVars(r, obj); err != nil {
		return err
	}

	return nil
}

func decodeForm(r *http.Request, obj interface{}) error {
	if err := r.ParseForm(); err != nil {
		return err
	}

	if err := decoder.Decode(obj, r.PostForm); err != nil {
		return err
	}

	return nil
}

func decodeQuery(r *http.Request, obj interface{}) error {
	if err := decoder.Decode(obj, r.URL.Query()); err != nil {
		return err
	}

	return nil
}

func decodeRouteVars(r *http.Request, obj interface{}) error {
	// decoder only takes map[string][]string, not map[string]string
	vars := convertStrMapToStrSliceMap(mux.Vars(r))

	if err := decoder.Decode(obj, vars); err != nil {
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

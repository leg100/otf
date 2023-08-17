package tfe

import (
	"context"
	"net/http"
	"reflect"
	"strings"
	"sync"
)

const (
	IncludeOrganization IncludeName = "organization"
	IncludeWorkspace    IncludeName = "workspace"
	IncludeCurrentRun   IncludeName = "current-run"
	IncludeConfig       IncludeName = "configuration_version"
	IncludeIngress      IncludeName = "ingress_attributes"
	IncludeUsers        IncludeName = "users"
	IncludeCreatedBy    IncludeName = "created-by"
	IncludeOutputs      IncludeName = "outputs"
)

type (
	// includer includes related resources in API responses, as documented here:
	//
	// https://developer.hashicorp.com/terraform/cloud-docs/api-docs#inclusion-of-related-resources
	includer struct {
		registrations map[IncludeName]IncludeFunc
		mu            sync.Mutex
	}

	// IncludeName is the name used in the query parameter to request a resource
	// be included, i.e. /?include=<IncludeName>
	IncludeName string

	// IncludeFunc retrieves the resource for inclusion
	IncludeFunc func(context.Context, any) (any, error)
)

// Register registers an IncludeFunc to be called whenever IncludeName is
// requested in an API query.
func (i *includer) Register(name IncludeName, f IncludeFunc) {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.registrations[name] = f
}

// addIncludes includes related resources if the API request requests them via
// the ?include= query parameter. Multiple include values are comma separated, and
// each include names the resource type to be included. Optionally, the include
// may include a period ('.') as a separator, which specifies that the further
// transitive resources be included too, e.g. ?include=a.b means fetch resource
// of type a, and then fetch resource of type b that is a relation of type a.
func (i *includer) addIncludes(r *http.Request, v any) ([]any, error) {
	// only include resources in response to GET requests.
	if r.Method != "GET" {
		return nil, nil
	}
	var includes []any
	if q := r.URL.Query().Get("include"); q != "" {
		for _, relation := range strings.Split(q, ",") {
			parent := v
			for _, resource := range strings.Split(relation, ".") {
				f, ok := i.registrations[IncludeName(resource)]
				if !ok {
					continue
				}
				add := func(v any) error {
					inc, err := f(r.Context(), v)
					if err != nil {
						return err
					}
					includes = append(includes, inc)
					return nil
				}
				// handle when v is a slice
				if dst := reflect.ValueOf(parent); dst.Kind() == reflect.Slice {
					for i := 0; i < dst.Len(); i++ {
						if err := add(dst.Index(i).Interface()); err != nil {
							return nil, err
						}
					}
				}
				if err := add(parent); err != nil {
					return nil, err
				}
				parent = resource
			}
		}
	}
	return includes, nil
}

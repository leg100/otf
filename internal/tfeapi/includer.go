package tfeapi

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
	IncludeFunc func(context.Context, any) ([]any, error)
)

// Register registers an IncludeFunc to be called whenever IncludeName is
// requested in an API query.
func (i *includer) Register(name IncludeName, f IncludeFunc) {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.registrations[name] = f
}

// addIncludes handles API queries of the form ?include=v,..., which is a comma
// separated list of related resource types to include. For example, the query:
//
// /foo?include=a,b
//
// requests the 'foo' resource, but also requests that the related resource 'a' be
// included, as well as 'b'.
//
// an included resource may be period ('.') separated, in which case its
// relations are included too. For example:
//
// /foo?include=a.b
//
// results in not only the resource 'a' being included, but also resource 'b'
// where 'b' is  relation of 'a'.
//
// each resource may return multiple items. For example:
//
// /foo?include=a.b
//
// /foo may return multiple resources of type 'foo', in which case the resource
// 'a' is included for each 'foo' resource, and the resource 'b' is included for
// each 'a' resource.
func (i *includer) addIncludes(r *http.Request, v any) ([]any, error) {
	// only include resources in response to GET requests.
	if r.Method != "GET" {
		return nil, nil
	}
	fetchChildren := func(f IncludeFunc, v any) ([]any, error) {
		// handle when v is a slice
		if dst := reflect.ValueOf(v); dst.Kind() == reflect.Slice {
			var children []any
			for i := 0; i < dst.Len(); i++ {
				results, err := f(r.Context(), dst.Index(i).Interface())
				if err != nil {
					return nil, err
				}
				children = append(children, results...)
			}
			return children, nil
		}
		return f(r.Context(), v)
	}
	var includes []any
	if q := r.URL.Query().Get("include"); q != "" {
		for _, relation := range strings.Split(q, ",") {
			parents := []any{v}
			for _, resource := range strings.Split(relation, ".") {
				f, ok := i.registrations[IncludeName(resource)]
				if !ok {
					continue
				}
				var children []any
				for _, p := range parents {
					c, err := fetchChildren(f, p)
					if err != nil {
						return nil, err
					}
					children = append(children, c...)
				}
				includes = append(includes, children...)
				// children become parents
				parents = children
			}
		}
	}
	return includes, nil
}

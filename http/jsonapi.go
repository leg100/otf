package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/google/jsonapi"
)

// Marshal a JSON-API response.
func MarshalPayload(w io.Writer, r *http.Request, models interface{}) error {
	include := strings.Split(r.URL.Query().Get("include"), ",")

	// Get the value of models so we can test if it's a struct.
	dst := reflect.Indirect(reflect.ValueOf(models))

	// Return an error if model is not a struct or an io.Writer.
	if dst.Kind() != reflect.Struct {
		return fmt.Errorf("v must be a struct or an io.Writer")
	}

	// Try to get the Items and Pagination struct fields.
	items := dst.FieldByName("Items")
	pagination := dst.FieldByName("Pagination")

	// Marshal a single value if v does not contain the Items and Pagination
	// struct fields.
	if !items.IsValid() || !pagination.IsValid() {
		return marshalSinglePayload(w, models, include...)
	}

	// Return an error if v.Items is not a slice.
	if items.Type().Kind() != reflect.Slice {
		return fmt.Errorf("v.Items must be a slice")
	}

	payload, err := jsonapi.Marshal(items.Interface())
	if err != nil {
		return fmt.Errorf("unable to marshal payload: %w", err)
	}

	manyPayload, ok := payload.(*jsonapi.ManyPayload)
	if !ok {
		return fmt.Errorf("unable to extract concrete value")
	}

	manyPayload.Included = filterIncluded(manyPayload.Included, include...)

	manyPayload.Meta = &jsonapi.Meta{"pagination": pagination.Interface()}

	return json.NewEncoder(w).Encode(payload)
}

func marshalSinglePayload(w io.Writer, model interface{}, include ...string) error {
	payload, err := jsonapi.Marshal(model)
	if err != nil {
		return fmt.Errorf("unable to marshal payload: %w", err)
	}

	onePayload, ok := payload.(*jsonapi.OnePayload)
	if !ok {
		return fmt.Errorf("unable to extract concrete value")
	}

	onePayload.Included = filterIncluded(onePayload.Included, include...)

	return json.NewEncoder(w).Encode(onePayload)
}

func filterIncluded(included []*jsonapi.Node, filters ...string) (filtered []*jsonapi.Node) {
	for _, f := range filters {
		// Filter should always be pural but sometimes the client (go-tfe)
		// neglects to do this
		if !strings.HasSuffix(f, "s") {
			f = f + "s"
		}

		for _, inc := range included {
			if inc.Type == f {
				filtered = append(filtered, inc)
			}
		}
	}
	return filtered
}

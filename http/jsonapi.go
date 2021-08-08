package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/leg100/jsonapi"
)

// MarshalPayload marshals the models object into a JSON-API response.
func MarshalPayload(w io.Writer, r *http.Request, models interface{}) error {
	include := strings.Split(r.URL.Query().Get("include"), ",")
	include = sanitizeIncludes(include)

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

	payload, err := jsonapi.Marshal(items.Interface(), include)
	if err != nil {
		return fmt.Errorf("unable to marshal payload: %w", err)
	}

	manyPayload, ok := payload.(*jsonapi.ManyPayload)
	if !ok {
		return fmt.Errorf("unable to extract concrete value")
	}

	manyPayload.Meta = &jsonapi.Meta{"pagination": pagination.Interface()}

	return json.NewEncoder(w).Encode(payload)
}

func marshalSinglePayload(w io.Writer, model interface{}, include ...string) error {
	if err := jsonapi.MarshalPayload(w, model, include); err != nil {
		return fmt.Errorf("unable to marshal payload: %w", err)
	}
	return nil
}

func sanitizeIncludes(includes []string) (sanitized []string) {
	for _, i := range includes {
		sanitized = append(sanitized, strings.ReplaceAll(i, "_", "-"))
	}
	return
}

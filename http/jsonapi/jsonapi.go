// Package jsonapi handles marshaling/unmarshaling into/from json-api
package jsonapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/leg100/jsonapi"
)

type ErrorsPayload jsonapi.ErrorsPayload

// Pagination is used to return the pagination details of an API request.
type Pagination struct {
	CurrentPage  int  `json:"current-page"`
	PreviousPage *int `json:"prev-page"`
	NextPage     *int `json:"next-page"`
	TotalPages   int  `json:"total-pages"`
	TotalCount   int  `json:"total-count"`
}

func UnmarshalPayload(in io.Reader, model interface{}) error {
	return jsonapi.UnmarshalPayload(in, model)
}

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

func MarshalPayloadWithoutIncluded(w io.Writer, model interface{}) error {
	return jsonapi.MarshalPayloadWithoutIncluded(w, model)
}

func UnmarshalManyPayload(in io.Reader, t reflect.Type) ([]interface{}, error) {
	return jsonapi.UnmarshalManyPayload(in, t)
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

// WriteResponse writes an HTTP response with a JSON-API marshalled payload.
func WriteResponse(w http.ResponseWriter, r *http.Request, obj Assembler, opts ...func(http.ResponseWriter)) {
	w.Header().Set("Content-type", jsonapi.MediaType)
	for _, o := range opts {
		o(w)
	}
	// Only sideline relationships for responses to GET requests
	var err error
	if r.Method == "GET" {
		err = MarshalPayload(w, r, obj.ToJSONAPI())
	} else {
		err = jsonapi.MarshalPayloadWithoutIncluded(w, obj.ToJSONAPI())
	}
	if err != nil {
		Error(w, http.StatusInternalServerError, err)
	}
}

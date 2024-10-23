package tfeapi

import (
	"errors"
	"net/http"

	"github.com/DataDog/jsonapi"
	"github.com/leg100/otf/internal"
)

var codes = map[error]int{
	internal.ErrResourceNotFound:        http.StatusNotFound,
	internal.ErrAccessNotPermitted:      http.StatusForbidden,
	internal.ErrInvalidTerraformVersion: http.StatusUnprocessableEntity,
	internal.ErrResourceAlreadyExists:   http.StatusConflict,
	internal.ErrConflict:                http.StatusConflict,
}

// lookupHTTPCode maps an OTF domain error to a http status code
func lookupHTTPCode(err error) int {
	for otfError, httpError := range codes {
		if errors.Is(err, otfError) {
			return httpError
		}
	}
	return http.StatusInternalServerError
}

type ErrorOption func(*jsonapi.Error)

func WithStatus(httpStatusCode int) ErrorOption {
	return func(err *jsonapi.Error) {
		err.Status = &httpStatusCode
	}
}

// Error writes an HTTP response with a JSON-API encoded error.
func Error(w http.ResponseWriter, err error, opts ...ErrorOption) {
	jsonapiError := &jsonapi.Error{
		Detail: err.Error(),
	}
	for _, fn := range opts {
		fn(jsonapiError)
	}
	if jsonapiError.Status == nil {
		var missing *internal.ErrMissingParameter
		if errors.As(err, &missing) {
			// report missing parameter errors as a 422
			jsonapiError.Status = internal.Int(http.StatusUnprocessableEntity)
		} else {
			jsonapiError.Status = internal.Int(lookupHTTPCode(err))
		}
	}
	jsonapiError.Title = http.StatusText(*jsonapiError.Status)

	b, err := jsonapi.Marshal(jsonapiError)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-type", mediaType)
	w.WriteHeader(*jsonapiError.Status)
	w.Write(b)
}

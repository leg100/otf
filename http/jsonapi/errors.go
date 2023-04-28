package jsonapi

import (
	"errors"
	"net/http"

	"github.com/DataDog/jsonapi"
	"github.com/leg100/otf"
)

const mediaType = "application/vnd.api+json"

var codes = map[error]int{
	otf.ErrResourceNotFound:         http.StatusNotFound,
	otf.ErrAccessNotPermitted:       http.StatusForbidden,
	otf.ErrUploadTooLarge:           http.StatusUnprocessableEntity,
	otf.ErrResourceAlreadyExists:    http.StatusConflict,
	otf.ErrWorkspaceAlreadyLocked:   http.StatusConflict,
	otf.ErrWorkspaceAlreadyUnlocked: http.StatusConflict,
	otf.ErrRunDiscardNotAllowed:     http.StatusConflict,
	otf.ErrRunCancelNotAllowed:      http.StatusConflict,
	otf.ErrRunForceCancelNotAllowed: http.StatusConflict,
}

func lookupHTTPCode(err error) int {
	if v, ok := codes[err]; ok {
		return v
	}
	return http.StatusInternalServerError
}

// Error writes an HTTP response with a JSON-API encoded error.
func Error(w http.ResponseWriter, err error) {
	var (
		httpError *otf.HTTPError
		missing   *otf.MissingParameterError
		code      int
	)
	// If error is type otf.HTTPError then extract its status code
	if errors.As(err, &httpError) {
		code = httpError.Code
	} else if errors.As(err, &missing) {
		// report missing parameter errors as a 422
		code = http.StatusUnprocessableEntity
	} else {
		code = lookupHTTPCode(err)
	}
	b, err := jsonapi.Marshal(&jsonapi.Error{
		Status: &code,
		Title:  http.StatusText(code),
		Detail: err.Error(),
	})
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-type", mediaType)
	w.WriteHeader(code)
	w.Write(b)
}

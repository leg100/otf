package tfeapi

import (
	"errors"
	"net/http"

	"github.com/DataDog/jsonapi"
	"github.com/leg100/otf/internal"
)

var codes = map[error]int{
	internal.ErrResourceNotFound:         http.StatusNotFound,
	internal.ErrAccessNotPermitted:       http.StatusForbidden,
	internal.ErrUploadTooLarge:           http.StatusUnprocessableEntity,
	internal.ErrInvalidTerraformVersion:  http.StatusUnprocessableEntity,
	internal.ErrResourceAlreadyExists:    http.StatusConflict,
	internal.ErrWorkspaceAlreadyLocked:   http.StatusConflict,
	internal.ErrWorkspaceAlreadyUnlocked: http.StatusConflict,
	internal.ErrWorkspaceLockedByRun:     http.StatusConflict,
	internal.ErrRunDiscardNotAllowed:     http.StatusConflict,
	internal.ErrRunCancelNotAllowed:      http.StatusConflict,
	internal.ErrRunForceCancelNotAllowed: http.StatusConflict,
	internal.ErrConflict:                 http.StatusConflict,
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
		httpError *internal.HTTPError
		missing   *internal.MissingParameterError
		code      int
	)
	// If error is type internal.HTTPError then extract its status code
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

package jsonapi

import (
	"net/http"
	"strconv"

	"github.com/leg100/jsonapi"
)

// Error writes an HTTP response with a JSON-API marshalled error obj.
func Error(w http.ResponseWriter, code int, err error) {
	w.Header().Set("Content-type", jsonapi.MediaType)
	w.WriteHeader(code)
	jsonapi.MarshalErrors(w, []*jsonapi.ErrorObject{
		{
			Status: strconv.Itoa(code),
			Title:  http.StatusText(code),
			Detail: err.Error(),
		},
	})
}

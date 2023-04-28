// Package jsonapi handles marshaling/unmarshaling into/from json-api
package jsonapi

import (
	"io"
	"net/http"
	"strings"

	"github.com/DataDog/jsonapi"
	"github.com/leg100/otf"
)

// Pagination is used to return the pagination details of an API request.
type Pagination struct {
	CurrentPage  int  `json:"current-page"`
	PreviousPage *int `json:"prev-page"`
	NextPage     *int `json:"next-page"`
	TotalPages   int  `json:"total-pages"`
	TotalCount   int  `json:"total-count"`
}

func NewPagination(p *otf.Pagination) *Pagination {
	return &Pagination{
		CurrentPage:  p.Opts.SanitizedPageNumber(),
		PreviousPage: p.PrevPage(),
		NextPage:     p.NextPage(),
		TotalPages:   p.TotalPages(),
		TotalCount:   p.Count,
	}
}

// NewPaginationFromJSONAPI constructs pagination from a json:api struct
func NewPaginationFromJSONAPI(json *Pagination) *otf.Pagination {
	return &otf.Pagination{
		Count: json.TotalCount,
		// we can't determine the page size so we'll just pass in 0 which
		// ListOptions interprets as the default page size
		Opts: otf.ListOptions{PageNumber: json.CurrentPage, PageSize: 0},
	}
}

// Unmarshal reads from r and unmarshals the contents into v.
func Unmarshal(r io.Reader, v any) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return jsonapi.Unmarshal(b, v)
}

func Marshal(v any) ([]byte, error) {
	return jsonapi.Marshal(v)
}

// WriteResponse writes an HTTP response with a JSON-API marshalled payload.
func WriteResponse(w http.ResponseWriter, r *http.Request, v any, opts ...func(http.ResponseWriter)) {
	b, err := jsonapi.Marshal(v)
	if err != nil {
		Error(w, err)
	}
	w.Header().Set("Content-type", mediaType)
	for _, o := range opts {
		o(w)
	}
	w.Write(b)
}

// WithCode is a helper func for writing an HTTP status code to a response
// stream.  For use with WriteResponse.
func WithCode(code int) func(w http.ResponseWriter) {
	return func(w http.ResponseWriter) {
		w.WriteHeader(code)
	}
}

func sanitizeIncludes(includes []string) (sanitized []string) {
	for _, i := range includes {
		sanitized = append(sanitized, strings.ReplaceAll(i, "_", "-"))
	}
	return
}

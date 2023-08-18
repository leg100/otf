package tfeapi

import (
	"net/http"

	"github.com/DataDog/jsonapi"
	"github.com/leg100/otf/internal/resource"
)

const mediaType = "application/vnd.api+json"

// Responder handles responding to API requests.
type Responder struct {
	*includer
}

func NewResponder() *Responder {
	return &Responder{
		includer: &includer{
			registrations: make(map[IncludeName][]IncludeFunc),
		},
	}
}

func (res *Responder) RespondWithPage(w http.ResponseWriter, r *http.Request, items any, pagination *resource.Pagination) {
	meta := jsonapi.MarshalMeta(map[string]*resource.Pagination{
		"pagination": pagination,
	})
	res.Respond(w, r, items, http.StatusOK, meta)
}

func (res *Responder) Respond(w http.ResponseWriter, r *http.Request, payload any, status int, opts ...jsonapi.MarshalOption) {
	includes, err := res.addIncludes(r, payload)
	if err != nil {
		Error(w, err)
		return
	}
	if len(includes) > 0 {
		opts = append(opts, jsonapi.MarshalInclude(includes...))
	}
	b, err := jsonapi.Marshal(payload, opts...)
	if err != nil {
		Error(w, err)
		return
	}
	w.Header().Set("Content-type", mediaType)
	w.WriteHeader(status)
	w.Write(b)
}

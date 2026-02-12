package tfeapi

import (
	"net/http"

	"github.com/DataDog/jsonapi"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/resource"
)

const mediaType = "application/vnd.api+json"

// Responder handles responding to API requests.
type Responder struct {
	*includer
	logger logr.Logger
}

func NewResponder(logger logr.Logger) *Responder {
	return &Responder{
		includer: &includer{
			registrations: make(map[IncludeName][]IncludeFunc),
		},
		logger: logger,
	}
}

func (res *Responder) RespondWithPage(w http.ResponseWriter, r *http.Request, items any, pagination *resource.Pagination) {
	meta := jsonapi.MarshalMeta(map[string]*resource.Pagination{
		"pagination": pagination,
	})
	res.Respond(w, r, items, http.StatusOK, meta)
}

func (res *Responder) Respond(w http.ResponseWriter, r *http.Request, payload any, status int, opts ...jsonapi.MarshalOption) {

	if err := res.do(w, r, payload, status, opts...); err != nil {
		res.logger.Error(err, "sending API response", "url", r.URL)

		Error(w, err)
		return
	}
}

func (res *Responder) do(w http.ResponseWriter, r *http.Request, payload any, status int, opts ...jsonapi.MarshalOption) error {
	includes, err := res.addIncludes(r, payload)
	if err != nil {
		return err
	}
	opts = append(internal.DefaultJSONAPIMarshalOptions, opts...)
	if len(includes) > 0 {
		opts = append(opts, jsonapi.MarshalInclude(includes...))
	}
	b, err := jsonapi.Marshal(payload, opts...)
	if err != nil {
		return err
	}
	w.Header().Set("Content-type", mediaType)
	w.WriteHeader(status)
	w.Write(b)
	return nil
}

package http

import (
	"encoding/json"
	"io"

	"github.com/google/jsonapi"
	"github.com/leg100/ots"
)

func MarshalPayload(w io.Writer, models interface{}, include ...string) (err error) {
	var payload interface{}

	// Handle pagination links
	if paginated, ok := models.(ots.Paginated); ok {
		payload, err = jsonapi.Marshal(paginated.GetItems())
		if err != nil {
			return err
		}

		pagination := ots.NewPagination(paginated)
		payload.(*jsonapi.ManyPayload).Links = pagination.JSONAPIPaginationLinks()
		payload.(*jsonapi.ManyPayload).Meta = pagination.JSONAPIPaginationMeta()
	} else {
		payload, err = jsonapi.Marshal(models)
		if err != nil {
			return err
		}
	}

	switch payload := payload.(type) {
	case *jsonapi.ManyPayload:
		payload.Included = filterIncluded(payload.Included, include...)
	case *jsonapi.OnePayload:
		payload.Included = filterIncluded(payload.Included, include...)
	}

	return json.NewEncoder(w).Encode(payload)
}

func filterIncluded(included []*jsonapi.Node, filters ...string) (filtered []*jsonapi.Node) {
	for _, inc := range included {
		for _, f := range filters {
			if inc.Type == f {
				filtered = append(filtered, inc)
			}
		}
	}
	return filtered
}

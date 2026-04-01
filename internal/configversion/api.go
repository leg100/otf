package configversion

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/tfeapi"
)

type API struct {
	*tfeapi.Responder
	Client apiClient
}

type apiClient interface {
	DownloadConfig(ctx context.Context, id resource.ID) ([]byte, error)
}

func (a *API) AddHandlers(r *mux.Router) {
	r.HandleFunc("/configuration-versions/{id}/download", a.download).Methods("GET")
}

func (a *API) download(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	resp, err := a.Client.DownloadConfig(r.Context(), id)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	w.Write(resp)
}

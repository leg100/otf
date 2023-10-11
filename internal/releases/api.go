package releases

import (
	"net/http"

	"github.com/gorilla/mux"
	otfhttp "github.com/leg100/otf/internal/http"
)

type api struct {
	svc Service
}

func (a *api) addHandlers(r *mux.Router) {
	otfhttp.APIRouter(r).HandleFunc("/releases/latest", a.getLatest).Methods("GET")
}

func (a *api) getLatest(w http.ResponseWriter, r *http.Request) {
	v, _, err := a.svc.getLatest(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if _, err := w.Write([]byte(v)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

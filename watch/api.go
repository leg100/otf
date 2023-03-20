package watch

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	otfhttp "github.com/leg100/otf/http"
	"github.com/leg100/otf/http/jsonapi"
	"github.com/leg100/otf/run"
)

type api struct {
	logr.Logger

	svc Service
}

func (a *api) addHandlers(r *mux.Router) {
	r = otfhttp.APIRouter(r)

	r.HandleFunc("/watch", a.watch).Methods("GET")
}

// Watch handler responds with a stream of events, using json encoding.
//
// NOTE: Only run events are currently supported.
func (a *api) watch(w http.ResponseWriter, r *http.Request) {
	// TODO: populate watch options
	events, err := a.svc.Watch(r.Context(), otf.WatchOptions{})
	if err != nil {
		jsonapi.Error(w, http.StatusInternalServerError, err)
		return
	}

	rc := http.NewResponseController(w)
	w.Header().Set("Content-Type", "text/event-stream")

	for {
		select {
		case event, ok := <-events:
			if !ok {
				// server closed connection
				return
			}

			// Only run events are supported
			run, ok := event.Payload.(*run.Run)
			if !ok {
				continue
			}

			data, err := json.Marshal(run)
			if err != nil {
				a.Error(err, "marshalling event", "event", event.Type)
				continue
			}

			fmt.Fprintf(w, "data: %s\n", string(data))
			fmt.Fprintf(w, "event: %s\n", event.Type)
			fmt.Fprintln(w)

			rc.Flush()
		case <-r.Context().Done():
			// client closed connection
			return
		}
	}
}

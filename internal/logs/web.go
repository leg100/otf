package logs

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/http/decode"
	otfhtml "github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/pubsub"
)

const (
	EventLogChunk    string = "log_update"
	EventLogFinished string = "log_finished"
)

type (
	webHandlers struct {
		logr.Logger

		svc tailService
	}

	tailService interface {
		Tail(ctx context.Context, opts TailOptions) (<-chan Chunk, error)
	}
)

func (h *webHandlers) addHandlers(r *mux.Router) {
	r = otfhtml.UIRouter(r)

	r.HandleFunc("/runs/{run_id}/tail", h.tailRun)
}

func (h *webHandlers) tailRun(w http.ResponseWriter, r *http.Request) {
	var params TailOptions
	if err := decode.All(&params, r); err != nil {
		otfhtml.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ch, err := h.svc.Tail(r.Context(), params)
	if err != nil {
		otfhtml.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	rc := http.NewResponseController(w)
	rc.Flush()

	for {
		select {
		case chunk, ok := <-ch:
			if !ok {
				// no more logs
				pubsub.WriteSSEEvent(w, []byte("no more logs"), EventLogFinished, false)
				return
			}
			html := chunk.ToHTML()
			if len(html) == 0 {
				// don't send empty chunks
				continue
			}
			js, err := json.Marshal(struct {
				HTML       string `json:"html"`
				NextOffset int    `json:"offset"`
			}{
				HTML:       string(html) + "<br>",
				NextOffset: chunk.NextOffset(),
			})
			if err != nil {
				h.Error(err, "marshalling data")
				continue
			}
			pubsub.WriteSSEEvent(w, js, EventLogChunk, false)
			rc.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

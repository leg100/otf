package logs

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/pubsub"
)

type (
	webHandlers struct {
		logr.Logger

		svc tailService
	}

	tailService interface {
		Tail(ctx context.Context, opts internal.GetChunkOptions) (<-chan internal.Chunk, error)
	}
)

func (h *webHandlers) addHandlers(r *mux.Router) {
	r = html.UIRouter(r)

	r.HandleFunc("/runs/{run_id}/tail", h.tailRun)
}

func (h *webHandlers) tailRun(w http.ResponseWriter, r *http.Request) {
	var params struct {
		// ID of run to tail
		RunID string `schema:"run_id,required"`
		// Phase to tail. Must be either plan or apply.
		Phase internal.PhaseType `schema:"phase,required"`
		// Offset is number of bytes into logs to start tailing from
		Offset int `schema:"offset,required"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity, false)
		return
	}

	ch, err := h.svc.Tail(r.Context(), internal.GetChunkOptions{
		RunID:  params.RunID,
		Phase:  params.Phase,
		Offset: params.Offset,
	})
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError, false)
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
				pubsub.WriteSSEEvent(w, []byte("no more logs"), pubsub.EventLogFinished, false)
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
			pubsub.WriteSSEEvent(w, js, pubsub.EventLogChunk, false)
			rc.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

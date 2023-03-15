package logs

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/html"
)

type (
	tailService interface {
		tail(ctx context.Context, opts otf.GetChunkOptions) (<-chan otf.Chunk, error)
	}

	webHandlers struct {
		logr.Logger

		svc tailService
	}
)

func newWebHandlers(svc tailService, logger logr.Logger) *webHandlers {
	return &webHandlers{
		Logger: logger,
		svc:    svc,
	}
}

func (h *webHandlers) addHandlers(r *mux.Router) {
	r.HandleFunc("/runs/{run_id}/tail", h.tailRun)
}

func (h *webHandlers) tailRun(w http.ResponseWriter, r *http.Request) {
	var params struct {
		// ID of run to tail
		RunID string `schema:"run_id,required"`
		// Phase to tail. Must be either plan or apply.
		Phase otf.PhaseType `schema:"phase,required"`
		// Offset is number of bytes into logs to start tailing from
		Offset int `schema:"offset,required"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	rc := http.NewResponseController(w)

	ch, err := h.svc.tail(r.Context(), otf.GetChunkOptions{
		RunID:  params.RunID,
		Phase:  params.Phase,
		Offset: params.Offset,
	})
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")

	for {
		select {
		case chunk, ok := <-ch:
			if !ok {
				// no more logs
				fmt.Fprintln(w, "event: finished")
				fmt.Fprintln(w, "data: no more logs")
				fmt.Fprintln(w)
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
			fmt.Fprintf(w, "data: %s\n", string(js))
			fmt.Fprintln(w, "event: new-log-chunk")
			fmt.Fprintln(w)
			if err := rc.Flush(); err != nil {
				return
			}
		case <-r.Context().Done():
			return
		}
	}
}

package logs

import (
	"encoding/json"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/html"
	"github.com/r3labs/sse/v2"
)

type webHandlers struct {
	logr.Logger
	*sse.Server

	service
}

func newWebHandlers(app service, logger logr.Logger) *webHandlers {
	// Create and configure SSE server
	srv := sse.New()
	// we don't use last-event-item functionality so turn it off
	srv.AutoReplay = false
	// encode payloads into base64 otherwise the JSON string payloads corrupt
	// the SSE protocol
	srv.EncodeBase64 = true

	return &webHandlers{
		Server:  srv,
		Logger:  logger,
		service: app,
	}
}

func (h *webHandlers) addHandlers(r *mux.Router) {
	r.HandleFunc("/runs/{run_id}/tail", h.tailRun)
}

func (h *webHandlers) tailRun(w http.ResponseWriter, r *http.Request) {
	var params struct {
		// Phase to tail. Must be either plan or apply.
		Phase otf.PhaseType `schema:"phase,required"`
		// Offset is number of bytes into logs to start tailing from
		Offset int `schema:"offset,required"`
		// StreamID is the ID of the SSE stream
		StreamID string `schema:"stream,required"`
		// ID of run to tail
		RunID string `schema:"run_id,required"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ch, err := h.service.tail(r.Context(), otf.GetChunkOptions{
		RunID:  params.RunID,
		Phase:  params.Phase,
		Offset: params.Offset,
	})
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	go func() {
		for {
			select {
			case chunk, ok := <-ch:
				if !ok {
					// no more logs
					h.Server.Publish(params.StreamID, &sse.Event{
						Event: []byte("finished"),
						Data:  []byte("no more logs"),
					})
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
				h.Server.Publish(params.StreamID, &sse.Event{
					Data:  js,
					Event: []byte("new-log-chunk"),
				})
			case <-r.Context().Done():
				return
			}
		}
	}()
	h.Server.ServeHTTP(w, r)
}

package logs

import (
	"encoding/json"
	"html/template"
	"net/http"

	term2html "github.com/buildkite/terminal-to-html"
	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/html"
	"github.com/r3labs/sse/v2"
)

type web struct {
	logr.Logger
	*sse.Server

	app
}

func newWebHandlers(app app, logger logr.Logger) *web {
	// Create and configure SSE server
	srv := sse.New()
	// we don't use last-event-item functionality so turn it off
	srv.AutoReplay = false
	// encode payloads into base64 otherwise the JSON string payloads corrupt
	// the SSE protocol
	srv.EncodeBase64 = true

	return &web{
		Server: srv,
		Logger: logger,
		app:    app,
	}
}

func (h *web) addHandlers(r *mux.Router) {
	r.HandleFunc("/runs/{run_id}/tail", h.tailRun)
}

func (h *web) tailRun(w http.ResponseWriter, r *http.Request) {
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

	ch, err := h.app.tail(r.Context(), GetChunkOptions{
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
				htmlChunk := &htmlLogChunk{chunk}
				html := htmlChunk.ToHTML()
				if len(html) == 0 {
					// don't send empty chunks
					continue
				}
				js, err := json.Marshal(struct {
					HTML       string `json:"html"`
					NextOffset int    `json:"offset"`
				}{
					HTML:       string(html) + "<br>",
					NextOffset: htmlChunk.NextOffset(),
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

// htmlLogChunk is a log chunk rendered in html
type htmlLogChunk struct {
	Chunk
}

func (l *htmlLogChunk) ToHTML() template.HTML {
	chunk := l.RemoveStartMarker()
	chunk = chunk.RemoveEndMarker()

	// convert ANSI escape sequences to HTML
	data := string(term2html.Render(chunk.Data))

	return template.HTML(data)
}

// NextOffset returns the offset for the next chunk
func (l *htmlLogChunk) NextOffset() int {
	return l.Chunk.Offset + len(l.Chunk.Data)
}

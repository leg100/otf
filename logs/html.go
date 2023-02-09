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

type htmlHandlers struct {
	logr.Logger
	*sse.Server

	app
}

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

func (h *htmlHandlers) AddHandlers(r *mux.Router) {
	r.HandleFunc("/runs/{run_id}/tail", h.tailRun)
}

func (app *htmlHandlers) tailRun(w http.ResponseWriter, r *http.Request) {
	params := struct {
		// Phase to tail. Must be either plan or apply.
		Phase otf.PhaseType `schema:"phase,required"`
		// Offset is number of bytes into logs to start tailing from
		Offset int `schema:"offset,required"`
		// StreamID is the ID of the SSE stream
		StreamID string `schema:"stream,required"`
		// ID of run to tail
		RunID string `schema:"run_id,required"`
	}{}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ch, err := app.Tail(r.Context(), GetChunkOptions{
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
					app.Server.Publish(params.StreamID, &sse.Event{
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
					app.Error(err, "marshalling data")
					continue
				}
				app.Server.Publish(params.StreamID, &sse.Event{
					Data:  js,
					Event: []byte("new-log-chunk"),
				})
			case <-r.Context().Done():
				return
			}
		}
	}()
	app.Server.ServeHTTP(w, r)
}

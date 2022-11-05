package html

import (
	"encoding/json"
	"html/template"
	"net/http"

	term2html "github.com/buildkite/terminal-to-html"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/r3labs/sse/v2"
)

type htmlLogChunk struct {
	otf.Chunk
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

func (app *Application) listRuns(w http.ResponseWriter, r *http.Request) {
	opts := otf.RunListOptions{
		// We don't list speculative runs on the UI
		Speculative: otf.Bool(false),
	}
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	if err := decode.Route(&opts, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	runs, err := app.ListRuns(r.Context(), opts)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	streamID := "watch-ws-runs-" + otf.GenerateRandomString(5)
	app.render("run_list.tmpl", w, r, runList{runs, opts, streamID})
}

func (app *Application) newRun(w http.ResponseWriter, r *http.Request) {
	app.render("run_new.tmpl", w, r, struct {
		Organization string
		Workspace    string
	}{
		Organization: mux.Vars(r)["organization_name"],
		Workspace:    mux.Vars(r)["workspace_name"],
	})
}

func (app *Application) getRun(w http.ResponseWriter, r *http.Request) {
	run, err := app.GetRun(r.Context(), mux.Vars(r)["run_id"])
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Get existing logs thus far received for each phase. If none are found then don't treat
	// that as an error because it merely means no logs have yet been received.
	planLogs, err := app.GetChunk(r.Context(), otf.GetChunkOptions{
		RunID: run.ID(),
		Phase: otf.PlanPhase,
	})
	if err != nil && err != otf.ErrResourceNotFound {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	applyLogs, err := app.GetChunk(r.Context(), otf.GetChunkOptions{
		RunID: run.ID(),
		Phase: otf.ApplyPhase,
	})
	if err != nil && err != otf.ErrResourceNotFound {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	app.render("run_get.tmpl", w, r, struct {
		Run       *otf.Run
		PlanLogs  *htmlLogChunk
		ApplyLogs *htmlLogChunk
	}{
		Run:       run,
		PlanLogs:  &htmlLogChunk{planLogs},
		ApplyLogs: &htmlLogChunk{applyLogs},
	})
}

func (app *Application) deleteRun(w http.ResponseWriter, r *http.Request) {
	err := app.DeleteRun(r.Context(), mux.Vars(r)["run_id"])
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, getWorkspacePath(workspaceRequest{r}), http.StatusFound)
}

func (app *Application) cancelRun(w http.ResponseWriter, r *http.Request) {
	err := app.CancelRun(r.Context(), mux.Vars(r)["run_id"], otf.RunCancelOptions{})
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, listRunPath(workspaceRequest{r}), http.StatusFound)
}

func (app *Application) tailRun(w http.ResponseWriter, r *http.Request) {
	opts := struct {
		// Phase to tail. Must be either plan or apply.
		Phase otf.PhaseType `schema:"phase,required"`
		// Offset is number of bytes into logs to start tailing from
		Offset int `schema:"offset,required"`
		// StreamID is the ID of the SSE stream
		StreamID string `schema:"stream,required"`
	}{}
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	ch, err := app.Tail(r.Context(), otf.GetChunkOptions{
		RunID:  mux.Vars(r)["run_id"],
		Phase:  opts.Phase,
		Offset: opts.Offset,
	})
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	go func() {
		for {
			select {
			case chunk, ok := <-ch:
				if !ok {
					// no more logs
					app.Server.Publish(opts.StreamID, &sse.Event{
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
				app.Server.Publish(opts.StreamID, &sse.Event{
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

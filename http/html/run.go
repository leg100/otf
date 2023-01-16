package html

import (
	"encoding/json"
	"errors"
	"html/template"
	"net/http"

	term2html "github.com/buildkite/terminal-to-html"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/html/paths"
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
	type parameters struct {
		WorkspaceID string `schema:"workspace_id,required"`
		otf.ListOptions
	}
	var params parameters
	if err := decode.All(&params, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := app.GetWorkspace(r.Context(), params.WorkspaceID)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	runs, err := app.ListRuns(r.Context(), otf.RunListOptions{
		ListOptions: params.ListOptions,
		Speculative: otf.Bool(false), // we don't list speculative runs on the UI
		WorkspaceID: &params.WorkspaceID,
	})
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.render("run_list.tmpl", w, r, struct {
		*otf.RunList
		*otf.Workspace
		StreamID string
	}{
		RunList:   runs,
		Workspace: ws,
		StreamID:  "watch-ws-runs-" + otf.GenerateRandomString(5),
	})
}

func (app *Application) getRun(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.Param("run_id", r)
	if err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	run, err := app.GetRun(r.Context(), runID)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ws, err := app.GetWorkspace(r.Context(), run.WorkspaceID())
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
	if err != nil && !errors.Is(err, otf.ErrResourceNotFound) {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	applyLogs, err := app.GetChunk(r.Context(), otf.GetChunkOptions{
		RunID: run.ID(),
		Phase: otf.ApplyPhase,
	})
	if err != nil && !errors.Is(err, otf.ErrResourceNotFound) {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	app.render("run_get.tmpl", w, r, struct {
		*otf.Run
		Workspace *otf.Workspace
		PlanLogs  *htmlLogChunk
		ApplyLogs *htmlLogChunk
	}{
		Run:       run,
		Workspace: ws,
		PlanLogs:  &htmlLogChunk{planLogs},
		ApplyLogs: &htmlLogChunk{applyLogs},
	})
}

func (app *Application) deleteRun(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.Param("run_id", r)
	if err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	run, err := app.GetRun(r.Context(), runID)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = app.DeleteRun(r.Context(), runID)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, paths.Workspace(run.WorkspaceID()), http.StatusFound)
}

func (app *Application) cancelRun(w http.ResponseWriter, r *http.Request) {
	runID, err := decode.Param("run_id", r)
	if err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	run, err := app.GetRun(r.Context(), runID)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = app.CancelRun(r.Context(), runID, otf.RunCancelOptions{})
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, paths.Runs(run.WorkspaceID()), http.StatusFound)
}

func (app *Application) applyRun(w http.ResponseWriter, r *http.Request) {
	run, err := app.GetRun(r.Context(), mux.Vars(r)["run_id"])
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = app.ApplyRun(r.Context(), mux.Vars(r)["run_id"], otf.RunApplyOptions{})
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, paths.Run(run.ID())+"#apply", http.StatusFound)
}

func (app *Application) discardRun(w http.ResponseWriter, r *http.Request) {
	run, err := app.GetRun(r.Context(), mux.Vars(r)["run_id"])
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = app.DiscardRun(r.Context(), mux.Vars(r)["run_id"], otf.RunDiscardOptions{})
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, paths.Run(run.ID()), http.StatusFound)
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

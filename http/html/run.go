package html

import (
	"encoding/json"
	"net/http"

	term2html "github.com/buildkite/terminal-to-html"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/r3labs/sse/v2"
)

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

func (app *Application) createRun(w http.ResponseWriter, r *http.Request) {
	var opts otf.RunCreateOptions
	if err := decode.Route(&opts, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	if err := decode.Form(&opts, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	ws := workspaceRequest{r}.Spec()
	created, err := app.CreateRun(r.Context(), ws, opts)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, getRunPath(created), http.StatusFound)
}

func (app *Application) getRun(w http.ResponseWriter, r *http.Request) {
	run, err := app.GetRun(r.Context(), mux.Vars(r)["run_id"])
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Get existing logs thus far received for each phase. If none are found then don't treat
	// that as an error because it merely means no logs have yet been received.
	planLogs, err := app.GetChunk(r.Context(), run.ID(), otf.PlanPhase, otf.GetChunkOptions{})
	if err != nil && err != otf.ErrResourceNotFound {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	applyLogs, err := app.GetChunk(r.Context(), run.ID(), otf.ApplyPhase, otf.GetChunkOptions{})
	if err != nil && err != otf.ErrResourceNotFound {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	app.render("run_get.tmpl", w, r, struct {
		Run       *otf.Run
		PlanLogs  *logs
		ApplyLogs *logs
	}{
		Run:       run,
		PlanLogs:  (*logs)(&planLogs),
		ApplyLogs: (*logs)(&applyLogs),
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
	client, err := app.Tail(r.Context(), mux.Vars(r)["run_id"], opts.Phase, opts.Offset)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	go func() {
		// keep running tally of offset
		offset := opts.Offset

		for {
			select {
			case <-r.Context().Done():
				app.Info("closing client")
				client.Close()
				return
			case chunk, ok := <-client.Read():
				app.Info("relaying chunk to client", "data", string(chunk), "phase", opts.Phase, "offset", offset, "stream", opts.StreamID)
				if !ok {
					// no more logs
					app.Server.Publish(opts.StreamID, &sse.Event{
						Event: []byte("finished"),
						Data:  []byte("no more logs"),
					})
					return
				}

				offset += len(chunk)

				js, err := json.Marshal(struct {
					Offset int    `json:"offset"`
					HTML   string `json:"html"`
				}{
					Offset: offset,
					HTML:   string(term2html.Render(chunk)) + "<br>",
				})
				if err != nil {
					app.Error(err, "marshalling data")
					continue
				}
				app.Server.Publish(opts.StreamID, &sse.Event{
					Data:  js,
					Event: []byte("new-log-chunk"),
				})
			}
		}
	}()
	app.Server.ServeHTTP(w, r)
}

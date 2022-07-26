package html

import (
	"bytes"
	"html/template"
	"net/http"
	"strings"

	term2html "github.com/buildkite/terminal-to-html"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/r3labs/sse/v2"
)

// LatestOptions are the options for watching the latest run for a workspace
type LatestOptions struct {
	// StreamID is the ID of the SSE stream
	StreamID string `schema:"stream,required"`
}

// TailOptions are the options for tailing logs for a run phase
type TailOptions struct {
	// Offset is number of bytes into logs to start tailing from
	Offset int `schema:"offset,required"`
	// StreamID is the ID of the SSE stream
	StreamID string `schema:"stream,required"`
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
	app.render("run_list.tmpl", w, r, runList{runs, opts})
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
	app.render("run_get.tmpl", w, r, run)
}

func (app *Application) watchLatestRun(w http.ResponseWriter, r *http.Request) {
	var spec otf.WorkspaceSpec
	if err := decode.Route(&spec, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	var opts LatestOptions
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	updates, err := app.WatchLatest(r.Context(), spec)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.Server.CreateStream(opts.StreamID)
	defer app.Server.RemoveStream(opts.StreamID)
	go func() {
		for {
			select {
			case <-r.Context().Done():
				return
			case run := <-updates:
				buf := new(bytes.Buffer)
				if err := app.renderTemplate("run_item.tmpl", buf, run); err != nil {
					app.Error(err, "rendering template for watched run")
					continue
				}
				// remove newlines otherwise sse interprets each line as a new
				// event
				content := strings.ReplaceAll(buf.String(), "\n", "")
				app.Server.Publish("messages", &sse.Event{Data: []byte(content)})
			}
		}
	}()
	app.Server.ServeHTTP(w, r)
}

func (app *Application) getPhase(phase string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		run, err := app.GetRun(r.Context(), mux.Vars(r)["run_id"])
		if err != nil {
			writeError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		chunk, err := app.GetChunk(r.Context(), run.ID(), otf.PhaseType(phase), otf.GetChunkOptions{})
		if err != nil {
			writeError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		app.render("phase_get.tmpl", w, r, struct {
			Run      *otf.Run
			Logs     template.HTML
			Phase    string
			Offset   int
			StreamID string
		}{
			Run:   run,
			Logs:  logsToHTML(chunk.Data),
			Phase: phase,
			// Add one to account for start marker
			Offset: len(chunk.Data) + 1,
			// Setup SSE stream with unique name because stream is unique to client
			StreamID: "tail-" + otf.GenerateRandomString(5),
		})
	}
}

func (app *Application) tailPhase(phase string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var opts TailOptions
		if err := decode.Query(&opts, r.URL.Query()); err != nil {
			writeError(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		run, err := app.GetRun(r.Context(), mux.Vars(r)["run_id"])
		if err != nil {
			writeError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		client, err := app.Tail(r.Context(), run.ID(), otf.PhaseType(phase), opts.Offset)
		if err != nil {
			writeError(w, err.Error(), http.StatusInternalServerError)
			return
		}

		app.Server.CreateStream(opts.StreamID)
		defer app.Server.RemoveStream(opts.StreamID)
		go func() {
			for {
				select {
				case <-r.Context().Done():
					client.Close()
					return
				case chunk, ok := <-client.Read():
					if !ok {
						return
					}
					html := string(term2html.Render(chunk))
					html = strings.ReplaceAll(html, "\n", "<br>")
					html = html + "<br>"
					app.Server.Publish(opts.StreamID, &sse.Event{Data: []byte(html)})
				}
			}
		}()
		app.Server.ServeHTTP(w, r)
	}
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

func logsToHTML(data []byte) template.HTML {
	// convert to string
	logs := string(data)
	// trim leading and trailing white space
	logs = strings.TrimSpace(logs)
	// convert ANSI escape sequences to HTML
	logs = string(term2html.Render([]byte(logs)))
	// trim leading and trailing white space
	logs = strings.TrimSpace(logs)

	return template.HTML(logs)
}

package html

import (
	"html/template"
	"net/http"
	"strings"

	term2html "github.com/buildkite/terminal-to-html"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
)

func (app *Application) listRuns(w http.ResponseWriter, r *http.Request) {
	// get runs
	var opts otf.RunListOptions
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	if err := decode.Route(&opts, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	runs, err := app.RunService().List(r.Context(), opts)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	app.render("run_list.tmpl", w, r, struct {
		List      *otf.RunList
		Workspace workspaceRoute
	}{
		List:      runs,
		Workspace: workspaceRequest{r},
	})
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
	created, err := app.RunService().Create(r.Context(), ws, opts)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, getRunPath(created), http.StatusFound)
}

func (app *Application) getRun(w http.ResponseWriter, r *http.Request) {
	run, err := app.RunService().Get(r.Context(), mux.Vars(r)["run_id"])
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	app.render("run_get.tmpl", w, r, run)
}

func (app *Application) getPlan(w http.ResponseWriter, r *http.Request) {
	run, err := app.RunService().Get(r.Context(), mux.Vars(r)["run_id"])
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	chunk, err := app.PlanService().GetChunk(r.Context(), run.Plan.ID(), otf.GetChunkOptions{})
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// convert to string
	logs := string(chunk.Data)
	// trim leading and trailing white space
	logs = strings.TrimSpace(logs)
	// convert ANSI escape sequences to HTML
	logs = string(term2html.Render([]byte(logs)))
	// trim leading and trailing white space
	logs = strings.TrimSpace(logs)
	app.render("plan_get.tmpl", w, r, struct {
		Run  *otf.Run
		Logs template.HTML
	}{
		Run:  run,
		Logs: template.HTML(logs),
	})
}

func (app *Application) getApply(w http.ResponseWriter, r *http.Request) {
	run, err := app.RunService().Get(r.Context(), mux.Vars(r)["run_id"])
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	chunk, err := app.ApplyService().GetChunk(r.Context(), run.Apply.ID(), otf.GetChunkOptions{})
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// convert to string
	logs := string(chunk.Data)
	// trim leading and trailing white space
	logs = strings.TrimSpace(logs)
	// convert ANSI escape sequences to HTML
	logs = string(term2html.Render([]byte(logs)))
	// trim leading and trailing white space
	logs = strings.TrimSpace(logs)
	app.render("apply_get.tmpl", w, r, struct {
		Run  *otf.Run
		Logs template.HTML
	}{
		Run:  run,
		Logs: template.HTML(logs),
	})
}

func (app *Application) deleteRun(w http.ResponseWriter, r *http.Request) {
	err := app.RunService().Delete(r.Context(), mux.Vars(r)["run_id"])
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, getWorkspacePath(workspaceRequest{r}), http.StatusFound)
}

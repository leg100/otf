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

func (c *Application) listRuns(w http.ResponseWriter, r *http.Request) {
	// get runs
	var opts otf.RunListOptions
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	if err := decode.Form(&opts, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	runs, err := c.RunService().List(r.Context(), opts)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// get workspace
	spec := otf.WorkspaceSpec{OrganizationName: opts.OrganizationName, Name: opts.WorkspaceName}
	workspace, err := c.WorkspaceService().Get(r.Context(), spec)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	c.render("run_list.tmpl", w, r, struct {
		List      *otf.RunList
		Options   otf.RunListOptions
		Workspace *otf.Workspace
	}{
		List:      runs,
		Options:   opts,
		Workspace: workspace,
	})
}

func (c *Application) newRun(w http.ResponseWriter, r *http.Request) {
	c.render("run_new.tmpl", w, r, struct {
		Organization string
		Workspace    string
	}{
		Organization: mux.Vars(r)["organization_name"],
		Workspace:    mux.Vars(r)["workspace_name"],
	})
}

func (c *Application) createRun(w http.ResponseWriter, r *http.Request) {
	var opts otf.RunCreateOptions
	if err := decode.Route(&opts, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	if err := decode.Form(&opts, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	created, err := c.RunService().Create(r.Context(), opts)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, c.relative(r, "getRun", "run_id", created.ID()), http.StatusFound)
}

func (c *Application) getRun(w http.ResponseWriter, r *http.Request) {
	run, err := c.RunService().Get(r.Context(), mux.Vars(r)["run_id"])
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	chunk, err := c.PlanService().GetChunk(r.Context(), run.Plan.ID(), otf.GetChunkOptions{})
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// convert to string
	logStr := string(chunk.Data)
	// trim leading and trailing white space
	logStr = strings.TrimSpace(logStr)
	// convert ANSI escape sequences to HTML
	logStr = string(term2html.Render([]byte(logStr)))
	// trim leading and trailing white space
	logStr = strings.TrimSpace(logStr)
	c.render("run_get.tmpl", w, r, struct {
		Run      *otf.Run
		PlanLogs template.HTML
	}{
		Run:      run,
		PlanLogs: template.HTML(logStr),
	})
}

func (c *Application) getPlan(w http.ResponseWriter, r *http.Request) {
	run, err := c.RunService().Get(r.Context(), mux.Vars(r)["run_id"])
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	chunk, err := c.PlanService().GetChunk(r.Context(), run.Plan.ID(), otf.GetChunkOptions{})
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
	c.render("plan_get.tmpl", w, r, struct {
		Run  *otf.Run
		Logs template.HTML
	}{
		Run:  run,
		Logs: template.HTML(logs),
	})
}

func (c *Application) getApply(w http.ResponseWriter, r *http.Request) {
	run, err := c.RunService().Get(r.Context(), mux.Vars(r)["run_id"])
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	chunk, err := c.ApplyService().GetChunk(r.Context(), run.Apply.ID(), otf.GetChunkOptions{})
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
	c.render("apply_get.tmpl", w, r, struct {
		Run  *otf.Run
		Logs template.HTML
	}{
		Run:  run,
		Logs: template.HTML(logs),
	})
}

func (c *Application) deleteRun(w http.ResponseWriter, r *http.Request) {
	err := c.RunService().Delete(r.Context(), mux.Vars(r)["run_id"])
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "../../", http.StatusFound)
}

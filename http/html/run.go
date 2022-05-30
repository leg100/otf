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

type RunController struct {
	otf.RunService
	otf.PlanService
	otf.ApplyService
	otf.WorkspaceService

	// HTML template renderer
	renderer

	*router

	// for setting flash messages
	sessions *sessions

	*templateDataFactory
}

func (c *RunController) addRoutes(router *mux.Router) {
	router.HandleFunc("/", c.List).Methods("GET").Name("listRun")
	router.HandleFunc("/new", c.New).Methods("GET").Name("newRun")
	router.HandleFunc("/create", c.Create).Methods("POST").Name("createRun")
	router.HandleFunc("/{run_id}", c.Get).Methods("GET").Name("getRun")
	router.HandleFunc("/{run_id}/plan", c.GetPlan).Methods("GET").Name("getPlan")
	router.HandleFunc("/{run_id}/apply", c.GetApply).Methods("GET").Name("getApply")
	router.HandleFunc("/{run_id}/delete", c.Delete).Methods("POST").Name("deleteRun")
}

func (c *RunController) List(w http.ResponseWriter, r *http.Request) {
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
	runs, err := c.RunService.List(r.Context(), opts)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// get workspace
	spec := otf.WorkspaceSpec{OrganizationName: opts.OrganizationName, Name: opts.WorkspaceName}
	workspace, err := c.WorkspaceService.Get(r.Context(), spec)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tdata := c.newTemplateData(r, struct {
		List      *otf.RunList
		Options   otf.RunListOptions
		Workspace *otf.Workspace
	}{
		List:      runs,
		Options:   opts,
		Workspace: workspace,
	})
	if err := c.renderTemplate("run_list.tmpl", w, tdata); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c *RunController) New(w http.ResponseWriter, r *http.Request) {
	tdata := c.newTemplateData(r, struct {
		Organization string
		Workspace    string
	}{
		Organization: mux.Vars(r)["organization_name"],
		Workspace:    mux.Vars(r)["workspace_name"],
	})

	if err := c.renderTemplate("run_new.tmpl", w, tdata); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c *RunController) Create(w http.ResponseWriter, r *http.Request) {
	var opts otf.RunCreateOptions
	if err := decode.Route(&opts, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	if err := decode.Form(&opts, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	created, err := c.RunService.Create(r.Context(), opts)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, c.relative(r, "getRun", "run_id", created.ID()), http.StatusFound)
}

func (c *RunController) Get(w http.ResponseWriter, r *http.Request) {
	run, err := c.RunService.Get(r.Context(), mux.Vars(r)["run_id"])
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	chunk, err := c.PlanService.GetChunk(r.Context(), run.Plan.ID(), otf.GetChunkOptions{})
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

	tdata := c.newTemplateData(r, struct {
		Run      *otf.Run
		PlanLogs template.HTML
	}{
		Run:      run,
		PlanLogs: template.HTML(logStr),
	})

	if err := c.renderTemplate("run_get.tmpl", w, tdata); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c *RunController) GetPlan(w http.ResponseWriter, r *http.Request) {
	run, err := c.RunService.Get(r.Context(), mux.Vars(r)["run_id"])
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	chunk, err := c.PlanService.GetChunk(r.Context(), run.Plan.ID(), otf.GetChunkOptions{})
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

	tdata := c.newTemplateData(r, struct {
		Run  *otf.Run
		Logs template.HTML
	}{
		Run:  run,
		Logs: template.HTML(logs),
	})

	if err := c.renderTemplate("plan_get.tmpl", w, tdata); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c *RunController) GetApply(w http.ResponseWriter, r *http.Request) {
	run, err := c.RunService.Get(r.Context(), mux.Vars(r)["run_id"])
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	chunk, err := c.ApplyService.GetChunk(r.Context(), run.Apply.ID(), otf.GetChunkOptions{})
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

	tdata := c.newTemplateData(r, struct {
		Run  *otf.Run
		Logs template.HTML
	}{
		Run:  run,
		Logs: template.HTML(logs),
	})

	if err := c.renderTemplate("apply_get.tmpl", w, tdata); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c *RunController) Delete(w http.ResponseWriter, r *http.Request) {
	err := c.RunService.Delete(r.Context(), mux.Vars(r)["run_id"])
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "../../", http.StatusFound)
}

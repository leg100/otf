package html

import (
	"net/http"
	"path"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
)

type RunController struct {
	otf.RunService

	// HTML template renderer
	renderer

	*templateDataFactory
}

func (c *RunController) addRoutes(router *mux.Router) {
	router = router.PathPrefix("/organizations/{organization_name}/workspaces/{workspace_name}/runs").Subrouter()

	router.HandleFunc("/", c.List).Methods("GET").Name("listRun")
	router.HandleFunc("/new", c.New).Methods("GET").Name("newRun")
	router.HandleFunc("/create", c.Create).Methods("POST").Name("createRun")
	router.HandleFunc("/{run_id}", c.Get).Methods("GET").Name("getRun")
	router.HandleFunc("/{run_id}/delete", c.Delete).Methods("POST").Name("deleteRun")
}

func (c *RunController) List(w http.ResponseWriter, r *http.Request) {
	var opts otf.RunListOptions
	if err := decodeAll(r, &opts); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
	}

	runs, err := c.RunService.List(r.Context(), opts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tdata := c.newTemplateData(r, struct {
		List    *otf.RunList
		Options otf.RunListOptions
	}{
		List:    runs,
		Options: opts,
	})

	if err := c.renderTemplate("runs_list.tmpl", w, tdata); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	if err := c.renderTemplate("runs_new.tmpl", w, tdata); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c *RunController) Create(w http.ResponseWriter, r *http.Request) {
	var opts otf.RunCreateOptions
	if err := decodeAll(r, &opts); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
	}

	created, err := c.RunService.Create(r.Context(), opts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, path.Join("..", created.ID), http.StatusFound)
}

func (c *RunController) Get(w http.ResponseWriter, r *http.Request) {
	run, err := c.RunService.Get(r.Context(), mux.Vars(r)["run_id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tdata := c.newTemplateData(r, run)

	if err := c.renderTemplate("runs_show.tmpl", w, tdata); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c *RunController) Delete(w http.ResponseWriter, r *http.Request) {
	err := c.RunService.Delete(r.Context(), mux.Vars(r)["run_id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "../../", http.StatusFound)
}

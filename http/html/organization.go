package html

import (
	"net/http"
	"path"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
)

type OrganizationController struct {
	otf.OrganizationService

	// UserService provides access to current user and their session
	otf.UserService

	renderer

	*templateDataFactory
}

func (c *OrganizationController) addRoutes(router *mux.Router) {
	router.HandleFunc("/", c.List).Methods("GET").Name("listOrganization")
	router.HandleFunc("/new", c.New).Methods("GET").Name("newOrganization")
	router.HandleFunc("/create", c.Create).Methods("POST").Name("createOrganization")
	router.HandleFunc("/{organization_name}", c.Get).Methods("GET").Name("getOrganization")
	router.HandleFunc("/{organization_name}/overview", c.Get).Methods("GET").Name("getOrganizationOverview")
	router.HandleFunc("/{organization_name}/edit", c.Edit).Methods("GET").Name("editOrganization")
	router.HandleFunc("/{organization_name}/update", c.Update).Methods("POST").Name("updateOrganization")
	router.HandleFunc("/{organization_name}/delete", c.Delete).Methods("POST").Name("deleteOrganization")
}

func (c *OrganizationController) List(w http.ResponseWriter, r *http.Request) {
	var opts otf.OrganizationListOptions

	// populate options struct from query and route paramters
	if err := decodeAll(r, &opts); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
	}

	workspaces, err := c.OrganizationService.List(r.Context(), opts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tdata := c.newTemplateData(r, struct {
		List    *otf.OrganizationList
		Options otf.OrganizationListOptions
	}{
		List:    workspaces,
		Options: opts,
	})

	if err := c.renderTemplate("organization_list.tmpl", w, tdata); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c *OrganizationController) New(w http.ResponseWriter, r *http.Request) {
	tdata := c.newTemplateData(r, nil)

	if err := c.renderTemplate("organization_new.tmpl", w, tdata); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c *OrganizationController) Create(w http.ResponseWriter, r *http.Request) {
	var opts otf.OrganizationCreateOptions
	if err := decodeAll(r, &opts); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
	}

	workspace, err := c.OrganizationService.Create(r.Context(), opts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, path.Join("..", workspace.Name), http.StatusFound)
}

func (c *OrganizationController) Get(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "./overview", http.StatusFound)
}

func (c *OrganizationController) Overview(w http.ResponseWriter, r *http.Request) {
	org, err := c.OrganizationService.Get(r.Context(), mux.Vars(r)["organization_name"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tdata := c.newTemplateData(r, org)

	if err := c.renderTemplate("organization_get.tmpl", w, tdata); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	http.Redirect(w, r, "./overview", http.StatusFound)
}

func (c *OrganizationController) Edit(w http.ResponseWriter, r *http.Request) {
	organization, err := c.OrganizationService.Get(r.Context(), mux.Vars(r)["organization_name"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tdata := c.newTemplateData(r, organization)

	if err := c.renderTemplate("organization_edit.tmpl", w, tdata); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c *OrganizationController) Update(w http.ResponseWriter, r *http.Request) {
	var opts otf.OrganizationUpdateOptions
	if err := decodeAll(r, &opts); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
	}

	_, err := c.OrganizationService.Update(r.Context(), mux.Vars(r)["organization_name"], &opts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "../edit", http.StatusFound)
}

func (c *OrganizationController) Delete(w http.ResponseWriter, r *http.Request) {
	err := c.OrganizationService.Delete(r.Context(), mux.Vars(r)["organization_name"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "../../", http.StatusFound)
}

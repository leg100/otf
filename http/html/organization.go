package html

import (
	"net/http"
	"path"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
)

type OrganizationController struct {
	otf.OrganizationService

	// UserService provides access to current user and their session
	otf.UserService

	renderer

	*router

	// for setting flash messages
	sessions *sessions

	*templateDataFactory
}

func (c *OrganizationController) addRoutes(router *mux.Router) {
	router.HandleFunc("/", c.List).Methods("GET").Name("listOrganization")
	router.HandleFunc("/new", c.New).Methods("GET").Name("newOrganization")
	router.HandleFunc("/create", c.Create).Methods("POST").Name("createOrganization")
	router.HandleFunc("/{organization_name}", c.Get).Methods("GET").Name("getOrganization")
	router.HandleFunc("/{organization_name}/overview", c.Overview).Methods("GET").Name("getOrganizationOverview")
	router.HandleFunc("/{organization_name}/edit", c.Edit).Methods("GET").Name("editOrganization")
	router.HandleFunc("/{organization_name}/update", c.Update).Methods("POST").Name("updateOrganization")
	router.HandleFunc("/{organization_name}/delete", c.Delete).Methods("POST").Name("deleteOrganization")
}

func (c *OrganizationController) List(w http.ResponseWriter, r *http.Request) {
	var opts otf.OrganizationListOptions

	// populate options struct from query and route paramters
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	workspaces, err := c.OrganizationService.List(r.Context(), opts)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
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
		writeError(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c *OrganizationController) New(w http.ResponseWriter, r *http.Request) {
	tdata := c.newTemplateData(r, nil)

	if err := c.renderTemplate("organization_new.tmpl", w, tdata); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (c *OrganizationController) Create(w http.ResponseWriter, r *http.Request) {
	var opts otf.OrganizationCreateOptions
	if err := decode.Form(&opts, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	organization, err := c.OrganizationService.Create(r.Context(), opts)
	if err == otf.ErrResourcesAlreadyExists {
		c.sessions.FlashError(r, "organization already exists: ", *opts.Name)
		http.Redirect(w, r, c.route("newOrganization"), http.StatusFound)
		return
	}
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	c.sessions.FlashSuccess(r, "created organization: ", organization.Name())

	http.Redirect(w, r, c.relative(r, "getOrganization", "organization_name", *opts.Name), http.StatusFound)
}

func (c *OrganizationController) Get(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, path.Join(r.URL.Path, "overview"), http.StatusFound)
}

func (c *OrganizationController) Overview(w http.ResponseWriter, r *http.Request) {
	org, err := c.OrganizationService.Get(r.Context(), mux.Vars(r)["organization_name"])
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tdata := c.newTemplateData(r, org)

	if err := c.renderTemplate("organization_get.tmpl", w, tdata); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c *OrganizationController) Edit(w http.ResponseWriter, r *http.Request) {
	organization, err := c.OrganizationService.Get(r.Context(), mux.Vars(r)["organization_name"])
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tdata := c.newTemplateData(r, organization)

	if err := c.renderTemplate("organization_edit.tmpl", w, tdata); err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c *OrganizationController) Update(w http.ResponseWriter, r *http.Request) {
	var opts otf.OrganizationUpdateOptions
	if err := decode.Form(&opts, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	organization, err := c.OrganizationService.Update(r.Context(), mux.Vars(r)["organization_name"], &opts)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	c.sessions.FlashSuccess(r, "updated organization")

	// Explicitly specify route variable for organization name because the user
	// might have updated it.
	http.Redirect(w, r, c.route("editOrganization", "organization_name", organization.Name()), http.StatusFound)
}

func (c *OrganizationController) Delete(w http.ResponseWriter, r *http.Request) {
	organizationName := mux.Vars(r)["organization_name"]

	err := c.OrganizationService.Delete(r.Context(), organizationName)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	c.sessions.FlashSuccess(r, "deleted organization: ", organizationName)
	http.Redirect(w, r, c.route("listOrganization"), http.StatusFound)
}

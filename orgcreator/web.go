package orgcreator

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/http/html/paths"
)

// web is the web application for organizations
type web struct {
	html.Renderer

	svc Service
}

func (a *web) addHandlers(r *mux.Router) {
	r = html.UIRouter(r)

	r.HandleFunc("/organizations/new", a.new).Methods("GET")
	r.HandleFunc("/organizations/create", a.create).Methods("POST")
}

func (a *web) new(w http.ResponseWriter, r *http.Request) {
	a.Render("organization_new.tmpl", w, html.NewSitePage(r, "new organization"))
}

func (a *web) create(w http.ResponseWriter, r *http.Request) {
	var opts OrganizationCreateOptions
	if err := decode.Form(&opts, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	org, err := a.svc.CreateOrganization(r.Context(), opts)
	if err == otf.ErrResourceAlreadyExists {
		html.FlashError(w, "organization already exists: "+*opts.Name)
		http.Redirect(w, r, paths.NewOrganization(), http.StatusFound)
		return
	}
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "created organization: "+org.Name)
	http.Redirect(w, r, paths.Organization(org.Name), http.StatusFound)
}

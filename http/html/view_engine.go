package html

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
)

// viewEngine is responsible for populating and rendering views
type viewEngine struct {
	renderer // render templates
}

func NewViewEngine(devmode bool) (*viewEngine, error) {
	renderer, err := newRenderer(devmode)
	if err != nil {
		return nil, err
	}
	return &viewEngine{
		renderer: renderer,
	}, nil
}

// Render the view using the template. Note this should be the last thing called
// in a handler because it writes an HTTP5xx to the response if there is an
// error.
func (ve *viewEngine) Render(name string, w http.ResponseWriter, r *http.Request, content interface{}) {
	flashes, err := PopFlashes(w, r)
	if err != nil {
		htmlPanic("reading flash messages: %v", err)
	}

	err = ve.RenderTemplate(name, w, &view{
		Content: content,
		Version: otf.Version,
		request: r,
		Flashes: flashes,
	})
	if err != nil {
		Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// view provides data and methods to a template
type view struct {
	Content interface{} // arbitary data made available to the template
	Version string      // otf version string in footer

	request *http.Request // info regarding current request
	Flashes []flash       // flash messages to render in template
}

func (v *view) CurrentUser() otf.Subject {
	subject, err := otf.SubjectFromContext(v.request.Context())
	if err != nil {
		return nil
	}
	return subject
}

// CurrentOrganization retrieves the user's current organization
func (v *view) CurrentOrganization() *string {
	name, err := organizationFromContext(v.request.Context())
	if err != nil {
		return nil
	}
	return &name
}

func (v *view) CurrentPath() string {
	return v.request.URL.Path
}

func (v *view) CurrentURL() string {
	return v.request.URL.String()
}

// IsOrganizationRoute determines if the current request is for a route that
// contains the current organization name, or the list of organizations.
func (v *view) IsOrganizationRoute() bool {
	if mux.CurrentRoute(v.request).GetName() == "listOrganization" {
		return true
	}
	_, ok := mux.Vars(v.request)["organization_name"]
	return ok
}

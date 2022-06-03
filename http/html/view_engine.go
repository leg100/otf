package html

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
)

// viewEngine is responsible for populating and rendering views
type viewEngine struct {
	// render templates
	renderer
}

func newViewEngine(dev bool) (*viewEngine, error) {
	renderer, err := newRenderer(dev)
	if err != nil {
		return nil, err
	}
	return &viewEngine{
		renderer: renderer,
	}, nil
}

// render the view using the template. Note this should be the last thing called
// in a handler because it writes an HTTP5xx to the response if there is an
// error.
func (ve *viewEngine) render(name string, w http.ResponseWriter, r *http.Request, content interface{}) {
	err := ve.renderTemplate(name, w, &view{
		Content:     content,
		flashPopper: popFlashFunc(w, r),
		request:     r,
		Version:     otf.Version,
	})
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
	}
}

// view provides data and methods to a template
type view struct {
	// arbitary data made available to the template
	Content interface{}
	// info regarding current request
	request *http.Request
	// pop flash message in template
	flashPopper func() *flash
	// oTF version string in footer
	Version string
}

func (v *view) PopFlash() *flash {
	return v.flashPopper()
}

func (v *view) CurrentUser() (*otf.User, error) {
	return userFromContext(v.request.Context())
}

func (v *view) CurrentSession() (*otf.Session, error) {
	return sessionFromContext(v.request.Context())
}

func (v *view) CurrentOrganization() (string, error) {
	return organizationFromContext(v.request.Context())
}

func (v *view) CurrentPath() string {
	return v.request.URL.Path
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

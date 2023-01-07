package html

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
)

// viewEngine is responsible for populating and rendering views
type viewEngine struct {
	renderer // render templates
	hostname string
}

type viewEngineOptions struct {
	devMode  bool
	hostname string
}

func newViewEngine(opts viewEngineOptions) (*viewEngine, error) {
	renderer, err := newRenderer(opts.devMode)
	if err != nil {
		return nil, err
	}
	return &viewEngine{
		renderer: renderer,
		hostname: opts.hostname,
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
		Hostname:    ve.hostname,
	})
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
	}
}

// view provides data and methods to a template
type view struct {
	Content     interface{}   // arbitary data made available to the template
	request     *http.Request // info regarding current request
	flashPopper func() *flash // pop flash message in template
	Version     string        // otf version string in footer
	Hostname    string        // user-facing hostname
}

func (v *view) PopFlash() *flash {
	return v.flashPopper()
}

func (v *view) CurrentUser() *otf.User {
	user, err := otf.UserFromContext(v.request.Context())
	if err != nil {
		return nil
	}
	return user
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

package html

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
)

// viewEngine is responsible for populating and rendering views
type viewEngine struct {
	// views look up routes for links
	router *router
	// render templates
	renderer
}

func newViewEngine(router *router, dev bool) (*viewEngine, error) {
	renderer, err := newRenderer(dev)
	if err != nil {
		return nil, err
	}

	return &viewEngine{
		router:   router,
		renderer: renderer,
	}, nil
}

func (ve *viewEngine) render(name string, w http.ResponseWriter, r *http.Request, content interface{}) {
	err := ve.renderTemplate(name, w, &view{
		Content:     content,
		router:      ve.router,
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
	// make routes available for producing links in template
	router *router
	// info regarding current request
	request *http.Request
	// pop flash message in template
	flashPopper func() *flash
	// oTF version string in footer
	Version string
}

// Path proxies access to the router's route method
func (v *view) Path(name string, pairs ...string) string {
	return v.router.route(name, pairs...)
}

// Relative proxies access to the router's relative method
func (v *view) Relative(name string, pairs ...string) string {
	return v.router.relative(v.request, name, pairs...)
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

func (v *view) RouteVars() map[string]string {
	return mux.Vars(v.request)
}

func (v *view) PopFlash() *flash {
	return v.flashPopper()
}

func (v *view) CurrentUser() *otf.User {
	user, err := getCtxUser(v.request.Context())
	if err != nil {
		return nil
	}
	return user
}

func (v *view) CurrentSession() *otf.Session {
	session, err := getCtxSession(v.request.Context())
	if err != nil {
		return nil
	}
	return session
}

func (v *view) CurrentPath() string {
	return v.request.URL.Path
}

func flattenMap(m map[string]string) (s []string) {
	for k, v := range m {
		s = append(s, k, v)
	}
	return
}

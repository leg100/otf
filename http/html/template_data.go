package html

import (
	"fmt"
	"net/http"
	"path"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
)

// templateDataFactory produces templateData structs
type templateDataFactory struct {
	// provide access to routes
	router *mux.Router
}

func (f *templateDataFactory) newTemplateData(w http.ResponseWriter, r *http.Request, content interface{}) templateData {
	return templateData{
		Content:     content,
		router:      &router{f.router},
		flashPopper: popFlashFunc(w, r),
		request:     r,
		Version:     otf.Version,
	}
}

type templateData struct {
	// Content is specific to the content being embedded within the layout.
	Content interface{}

	router *router

	request *http.Request

	flashPopper func() *flash

	// oTF version string for showing in footer
	Version string
}

// Path proxies access to the router's route method
func (td *templateData) Path(name string, pairs ...string) string {
	return td.router.route(name, pairs...)
}

// Relative proxies access to the router's relative method
func (td *templateData) Relative(name string, pairs ...string) string {
	return td.router.relative(td.request, name, pairs...)
}

// IsOrganizationRoute determines if the current request is for a route that
// contains the current organization name, or the list of organizations.
func (td *templateData) IsOrganizationRoute() bool {
	if mux.CurrentRoute(td.request).GetName() == "listOrganization" {
		return true
	}

	_, ok := mux.Vars(td.request)["organization_name"]
	return ok
}

func (td *templateData) Breadcrumbs() (crumbs []Anchor, err error) {
	route := mux.CurrentRoute(td.request)

	crumbs, err = td.makeBreadcrumbs(route, crumbs)
	if err != nil {
		return nil, err
	}

	return crumbs, nil
}

func (td *templateData) RouteVars() map[string]string {
	return mux.Vars(td.request)
}

func (td *templateData) PopFlash() *flash {
	return td.flashPopper()
}

func (td *templateData) CurrentUser() *otf.User {
	user, err := getCtxUser(td.request.Context())
	if err != nil {
		return nil
	}
	return user
}

func (td *templateData) CurrentSession() *otf.Session {
	session, err := getCtxSession(td.request.Context())
	if err != nil {
		return nil
	}
	return session
}

func (td *templateData) CurrentPath() string {
	return td.request.URL.Path
}

func (td *templateData) makeBreadcrumbs(route *mux.Route, crumbs []Anchor) ([]Anchor, error) {
	link, err := route.URLPath(flattenMap(mux.Vars(td.request))...)
	if err != nil {
		return nil, err
	}
	name := path.Base(link.Path)

	// place parent crumb in front
	crumbs = append([]Anchor{{Name: name, Link: link.Path}}, crumbs...)

	parent, ok := parentLookupTable[route.GetName()]
	if !ok {
		return crumbs, nil
	}

	parentRoute := td.router.Get(parent)
	if parentRoute == nil {
		return nil, fmt.Errorf("no such web route exists: %s", parent)
	}

	return td.makeBreadcrumbs(parentRoute, crumbs)
}

func flattenMap(m map[string]string) (s []string) {
	for k, v := range m {
		s = append(s, k, v)
	}
	return
}

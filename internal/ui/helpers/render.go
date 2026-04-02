package helpers

import (
	"context"
	"io"
	"net/http"

	"github.com/a-h/templ"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/workspace"
)

// RenderPage renders a component within a layout, e.g. header, footer, menus
// etc.
func RenderPage(c templ.Component, title string, w http.ResponseWriter, r *http.Request, opts ...RenderPageOption) {
	props := LayoutProps{
		Title: title,
	}
	for _, o := range opts {
		o(&props)
	}
	// Render component as a child of the layout component.
	layout := Layout(props)
	Render(layout, w, r, WithChildren(c))
}

// Render a template. Wraps the upstream templ handler to carry out additional
// actions every time a template is rendered.
func Render(c templ.Component, w http.ResponseWriter, r *http.Request, renderOptions ...RenderOption) {
	// add request to context for templates to access
	ctx := context.WithValue(r.Context(), requestKey{}, r)
	// add response to context for templates to access
	ctx = context.WithValue(ctx, responseKey{}, w)
	// apply rendering options to context
	for _, o := range renderOptions {
		ctx = o(ctx)
	}
	// handle errors
	opts := []func(*templ.ComponentHandler){
		templ.WithErrorHandler(func(r *http.Request, err error) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				Error(r, w, err.Error())
			})
		}),
	}
	// render only a fragment if ajax request.
	if r.Header.Get("HX-Target") != "" {
		opts = append(opts, templ.WithFragments(r.Header.Get("Hx-Target")))
	}
	templ.Handler(c, opts...).ServeHTTP(w, r.WithContext(ctx))
}

type RenderPageOption func(opts *LayoutProps)

func WithOrganization(org resource.ID) RenderPageOption {
	return func(opts *LayoutProps) {
		opts.Organization = org
	}
}

func WithWorkspace(ws *workspace.Workspace, authorizer authz.Interface) RenderPageOption {
	return func(opts *LayoutProps) {
		opts.Organization = new(ws.Organization)
		opts.Workspace = new(ws.Info())
		opts.Authorizer = authorizer
	}
}

func WithSideMenu(comp templ.Component) RenderPageOption {
	return func(opts *LayoutProps) {
		opts.SideMenu = comp
	}
}

func WithBreadcrumbs(crumbs ...Breadcrumb) RenderPageOption {
	return func(opts *LayoutProps) {
		opts.Breadcrumbs = append(opts.Breadcrumbs, crumbs...)
	}
}

func WithContentActions(comp templ.Component) RenderPageOption {
	return func(opts *LayoutProps) {
		opts.ContentActions = comp
	}
}

func WithPreContent(comp templ.Component) RenderPageOption {
	return func(opts *LayoutProps) {
		opts.PreContent = comp
	}
}

func WithPostContent(comp templ.Component) RenderPageOption {
	return func(opts *LayoutProps) {
		opts.PostContent = comp
	}
}

type RenderOption func(ctx context.Context) context.Context

func WithChildren(comp templ.Component) RenderOption {
	return func(ctx context.Context) context.Context {
		return templ.WithChildren(ctx, comp)
	}
}

func RenderSnippet(c templ.Component, w io.Writer, r *http.Request) error {
	// add request to context for templates to access
	ctx := context.WithValue(r.Context(), requestKey{}, r)
	// TODO: is it ok to omit response from context?
	return c.Render(ctx, w)
}

type requestKey struct{}

func RequestFromContext(ctx context.Context) *http.Request {
	if r, ok := ctx.Value(requestKey{}).(*http.Request); ok {
		return r
	}
	return nil
}

type responseKey struct{}

func ResponseFromContext(ctx context.Context) http.ResponseWriter {
	if w, ok := ctx.Value(responseKey{}).(http.ResponseWriter); ok {
		return w
	}
	return nil
}

package helpers

import (
	"bytes"
	"context"
	"html/template"
	"net/http"
	"net/url"
	"time"

	"github.com/gomarkdown/markdown"
	"github.com/leg100/otf/internal/authz"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/ui/paths"
)

func AssetPath(ctx context.Context, path string) (string, error) {
	return otfhttp.AssetsFS.Path(path)
}

func CurrentUsername(ctx context.Context) (string, error) {
	subject, err := authz.SubjectFromContext(ctx)
	if err != nil {
		return "", err
	}
	return subject.String(), nil
}

func Authenticated(ctx context.Context) bool {
	if _, err := authz.SubjectFromContext(ctx); err != nil {
		return false
	}
	return true
}

func IsOwner(ctx context.Context, organization resource.ID) bool {
	subject, err := authz.SubjectFromContext(ctx)
	if err != nil {
		return false
	}
	if user, ok := subject.(interface {
		IsOwner(resource.ID) bool
	}); ok {
		return user.IsOwner(organization)
	}
	return false
}

func IsSiteAdmin(ctx context.Context) bool {
	subject, err := authz.SubjectFromContext(ctx)
	if err != nil {
		return false
	}
	if user, ok := subject.(interface {
		IsSiteAdmin() bool
	}); ok {
		return user.IsSiteAdmin()
	}
	return false
}

func CurrentPath(ctx context.Context) string {
	request := RequestFromContext(ctx)
	if request == nil {
		return ""
	}
	return request.URL.Path
}

func CurrentURL(ctx context.Context) string {
	request := RequestFromContext(ctx)
	if request == nil {
		return ""
	}
	return request.URL.String()
}

func CurrentURLWithoutQuery(ctx context.Context) string {
	request := RequestFromContext(ctx)
	if request == nil {
		return ""
	}
	// Make a copy of URL to avoid mutating the original.
	u := new(url.URL)
	*u = *request.URL

	u.RawQuery = ""
	return u.String()
}

// TokenFlashMessage is a helper for rendering a flash message with an
// authentication token.
func TokenFlashMessage(w http.ResponseWriter, token []byte) error {
	// render a small templated flash message
	buf := new(bytes.Buffer)
	if err := flashToken(string(token)).Render(context.Background(), buf); err != nil {
		return err
	}
	FlashSuccess(w, buf.String())
	return nil
}

func Cookie(ctx context.Context, name string) string {
	request := RequestFromContext(ctx)
	if request == nil {
		return ""
	}
	for _, cookie := range request.Cookies() {
		if cookie.Name == name {
			return cookie.Value
		}
	}
	return ""
}

const (
	pathCookie = "path"
)

func MarkdownToHTML(md []byte) template.HTML {
	return template.HTML(string(markdown.ToHTML(md, nil, nil)))
}

// SendUserToLoginPage sends user to the login prompt page, saving the original
// path they tried to access so it can return them there after login.
func SendUserToLoginPage(w http.ResponseWriter, r *http.Request) {
	// if request path was for a background event-stream then save the referring
	// html page, otherwise the user will be returned to a blank page.
	path := r.URL.String()
	if r.Header.Get("Accept") == "text/event-stream" {
		path = r.Referer()
	}
	SetCookie(w, pathCookie, path, nil)

	// Force ajax requests to reload entire page
	if isHTMX := r.Header.Get("HX-Request"); isHTMX == "true" {
		w.Header().Add("HX-Refresh", "true")
		return
	}
	http.Redirect(w, r, paths.Login(), http.StatusFound)
}

// ReturnUserOriginalPage returns a user to the original page they tried to
// access before they were redirected to the login page.
func ReturnUserOriginalPage(w http.ResponseWriter, r *http.Request) {
	// Return user to the original path they attempted to access
	if cookie, err := r.Cookie(pathCookie); err == nil {
		SetCookie(w, pathCookie, "", &time.Time{})
		http.Redirect(w, r, cookie.Value, http.StatusFound)
	} else {
		http.Redirect(w, r, paths.Profile(), http.StatusFound)
	}
}

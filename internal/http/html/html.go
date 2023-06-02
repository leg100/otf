// Package html contains code relating specifically to the web UI.
package html

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/gomarkdown/markdown"
	"github.com/leg100/otf/internal/http/html/paths"
)

const (
	pathCookie = "path"
)

func MarkdownToHTML(md []byte) template.HTML {
	return template.HTML(string(markdown.ToHTML(md, nil, nil)))
}

// htmlPanic raises a panic - an upstream middleware handler should catch the panic
// and sends an HTTP500 to the user.
func htmlPanic(format string, a ...any) {
	panic(fmt.Sprintf(format, a...))
}

// SendUserToLoginPage sends user to the login prompt page, saving the original
// page they tried to access so it can return them there after login.
func SendUserToLoginPage(w http.ResponseWriter, r *http.Request) {
	SetCookie(w, pathCookie, r.URL.String(), nil)
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

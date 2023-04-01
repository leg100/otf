// Package html contains code relating specifically to the web UI.
package html

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"reflect"
	"time"

	"github.com/gomarkdown/markdown"
	"github.com/leg100/otf/http/html/paths"
)

const (
	pathCookie = "path"
)

// mergeQuery merges the query string into the given url, replacing any existing
// query parameters with the same name.
func mergeQuery(u string, q string) (string, error) {
	parsedURL, err := url.Parse(u)
	if err != nil {
		return "", err
	}
	mergeQuery, err := url.ParseQuery(q)
	if err != nil {
		return "", err
	}
	existingQuery := parsedURL.Query()
	for k, v := range mergeQuery {
		existingQuery.Set(k, v[0])
	}
	parsedURL.RawQuery = existingQuery.Encode()
	return parsedURL.String(), nil
}

func disabled(arg any, args ...any) (string, error) {
	return printIf("disabled", arg, args...)
}

func selected(arg any, args ...any) (string, error) {
	return printIf("selected", arg, args...)
}

func checked(arg any, args ...any) (string, error) {
	return printIf("checked", arg, args...)
}

// printIf prints a string if:
// (a) single arg provided, it is a boolean, and it is true.
// (b) multiple args provided, they are all strings, and they are all equal.
// otherwise it outputs an empty string
// This is useful for printing various strings in templates or not.
func printIf(s string, arg any, args ...any) (string, error) {
	if len(args) == 0 {
		if reflect.ValueOf(arg).Kind() == reflect.Bool {
			if reflect.ValueOf(arg).Bool() {
				return s, nil
			}
		}
		return "", nil
	}
	if reflect.ValueOf(arg).Kind() != reflect.String {
		return "", nil
	}
	lastarg := reflect.ValueOf(arg).String()
	for _, a := range args {
		if reflect.ValueOf(a).Kind() != reflect.String {
			return "", nil
		}
		if reflect.ValueOf(a).String() != lastarg {
			return "", nil
		}
	}
	return s, nil
}

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
	SetCookie(w, pathCookie, r.URL.Path, nil)
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

// Package html contains code relating specifically to the web UI.
package html

import (
	"html/template"
	"net/url"
	"reflect"

	"github.com/gomarkdown/markdown"
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

func markdownToHTML(md []byte) template.HTML {
	return template.HTML(string(markdown.ToHTML(md, nil, nil)))
}

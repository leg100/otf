/*
Package html provides the otf web app, serving up HTML formatted pages and associated assets (CSS, JS, etc).
*/
package html

import (
	"html/template"
	"net/url"

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

// selected returns the string "selected" if the given strings are equal -
// intended for use with the select options in an HTML template.
func selected(a, b string) string {
	if a == b {
		return "selected"
	}
	return ""
}

func markdownToHTML(md []byte) template.HTML {
	return template.HTML(string(markdown.ToHTML(md, nil, nil)))
}

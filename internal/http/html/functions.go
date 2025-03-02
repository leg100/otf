package html

import (
	"net/url"

	"github.com/a-h/templ"
)

// MergeQuery merges the query string into the given url, replacing any existing
// query parameters with the same name.
func MergeQuery(u string, q string) (templ.SafeURL, error) {
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
	return templ.SafeURL(parsedURL.String()), nil
}

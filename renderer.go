package otf

import "net/http"

// Renderer renders templated responses to http requests.
type Renderer interface {
	// Render the template with the given path to the response, populating the
	// template with content. The request can be used to provide further
	// request-related information to the template.
	Render(path string, w http.ResponseWriter, r *http.Request, content any)
}

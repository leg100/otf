package html

import (
	"net/http"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
)

// PageSize is the number of items in a paginated result to show on a single page
const PageSize int = 10

// SitePage contains data shared by all pages when rendering templates.
type SitePage struct {
	Title               string // page title
	Version             string // otf version string in footer
	CurrentOrganization string

	request *http.Request // current request
}

func NewSitePage(r *http.Request, title string) SitePage {
	return SitePage{
		Title:   title,
		Version: internal.Version,
		request: r,
	}
}

func (v SitePage) CurrentUser() string {
	subject, err := authz.SubjectFromContext(v.request.Context())
	if err != nil {
		return ""
	}
	return subject.String()
}

func (v SitePage) CurrentPath() string {
	return v.request.URL.Path
}

func (v SitePage) CurrentURL() string {
	return v.request.URL.String()
}

func (v SitePage) Flashes() ([]flash, error) {
	return PopFlashes(v.request)
}

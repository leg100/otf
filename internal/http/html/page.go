package html

import (
	"net/http"

	internal "github.com/leg100/otf"
)

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

func (v SitePage) CurrentUser() internal.Subject {
	subject, err := internal.SubjectFromContext(v.request.Context())
	if err != nil {
		return nil
	}
	return subject
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

package assets

import (
	"html/template"
	"net/http"
)

const (
	// Paths to static assets in relation to the package directory
	layoutTemplatePath  = "static/templates/layout.tmpl"
	contentTemplatesDir = "static/templates/content"
	stylesheetDir       = "static/css"
)

// Server provides the means to retrieve http assets (templates and static files
// such as CSS).
type Server interface {
	GetTemplate(string) *template.Template
	GetStaticFS() http.FileSystem
	Links() []string
}

type LayoutTemplateOptions struct {
	Title         string
	Stylesheets   []string
	FlashMessages []template.HTML
}

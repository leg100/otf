package assets

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
)

var (
	//go:embed static
	embedded embed.FS
)

// EmbeddedServer provides access to assets embedded in the go binary.
type EmbeddedServer struct {
	// templates maps template names to parsed contents
	templates map[string]*template.Template

	// filesystem containing static assets
	static *StaticFS

	// relative paths to stylesheets for use in <link...> tags
	links []string
}

func NewEmbeddedServer() (*EmbeddedServer, error) {
	cache, err := newTemplateCache(embedded, contentTemplatesGlob, layoutTemplatePath)
	if err != nil {
		return nil, err
	}

	server := EmbeddedServer{
		templates: cache,
		static:    NewStaticFS(embedded),
	}

	server.links, err = cacheBustingPaths(embedded, filepath.Join(stylesheetDir, "*.css"))
	if err != nil {
		return nil, fmt.Errorf("unable to read embedded stylesheets directory: %w", err)
	}

	return &server, nil
}

func (s *EmbeddedServer) GetTemplate(name string) *template.Template {
	return s.templates[name]
}

func (s *EmbeddedServer) GetStaticFS() http.FileSystem {
	return http.FS(s.static)
}

func (s *EmbeddedServer) Links() []string {
	return s.links
}

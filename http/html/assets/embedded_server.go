package assets

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"

	"github.com/Masterminds/sprig"
)

var (
	//go:embed static
	embedded embed.FS
)

// EmbeddedServer provides access to assets embedded in the go binary.
type EmbeddedServer struct {
	// templates maps template names to parsed contents
	templates map[string]*template.Template

	// The embedded filesystem
	filesystem embed.FS

	// relative paths to stylesheets for use in <link...> tags
	links []string
}

func NewEmbeddedServer() (*EmbeddedServer, error) {
	pattern := fmt.Sprintf("%s/*.tmpl", contentTemplatesDir)

	paths, err := fs.Glob(embedded, pattern)
	if err != nil {
		return nil, fmt.Errorf("unable to read embedded templates directory: %w", err)
	}

	server := EmbeddedServer{
		templates:  make(map[string]*template.Template, len(paths)),
		filesystem: embedded,
	}

	name := filepath.Base(layoutTemplatePath)
	for _, p := range paths {
		template, err := template.New(name).Funcs(sprig.FuncMap()).ParseFS(embedded, layoutTemplatePath, p)
		if err != nil {
			return nil, fmt.Errorf("unable to parse embedded template: %w", err)
		}

		server.templates[filepath.Base(p)] = template
	}

	server.links, err = CacheBustingPaths(embedded, filepath.Join(stylesheetDir, "*.css"))
	if err != nil {
		return nil, fmt.Errorf("unable to read embedded stylesheets directory: %w", err)
	}

	return &server, nil
}

func (s *EmbeddedServer) GetTemplate(name string) *template.Template {
	return s.templates[name]
}

func (s *EmbeddedServer) GetStaticFS() http.FileSystem {
	return http.FS(s.filesystem)
}

func (s *EmbeddedServer) Links() []string {
	return s.links
}

package assets

import (
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Masterminds/sprig"
)

const (
	// AssetsDir is the relative path in the source git repository for this
	// package.
	AssetsDir = "http/assets"
)

// DevServer reads assets from developer's machine, permitting use of
// something like livereload to see changes in real-time.
//
// !NOT FOR PRODUCTION USE!
type DevServer struct {
	// path to directory containing assets
	root string
}

func NewDevServer() *DevServer {
	server := DevServer{
		root: AssetsDir,
	}

	return &server
}

// RenderTemplate embeds the named content template within the layout template
// and then renders the combined template using the data, writing it out to w.
func (s *DevServer) RenderTemplate(name string, w io.Writer, data interface{}) error {
	layout := filepath.Join(s.root, layoutTemplatePath)
	content := filepath.Join(s.root, contentTemplatesDir, name)

	return template.Must(
		template.New(filepath.Base(layout)).Funcs(sprig.FuncMap()).ParseFiles(layout, content),
	).Execute(w, data)
}

func (s *DevServer) GetStaticFS() http.FileSystem {
	return http.Dir(s.root)
}

func (s *DevServer) LayoutOptions(title string) *LayoutTemplateOptions {
	return &LayoutTemplateOptions{
		Title:       title,
		Stylesheets: s.links(),
	}
}

func (s *DevServer) links() []string {
	links, err := CacheBustingPaths(os.DirFS(s.root), filepath.Join(stylesheetDir, "*.css"))
	if err != nil {
		panic("unable to read embedded stylesheets directory: " + err.Error())
	}

	return links
}

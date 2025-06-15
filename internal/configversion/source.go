package configversion

import (
	"sync"

	"github.com/a-h/templ"
)

var (
	DefaultSource          = SourceAPI
	SourceAPI       Source = "tfe-api"
	SourceUI        Source = "tfe-ui"
	SourceTerraform Source = "terraform+cloud"
)

// Source is the source or origin of the configuration
type Source string

func Ptr(s Source) *Source { return &s }

// sourceIconDB is a database of sources and their icons
type sourceIconDB struct {
	mu    sync.Mutex
	icons map[Source]templ.Component
}

func newSourceIconDB() *sourceIconDB {
	return &sourceIconDB{
		icons: map[Source]templ.Component{
			SourceAPI:       SourceIconAPI(),
			SourceUI:        SourceIconUI(),
			SourceTerraform: SourceIconTerraform(),
		},
	}
}

func (db *sourceIconDB) RegisterSourceIcon(source Source, icon templ.Component) {
	db.mu.Lock()
	db.icons[source] = icon
	db.mu.Unlock()
}

func (db *sourceIconDB) GetSourceIcon(source Source) templ.Component {
	db.mu.Lock()
	defer db.mu.Unlock()

	icon, ok := db.icons[source]
	if !ok {
		// No icon for the source could be found, so just render nothing rather
		// than go to the bother of returning an error
		return templ.Raw("")
	}
	return icon
}

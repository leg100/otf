package source

import (
	"sync"

	"github.com/a-h/templ"
)

var (
	Default          = API
	API       Source = "tfe-api"
	UI        Source = "tfe-ui"
	Terraform Source = "terraform+cloud"
)

// Source is the source or origin of the configuration
type Source string

// IconDB is a database of sources and their icons
type IconDB struct {
	mu    sync.Mutex
	icons map[Source]templ.Component
}

func NewIconDB() *IconDB {
	return &IconDB{
		icons: map[Source]templ.Component{
			API:       IconAPI(),
			UI:        IconUI(),
			Terraform: IconTerraform(),
		},
	}
}

func (db *IconDB) RegisterSourceIcon(source Source, icon templ.Component) {
	db.mu.Lock()
	db.icons[source] = icon
	db.mu.Unlock()
}

func (db *IconDB) GetSourceIcon(source Source) templ.Component {
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

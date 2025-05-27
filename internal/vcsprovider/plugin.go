package vcsprovider

import (
	"net/http"

	"github.com/leg100/otf/internal/vcs"
)

var (
	kinds               map[vcs.Kind]Plugin
	tfeServiceProviders map[TFEServiceProviderType]Plugin
)

func RegisterKind(kind vcs.Kind, plugin Plugin) {
	kinds[kind] = plugin
}

func RegisterTFEServiceProvider(service TFEServiceProviderType, plugin Plugin) {
	tfeServiceProviders[service] = plugin
}

type Plugin interface {
	NewClient(vcsProvider *VCSProvider, transport http.Transport) (vcs.Client, error)
	NewHandler(w http.ResponseWriter, r *http.Request)
	EditHandler(w http.ResponseWriter, r *http.Request)
}

package vcsprovider

import (
	"net/http"
)

type Plugin interface {
	// NewClient constructs a client for interacting with the upstream vcs provider
	NewClient(provider *VCSProvider)
	// NewHandler renders a web page for creating a new vcs provider
	NewHandler(w http.ResponseWriter, r *http.Request)
	// CreateHandler creates a vcs provider
	CreateHandler(w http.ResponseWriter, r *http.Request)
	// EditHandler renders a web page for updating an existing vcs provider
	EditHandler(w http.ResponseWriter, r *http.Request)
	// UpdateHandler updates an existing vcs provider
	UpdateHandler(w http.ResponseWriter, r *http.Request)
}

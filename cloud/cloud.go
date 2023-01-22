// Package cloud provides types for use with cloud providers.
package cloud

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// Cloud is an external provider of various cloud services e.g. identity provider, VCS
// repositories etc.
type Cloud interface {
	NewClient(context.Context, ClientOptions) (Client, error)
	EventHandler
}

type Service interface {
	GetCloudConfig(name string) (Config, error)
	ListCloudConfigs() []Config
}

// EventHandler handles incoming events
type EventHandler interface {
	// HandleEvent extracts a cloud-specific event from the http request, converting it into a
	// VCS event. Returns nil if the event is to be ignored.
	HandleEvent(w http.ResponseWriter, r *http.Request, opts HandleEventOptions) VCSEvent
}

type HandleEventOptions struct {
	Secret    string
	WebhookID uuid.UUID
}

// Repo is a VCS repository belonging to a cloud
//
type Repo struct {
	Identifier string `schema:"identifier,required"` // <repo_owner>/<repo_name>
	Branch     string `schema:"branch,required"`     // default branch
}

func (r Repo) ID() string { return r.Identifier }

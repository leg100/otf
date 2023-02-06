package otf

import (
	"context"
	"time"

	"github.com/gorilla/mux"
)

type RegistrySession interface {
	Token() string
	Organization() string
	Expiry() time.Time

	Subject
}

type RegistrySessionService interface {
	// AddHandlers adds handlers for the http api.
	AddHandlers(*mux.Router)

	RegistrySessionApp
}

type RegistrySessionApp interface {
	CreateRegistrySession(ctx context.Context, organization string) (RegistrySession, error)
	GetRegistrySession(ctx context.Context, token string) (RegistrySession, error)
}

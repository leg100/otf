package vcs

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"sync"

	"github.com/a-h/templ"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/configversion/source"
	"golang.org/x/exp/maps"
)

// KindID of vcs hosting provider
type KindID string

// Kind is a kind of vcs provider. Each kind represents a particular VCS hosting
// provider (e.g. github), and a way of interacting with the provider, including
// authentication, event handling. Typically there is one kind per VCS hosting
// provider, but providers sometimes offer more than one of interacting with it,
// e.g. GitHub uses both personal access tokens and a GitHub 'app' which is
// 'installed' via a private key.
type Kind struct {
	// ID distinguishes this kind from other kinds. NOTE: This must have first
	// been inserted into the vcs_kinds table via a database migration.
	ID KindID
	// DefaultURL is the default base URL for the provider.
	DefaultURL *internal.WebURL
	// Icon renders an icon identifying the VCS host kind.
	Icon templ.Component
	// TokenKind provides info about the token the provider expects. Mutually
	// exclusive with AppKind.
	TokenKind *TokenKind
	// AppKind provides info about installations for this ProviderKind.
	// Mutually exclusive with TokenKind.
	AppKind AppKind
	// NewClient constructs a client implementation.
	NewClient func(context.Context, ClientConfig) (Client, error)
	// EventHandler handles incoming events from the VCS host before relaying
	// them onwards for triggering actions like creating runs etc.
	EventHandler func(r *http.Request, secret string) (*EventPayload, error)
	// SkipRepohook if true skips the creation of a repository-level webhook.
	SkipRepohook bool
	// TFEServiceProviders registers the kind with the TFE API, permitting vcs
	// providers of this kind to be created via the TFE API. The value specified
	// here is the value that should be be provided by API clients via the
	// "service-provider" attribute:
	//
	// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/oauth-clients#request-body
	//
	// Must provide at least one type. A kind can register more than one type,
	// e.g. gitlab-ce and gitlab-ee, in which case if a client specifies either
	// of these types then OTF creates a provider of the gitlab kind. Note that
	// OTF makes no distinction between multiple types but it's merely there to
	// support the TFE API and the various values that clients might provide
	// according to the API documentation (though OTF may in future make a
	// distinction!).
	TFEServiceProviders []TFEServiceProviderType
	// Source sets the source identifier for this kind, to inform users which
	// kind is the source of a run configuration. By default the ID is used as
	// the source identifier but Source takes precedence if it is non-nil.
	Source *source.Source
}

func (k Kind) GetSource() source.Source {
	if k.Source != nil {
		return *k.Source
	}
	return source.Source(k.ID)
}

type TokenKind struct {
	// TokenDescription renders a helpful description of what is expected of the
	// token, e.g. what permissions it should possess.
	Description templ.Component
}

type AppKind interface {
	GetApp(context.Context) (App, error)
}

type App interface {
	// ListInstallations retrieves a list of installations.
	ListInstallations(context.Context) ([]Installation, error)
	// GetInstallation retrieves an installation by its ID.
	GetInstallation(context.Context, int64) (Installation, error)
	// InstallationLink is a link to the site where a user can create an
	// installation.
	InstallationLink() templ.SafeURL
}

type SourceIconRegistrar interface {
	RegisterSourceIcon(source source.Source, icon templ.Component)
}

// kindDB is a database of vcs provider kinds
type kindDB struct {
	mu            sync.Mutex
	kinds         map[KindID]Kind
	iconRegistrar SourceIconRegistrar
}

func newKindDB(registrar SourceIconRegistrar) *kindDB {
	return &kindDB{
		kinds:         make(map[KindID]Kind),
		iconRegistrar: registrar,
	}
}

func (db *kindDB) RegisterKind(kind Kind) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.kinds[kind.ID] = kind
	// Also register its icon to be rendered on the UI next to runs triggered
	// by this kind.
	db.iconRegistrar.RegisterSourceIcon(kind.GetSource(), triggerIcon(kind.GetSource(), kind.Icon))
}

func (db *kindDB) GetKind(id KindID) (Kind, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	kind, ok := db.kinds[id]
	if !ok {
		return Kind{}, fmt.Errorf("no such vcs provider kind exists: %s", id)
	}
	return kind, nil
}

// GetKindByTFEServiceProviderType retrieves a kind by its TFE service provider
// type.
func (db *kindDB) GetKindByTFEServiceProviderType(sp TFEServiceProviderType) (Kind, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	for _, kind := range db.kinds {
		if slices.Contains(kind.TFEServiceProviders, sp) {
			return kind, nil
		}
	}
	return Kind{}, fmt.Errorf("no such vcs provider kind with TFE service provider type exists: %s", sp)
}

func (db *kindDB) GetKinds() []Kind {
	db.mu.Lock()
	defer db.mu.Unlock()

	return maps.Values(db.kinds)
}

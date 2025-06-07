package vcs

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/a-h/templ"
	"golang.org/x/exp/maps"
)

// KindID of vcs hosting provider
type KindID string

type Kind struct {
	// ID distinguishes this kind from other kinds
	ID KindID
	// Name provides a human meaningful identification of this provider kind.
	Name string
	// Hostname is the hostname of the VCS host, not including scheme or path.
	Hostname string
	// Icon renders a icon distinguishing the VCS host kind.
	Icon templ.Component
	// TokenKind provides info about the token the provider expects. Mutually
	// exclusive with InstallationKind.
	TokenKind *TokenKind
	// InstallationKind provides info about installations for this ProviderKind.
	// Mutually exclusive with TokenKind.
	InstallationKind InstallationKind
	// NewClient constructs a client implementation.
	NewClient func(context.Context, Config) (Client, error)
	// EventHandler handles incoming events from the VCS host before relaying
	// them onwards for triggering actions like creating runs etc.
	EventHandler func(r *http.Request, secret string) (*EventPayload, error)
	// SkipRepohook if true skips the creation of a repository-level webhook.
	SkipRepohook bool
	// TFEServiceProvider optionally registers the kind with the TFE API, permitting vcs
	// providers of this kind to be created via the TFE API. The value specified
	// here is the value that should be be provided by API clients via the
	// "service-provider" attribute:
	//
	// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/oauth-clients#request-body
	TFEServiceProvider TFEServiceProviderType
}

type TokenKind struct {
	// TokenDescription renders a helpful description of what is expected of the
	// token, e.g. what permissions it should possess.
	Description templ.Component
}

type InstallationKind interface {
	// ListInstallations retrieves a list of installations.
	ListInstallations(context.Context) (ListInstallationsResult, error)
	// GetInstallation retrieves an installation by its ID.
	GetInstallation(context.Context, int64) (Installation, error)
}

type ListInstallationsResult struct {
	// InstallationLink is a link to the site where a user can create an
	// installation.
	InstallationLink templ.SafeURL
	// Results is a map of IDs of existing installations keyed by a human
	// meaningful name.
	Results []Installation
}

type Installation struct {
	ID           int64
	AppID        int64
	Username     *string
	Organization *string
}

func (i Installation) String() string {
	if i.Organization != nil {
		return "org/" + *i.Organization
	}
	return "user/" + *i.Username
}

// kindDB is a database of sources and their icons
type kindDB struct {
	mu    sync.Mutex
	kinds map[KindID]Kind
}

func newKindDB() *kindDB {
	return &kindDB{
		kinds: make(map[KindID]Kind),
	}
}

func (db *kindDB) RegisterKind(kind Kind) {
	db.mu.Lock()
	db.kinds[kind.ID] = kind
	db.mu.Unlock()
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

func (db *kindDB) GetKindByTFEServiceProviderType(sp TFEServiceProviderType) (Kind, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	for _, kind := range db.kinds {
		if kind.TFEServiceProvider == sp {
			return kind, nil
		}
	}
	return Kind{}, fmt.Errorf("no such vcs provider kind with TFE service provider type exists: %s", sp)
}

func (db *kindDB) GetKindIDs() []KindID {
	db.mu.Lock()
	defer db.mu.Unlock()

	return maps.Keys(db.kinds)
}

package tokens

import (
	"context"
	"fmt"
	"sync"

	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/resource"
)

type (
	// registry provides a means of registering different kinds of tokens with the
	// authentication middleware, e.g. user tokens, team tokens, etc.
	//
	// Note: the registry provides a level of indirection, helping to not only avoid
	// package import cycles, but also keep the this package focused on token
	// generation and verification for subjects, and leaving the particular handling
	// of the different types of subjects - organizations, teams, users, etc, - to
	// other packages.
	registry struct {
		GetOrCreateUser

		SiteToken string
		SiteAdmin authz.Subject

		kinds map[resource.Kind]SubjectGetter
		mu    sync.Mutex
	}

	// SubjectGetter retrieves an OTF subject given the jwtSubject string, which is the
	// value of the 'subject' field parsed from a JWT.
	SubjectGetter func(ctx context.Context, jwtSubject resource.ID) (authz.Subject, error)

	// GetOrCreateUser retrieves the user with the given username. If the
	// user does not exist it is created.
	GetOrCreateUser func(ctx context.Context, username string) (authz.Subject, error)
)

// RegisterKind registers a kind of authentication token, providing a func that
// can retrieve the OTF subject indicated in the token.
func (r *registry) RegisterKind(k resource.Kind, fn SubjectGetter) {
	r.mu.Lock()
	r.kinds[k] = fn
	r.mu.Unlock()
}

func (r *registry) GetSubject(ctx context.Context, jwtSubject resource.ID) (authz.Subject, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	subjectGetter, ok := r.kinds[jwtSubject.Kind]
	if !ok {
		return nil, fmt.Errorf("unknown authentication token kind: %s", jwtSubject.Kind)
	}
	return subjectGetter(ctx, jwtSubject)
}

// RegisterSiteToken registers a site token which the middleware, and the
// subject to return as the site admin upon successful authentication.
func (r *registry) RegisterSiteToken(token string, siteAdmin authz.Subject) {
	r.SiteToken = token
	r.SiteAdmin = siteAdmin
}

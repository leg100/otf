package tokens

import (
	"context"
	"fmt"
	"sync"

	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/resource"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
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
		kinds map[resource.Kind]SubjectGetter
		mu    sync.Mutex
		key   jwk.Key
	}

	// SubjectGetter retrieves an OTF subject given the jwtSubject string, which is the
	// value of the 'subject' field parsed from a JWT.
	SubjectGetter func(ctx context.Context, jwtSubject resource.TfeID) (authz.Subject, error)
)

// RegisterKind registers a kind of authentication token, providing a func that
// can retrieve the OTF subject indicated in the token.
func (r *registry) RegisterKind(k resource.Kind, fn SubjectGetter) {
	r.mu.Lock()
	r.kinds[k] = fn
	r.mu.Unlock()
}

// GetSubject retrieves the subject from a JWT.
func (r *registry) GetSubject(ctx context.Context, token []byte) (authz.Subject, error) {
	parsed, err := jwt.Parse(token, jwt.WithKey(jwa.HS256, r.key))
	if err != nil {
		return nil, err
	}
	id, err := resource.ParseTfeID(parsed.Subject())
	if err != nil {
		return nil, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	subjectGetter, ok := r.kinds[id.Kind()]
	if !ok {
		return nil, fmt.Errorf("unknown authentication token kind: %s", id.Kind())
	}
	return subjectGetter(ctx, id)
}

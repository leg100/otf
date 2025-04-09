package runner

import "github.com/leg100/otf/internal/authz"

// unregistered describes a runner that is not yet registered.
type unregistered struct {
	// unregistered is a subject only for the purposes of satisfying the
	// token-handling middleware which doesn't call any of the interface
	// methods.
	authz.Subject

	// pool is non-nil if the runner is an agent.
	pool *Pool
}

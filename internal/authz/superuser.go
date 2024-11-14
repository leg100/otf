package authz

import (
	"github.com/leg100/otf/internal/rbac"
)

// Superuser is a subject with unlimited privileges.
type Superuser struct {
	Username string
}

func (*Superuser) CanAccess(rbac.Action, *AccessRequest) bool { return true }
func (s *Superuser) Organizations() []string                  { return nil }
func (s *Superuser) String() string                           { return s.Username }

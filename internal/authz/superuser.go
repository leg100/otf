package authz

import "github.com/leg100/otf/internal/resource"

// Superuser is a subject with unlimited privileges.
type Superuser struct {
	Username string
}

func (*Superuser) CanAccess(resource.Action, resource.Kind, Request) bool { return true }
func (s *Superuser) String() string                                       { return s.Username }

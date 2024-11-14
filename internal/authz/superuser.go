package authz

import (
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/resource"
)

// Superuser is a subject with unlimited privileges.
type Superuser struct {
	Username string
}

func (*Superuser) CanAccess(rbac.Action, *AccessRequest) bool { return true }
func (s *Superuser) Organizations() []string                  { return nil }
func (s *Superuser) String() string                           { return s.Username }
func (s *Superuser) GetID() resource.ID                       { return resource.NewID(resource.UserKind) }
func (s *Superuser) IsSiteAdmin() bool                        { return true }
func (s *Superuser) IsOwner(string) bool                      { return true }

package authz

// Superuser is a subject with unlimited privileges.
type Superuser struct {
	Username string
}

func (*Superuser) CanAccess(Action, AccessRequest) bool { return true }
func (s *Superuser) String() string                     { return s.Username }

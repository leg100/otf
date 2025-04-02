package authz

// Superuser is a subject with unlimited privileges.
type Superuser struct {
	Username string
}

func (*Superuser) CanAccess(Action, Request) bool { return true }
func (s *Superuser) String() string               { return s.Username }

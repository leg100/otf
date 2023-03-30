package cloud

type (
	// User is a user account on a cloud provider.
	User struct {
		Name  string
		Teams []Team // team memberships
	}

	// Team is a team account on a cloud provider.
	Team struct {
		Name         string
		Organization string // team belongs to an organization
	}
)

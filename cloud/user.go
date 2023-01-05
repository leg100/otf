package cloud

// User is a user account on a cloud provider.
type User struct {
	Name string

	Organizations []string // org memberships
	Teams         []Team   // team memberships
}

// Team is a team account on a cloud provider.
type Team struct {
	Name         string
	Organization string // team belongs to an organization
}

package cloud

// User is a user account on a cloud provider.
type User struct {
	Name string

	organizations []string // org memberships
	teams         []Team   // team memberships
}

func (u *User) AddOrganization(org string) {
	u.organizations = append(u.organizations, org)
}

func (u *User) AddTeam(team Team) {
	u.teams = append(u.teams, team)
}

// Team is a team account on a cloud provider.
type Team struct {
	Name         string
	Organization string // team belongs to an organization
}

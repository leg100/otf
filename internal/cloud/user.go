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

func (u User) IsOwner(organization string) bool {
	for _, team := range u.Teams {
		if team.Organization == organization {
			if team.IsOwners() {
				return true
			}
		}
	}
	return false
}

func (u User) Organizations() (organizations []string) {
	// De-dup organizations
	seen := make(map[string]bool)
	for _, t := range u.Teams {
		if _, ok := seen[t.Organization]; ok {
			continue
		}
		organizations = append(organizations, t.Organization)
		seen[t.Organization] = true
	}
	return organizations
}

func (t Team) IsOwners() bool {
	return t.Name == "owners"
}

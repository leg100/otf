package html

import (
	"context"
	"sort"
	"testing"

	"github.com/google/go-github/v41/github"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
)

func TestSynchronise(t *testing.T) {
	tests := []struct {
		name                string
		githubUser          *string
		githubOrganizations []string
		otfUsers            []string
		otfOrganizations    []string
		// expected error
		err error
	}{
		{
			name:                "new user, new org",
			githubUser:          otf.String("leg100"),
			githubOrganizations: []string{"automatize"},
		},
		{
			name:                "existing user, new org",
			githubUser:          otf.String("leg100"),
			githubOrganizations: []string{"automatize"},
			otfUsers:            []string{"leg100"},
		},
		{
			name:                "existing user, existing org",
			githubUser:          otf.String("leg100"),
			githubOrganizations: []string{"automatize"},
			otfUsers:            []string{"leg100"},
			otfOrganizations:    []string{"automatize"},
		},
		{
			name:                "new user, existing org",
			githubUser:          otf.String("leg100"),
			githubOrganizations: []string{"automatize"},
			otfOrganizations:    []string{"automatize"},
		},
		{
			name:                "existing user, two new orgs",
			githubUser:          otf.String("leg100"),
			githubOrganizations: []string{"automatize", "garman"},
			otfUsers:            []string{"leg100"},
		},
		{
			name:                "existing user, existing non-matching org, two new orgs",
			githubUser:          otf.String("leg100"),
			githubOrganizations: []string{"automatize", "garman"},
			otfUsers:            []string{"leg100"},
			otfOrganizations:    []string{"oldcorp"},
		},
		{
			name:       "no github organizations",
			githubUser: otf.String("leg100"),
			err:        ErrNoGithubOrganizationsFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &fakeGithubClient{}
			if tt.githubUser != nil {
				client.user = &github.User{Login: tt.githubUser}
			}
			for _, org := range tt.githubOrganizations {
				client.orgs = append(client.orgs, &github.Organization{Login: otf.String(org)})
			}

			userService := &fakeUserService{database: make(map[string]*otf.User)}
			for _, u := range tt.otfUsers {
				userService.database[u] = &otf.User{Username: u}
			}

			orgService := &fakeOrganizationService{database: make(map[string]*otf.Organization)}
			for _, o := range tt.otfOrganizations {
				orgService.database[o] = &otf.Organization{Name: o}
			}

			user, err := synchronise(context.Background(), client, userService, orgService)
			if err != nil {
				// if there is an error then check that the test is expecting an
				// error
				assert.Equal(t, tt.err, err)
				return
			}

			// Check synchronised user's username matches github login
			assert.Equal(t, *tt.githubUser, user.Username)

			// Check their org membership exactly match their github org
			// membership
			assert.Equal(t, tt.githubOrganizations, getOrganizationNames(user.Organizations))
		})
	}
}

func getOrganizationNames(orgs []*otf.Organization) (names []string) {
	for _, org := range orgs {
		names = append(names, org.Name)
	}
	sort.Strings(names)
	return
}

// Package module is reponsible for registry modules
package module

import (
	"context"
	"fmt"
	"strings"

	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
)

// listTerraformModuleRepos wraps a cloud's ListRepositories endpoint, returning only
// those repositories with a name matching the format
// '<something>-<provider>-<module>'.
//
// NOTE: no pagination is performed, and only matching results from the first page
// are retrieved
func listTerraformModuleRepos(ctx context.Context, client cloud.Client) ([]cloud.Repo, error) {
	list, err := client.ListRepositories(ctx, cloud.ListRepositoriesOptions{
		PageSize: otf.MaxPageSize,
	})
	if err != nil {
		return nil, err
	}
	var filtered []cloud.Repo
	for _, repo := range list {
		_, name, found := strings.Cut(repo.Identifier, "/")
		if !found {
			return nil, fmt.Errorf("malformed identifier: %s", repo.Identifier)
		}
		parts := strings.SplitN(name, "-", 3)
		if len(parts) >= 3 {
			filtered = append(filtered, repo)
		}
	}
	return filtered, nil
}

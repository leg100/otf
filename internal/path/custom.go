package path

import (
	"fmt"

	"github.com/leg100/otf/internal/resource"
)

func Admin() string {
	return "/app/admin"
}

func AdminLogin() string {
	return "/admin/login"
}

func Profile() string {
	return "/app/profile"
}

func Login() string {
	return "/login"
}

func Logout() string {
	return "/app/logout"
}

func Tokens() string {
	return "/app/current-user/tokens"
}

func NewToken() string {
	return "/app/current-user/tokens/new"
}

func CreateToken() string {
	return "/app/current-user/tokens/create"
}

func DeleteToken() string {
	return "/app/current-user/tokens/delete"
}

func SelectGhappOwner() string {
	return "/app/admin/ghapp/select-owner"
}

func GithubApps() string {
	return "/app/github-apps"
}

func CreateGithubApp() string {
	return "/app/github-apps/create"
}

func NewGithubApp() string {
	return "/app/github-apps/new"
}

func GithubApp(githubApp any) string {
	return fmt.Sprintf("/app/github-apps/%v", githubApp)
}

func EditGithubApp(githubApp any) string {
	return fmt.Sprintf("/app/github-apps/%v/edit", githubApp)
}

func UpdateGithubApp(githubApp any) string {
	return fmt.Sprintf("/app/github-apps/%v/update", githubApp)
}

func DeleteGithubApp(githubApp any) string {
	return fmt.Sprintf("/app/github-apps/%v/delete", githubApp)
}

func ExchangeCodeGithubApp() string {
	return "/app/github-apps/exchange-code"
}

func CompleteGithubApp() string {
	return "/app/github-apps/complete"
}

func DeleteInstallGithubApp(githubApp any) string {
	return fmt.Sprintf("/app/github-apps/%v/delete-install", githubApp)
}

func OrganizationToken(org resource.ID) string {
	return fmt.Sprintf("/app/organizations/%v/tokens/show", org)
}

func CreateOrganizationToken(org resource.ID) string {
	return fmt.Sprintf("/app/organizations/%v/tokens/create", org)
}

func DeleteOrganizationToken(org resource.ID) string {
	return fmt.Sprintf("/app/organizations/%v/tokens/delete", org)
}

func VariableSetVariables(vs resource.ID) string {
	return fmt.Sprintf("/app/variable-sets/%v/variable-set-variables", vs)
}

func NewVariableSetVariable(vs resource.ID) string {
	return fmt.Sprintf("/app/variable-sets/%v/variable-set-variables/new", vs)
}

func CreateVariableSetVariable(vs resource.ID) string {
	return fmt.Sprintf("/app/variable-sets/%v/variable-set-variables/create", vs)
}

func VariableSetVariable(id resource.ID) string {
	return fmt.Sprintf("/app/variable-set-variables/%v", id)
}

func EditVariableSetVariable(id resource.ID) string {
	return fmt.Sprintf("/app/variable-set-variables/%v/edit", id)
}

func UpdateVariableSetVariable(id resource.ID) string {
	return fmt.Sprintf("/app/variable-set-variables/%v/update", id)
}

func DeleteVariableSetVariable(id resource.ID) string {
	return fmt.Sprintf("/app/variable-set-variables/%v/delete", id)
}

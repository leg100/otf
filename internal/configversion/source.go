package configversion

const (
	SourceAPI       Source = "tfe-api"
	SourceGithub    Source = "github"
	SourceGitlab    Source = "gitlab"
	SourceTerraform Source = "terraform+cloud"

	DefaultSource = SourceAPI
)

// Source representse of a run.
type Source string

package vcs

const (
	ForgejoKind     Kind = "forgejo"
	GithubTokenKind Kind = "github-token"
	GithubAppKind   Kind = "github-app"
	GitlabKind      Kind = "gitlab"
)

// Kind of vcs hosting provider
type Kind string

func KindPtr(k Kind) *Kind { return &k }

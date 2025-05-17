package vcs

const (
	ForgejoKind Kind = "forgejo"
	GithubKind  Kind = "github"
	GitlabKind  Kind = "gitlab"
)

// Kind of vcs hosting provider
type Kind string

func KindPtr(k Kind) *Kind { return &k }

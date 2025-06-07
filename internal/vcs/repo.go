package vcs

// Repo is a hosted VCS repository
type Repo struct {
	URL   string
	Owner string
	Name  string
}

func (r Repo) String() string {
	return r.Owner + "/" + r.Name
}

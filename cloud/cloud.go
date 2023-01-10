// Package cloud provides types for use with cloud providers.
package cloud

// Webhook is a cloud's configuration for a webhook on OTF.
type Webhook struct {
	ID         string // vcs' ID
	Identifier string // identifier is <repo_owner>/<repo_name>
	Events     []VCSEventType
	Endpoint   string // the OTF URL that receives events
}

// Repo is a VCS repository belonging to a cloud
//
type Repo struct {
	Identifier string `schema:"identifier,required"` // <repo_owner>/<repo_name>
	Branch     string `schema:"branch,required"`     // default branch
}

func (r Repo) ID() string { return r.Identifier }

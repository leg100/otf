package vcs

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"
)

var ErrInvalidRepo = errors.New("repository path is invalid")

// Repo is a hosted VCS repository.
type Repo struct {
	// Owner is the owner of the repo. On some kinds, e.g. Gitlab, the owner is
	// the 'namespace', which may include forward slashes to indicate subgroups.
	Owner string
	// Name is the actual name of the repo, not including the owner or
	// namespace.
	Name string
}

func NewRepoFromString(s string) (Repo, error) {
	var repo Repo

	// split string on last instance of /.
	i := strings.LastIndex(s, "/")
	switch {
	case i < 0:
		// '/' not found
		return Repo{}, errors.New("invalid repo")
	case i == len(s)-1:
		// '/' is last character which is invalid
		return Repo{}, errors.New("invalid repo")
	default:
		// everything before last '/' is the owner
		repo.Owner = s[:i]
		// everything after last '/' is the name
		repo.Name = s[i+1:]
	}
	return repo, nil
}

func (r Repo) String() string {
	return r.Owner + "/" + r.Name
}

func (r *Repo) Scan(text any) error {
	if text == nil {
		return nil
	}
	s, ok := text.(string)
	if !ok {
		return fmt.Errorf("expected database value to be a string: %#v", text)
	}
	repo, err := NewRepoFromString(s)
	if err != nil {
		return err
	}
	*r = repo
	return nil
}

func (r *Repo) Value() (driver.Value, error) {
	if r == nil {
		return nil, nil
	}
	return r.String(), nil
}

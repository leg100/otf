package vcs

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/leg100/otf/internal"
)

var ErrInvalidRepo = errors.New("repository path is invalid")

// Repo is a hosted VCS repository.
type Repo struct {
	owner string
	name  string
}

// Owner is the owner of the repo. On some kinds, e.g. Gitlab, the owner is
// the 'namespace', which may include forward slashes to indicate subgroups.
func (r Repo) Owner() string { return r.owner }

// Name is the actual name of the repo, not including the owner or
// namespace.
func (r Repo) Name() string { return r.name }

func RepoPtr(r Repo) *Repo { return &r }

func NewRepo(owner, name string) (Repo, error) {
	if len(owner) == 0 {
		return Repo{}, errors.New("repo owner cannot be an empty string")
	}
	if len(name) == 0 {
		return Repo{}, errors.New("repo name cannot be an empty string")
	}
	return Repo{owner: owner, name: name}, nil
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
		repo.owner = s[:i]
		// everything after last '/' is the name
		repo.name = s[i+1:]
	}
	return repo, nil
}

func NewMustRepo(owner, name string) Repo {
	repo, err := NewRepo(owner, name)
	if err != nil {
		panic("invalid repo: " + err.Error())
	}
	return repo
}

func NewRandomRepo() Repo {
	return Repo{owner: "repo-owner-" + internal.GenerateRandomString(4), name: "repo-" + internal.GenerateRandomString(4)}
}

func NewRandomModuleRepo(provider, name string) Repo {
	return Repo{owner: uuid.NewString(), name: fmt.Sprintf("terraform-%s-%s", provider, name)}
}

func (r Repo) String() string {
	return r.owner + "/" + r.name
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

func (r Repo) MarshalText() ([]byte, error) {
	return []byte(r.String()), nil
}

func (r *Repo) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		return nil
	}
	s := string(text)
	repo, err := NewRepoFromString(s)
	if err != nil {
		return err
	}
	*r = repo
	return nil
}

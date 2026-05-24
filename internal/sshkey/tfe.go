package sshkey

import (
	"time"

	"github.com/leg100/otf/internal/resource"
)

// TFESSHKey represents an SSH key in the TFE API.
type TFESSHKey struct {
	ID        resource.TfeID `jsonapi:"primary,ssh-keys"`
	CreatedAt time.Time      `jsonapi:"attribute" json:"created-at"`
	UpdatedAt time.Time      `jsonapi:"attribute" json:"updated-at"`
	Name      string         `jsonapi:"attribute" json:"name"`
}

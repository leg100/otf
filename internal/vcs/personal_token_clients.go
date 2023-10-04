package vcs

import (
	"fmt"

	"github.com/leg100/otf/internal"
)

// personalTokenClients is a database of vcs kinds and their constructors for
// creating a vcs client from a personal access token.
var personalTokenClients = internal.NewSafeMap[Kind, PersonalTokenClientConstructor]()

// PersonalTokenClientConstructor constructs a vcs client from a personal
// access token (and hostname).
type PersonalTokenClientConstructor func(hostname, token string) (Client, error)

// RegisterPersonalTokenClientConstructor registers a constructor capable of
// constructing a vcs client from a personal access token.
func RegisterPersonalTokenClientConstructor(kind Kind, constructor PersonalTokenClientConstructor) {
	personalTokenClients.Set(kind, constructor)
}

func GetPersonalTokenClient(kind Kind, hostname, token string) (Client, error) {
	client, ok := personalTokenClients.Get(kind)
	if !ok {
		return nil, fmt.Errorf("unknown vcs kind: %s", kind)
	}
	return client(hostname, token)
}

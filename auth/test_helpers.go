package auth

import (
	"time"
)

type NewTestSessionOption func(*Session)

func OverrideTestRegistrySessionExpiry(expiry time.Time) NewTestSessionOption {
	return func(session *Session) {
		session.expiry = expiry
	}
}

type fakeService struct {
	agentTokenService
	registrySessionService
	sessionService
	userService
}

type fakeHostnameService struct {
	hostname string
}

func (f fakeHostnameService) Hostname() string { return f.hostname }

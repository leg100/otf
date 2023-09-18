package authenticator

import (
	"encoding/json"
	"errors"
)

const (
	EmailClaim claim = "email"
	SubClaim   claim = "sub"
	NameClaim  claim = "name"

	DefaultUsernameClaim = NameClaim
)

type (
	// usernameClaim retrieves a username from an OIDC claim
	usernameClaim struct {
		kind  claim
		value string
	}
	claim string
)

func newUsernameClaim(s string) (*usernameClaim, error) {
	uc := usernameClaim{kind: claim(s)}
	switch uc.kind {
	case EmailClaim, SubClaim, NameClaim:
		return &uc, nil
	default:
		return nil, errors.New("unknown username claim: must be one of email, sub, or name")
	}
}

// oidcClaims depicts the claims returned from the OIDC id-token.
func (uc *usernameClaim) UnmarshalJSON(b []byte) error {
	var token struct {
		Name  string `json:"name"`
		Sub   string `json:"sub"`
		Email string `json:"email"`
	}
	if err := json.Unmarshal(b, &token); err != nil {
		return err
	}
	switch uc.kind {
	case NameClaim:
		uc.value = token.Name
	case SubClaim:
		uc.value = token.Sub
	case EmailClaim:
		uc.value = token.Email
	}
	return nil
}

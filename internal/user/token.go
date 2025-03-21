package user

import (
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/tokens"
)

type (
	// UserToken provides information about an API token for a user.
	UserToken struct {
		ID          resource.TfeID
		CreatedAt   time.Time
		Description string
		Username    string // Token belongs to a user
	}

	// CreateUserTokenOptions are options for creating a user token via the service
	// endpoint
	CreateUserTokenOptions struct {
		Description string
	}

	userTokenFactory struct {
		tokens *tokens.Service
	}
)

func (f *userTokenFactory) NewUserToken(username string, opts CreateUserTokenOptions) (*UserToken, []byte, error) {
	ut := UserToken{
		ID:          resource.NewTfeID(resource.UserTokenKind),
		CreatedAt:   internal.CurrentTimestamp(nil),
		Description: opts.Description,
		Username:    username,
	}
	token, err := f.tokens.NewToken(ut.ID)
	if err != nil {
		return nil, nil, err
	}
	return &ut, token, nil
}

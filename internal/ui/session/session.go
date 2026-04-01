package session

import (
	"net/http"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/ui/helpers"
)

type Service struct {
	client SessionClient
	logger logr.Logger
}

type SessionClient interface {
	NewToken(subjectID resource.ID, expiry *time.Time) ([]byte, error)
}

func NewService(logger logr.Logger, client SessionClient) *Service {
	svc := &Service{
		client: client,
		logger: logger,
	}
	return svc
}

func (s *Service) StartSession(w http.ResponseWriter, r *http.Request, userID resource.ID) error {
	expiry := internal.CurrentTimestamp(nil).Add(defaultSessionExpiry)
	token, err := s.client.NewToken(userID, &expiry)
	if err != nil {
		return err
	}
	// Set cookie to expire at same time as token
	helpers.SetCookie(w, SessionCookie, string(token), new(expiry))
	helpers.ReturnUserOriginalPage(w, r)

	// TODO: log username instead
	s.logger.V(2).Info("started session", "user_id", userID)

	return nil
}

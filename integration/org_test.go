package integration

import (
	"testing"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/organization"
)

func newOrganizationService(t *testing.T, db otf.DB) organization.Service {
	return organization.NewService(organization.Options{
		Logger: logr.Discard(),
		DB:     db,
		Broker: testBroker(t, db),
	})
}

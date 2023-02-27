package auth

import (
	"github.com/go-logr/logr"
	"github.com/leg100/otf"
)

type application interface {
	agentTokenApp
	registrySessionApp
	sessionApp
	teamApp
	tokenApp
	userApp
}

type app struct {
	otf.OrganizationAuthorizer
	logr.Logger

	db *pgdb
	*synchroniser
}

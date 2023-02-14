package auth

import (
	"github.com/go-logr/logr"
	"github.com/leg100/otf"
)

type app interface {
	agentTokenApp
	registrySessionApp
	sessionApp
	teamApp
	userApp
}

type Application struct {
	otf.Authorizer
	logr.Logger

	db db
	*synchroniser
}

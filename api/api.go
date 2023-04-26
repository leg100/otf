// Package api providers http handlers for the API.
package api

type (
	api struct {
	}
)

func (a *api) addHandlers(r *mux.Router) {

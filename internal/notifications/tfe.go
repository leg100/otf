package notifications

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/api"
	"github.com/leg100/otf/internal/api/types"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/http/decode"
)

type tfe struct {
	Service
	*api.Responder
}

func (a *tfe) addHandlers(r *mux.Router) {
	r = otfhttp.APIRouter(r)

	r.HandleFunc("/workspaces/{workspace_id}/notification-configurations", a.createNotification).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/notification-configurations", a.listNotifications).Methods("GET")
	r.HandleFunc("/notification-configurations/{id}", a.getNotification).Methods("GET")
	r.HandleFunc("/notification-configurations/{id}", a.updateNotification).Methods("PATCH")
	r.HandleFunc("/notification-configurations/{id}", a.verifyNotification).Methods("POST")
	r.HandleFunc("/notification-configurations/{id}", a.deleteNotification).Methods("DELETE")
}

func (a *tfe) createNotification(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		api.Error(w, err)
		return
	}
	var params types.NotificationConfigurationCreateOptions
	if err := api.Unmarshal(r.Body, &params); err != nil {
		api.Error(w, err)
		return
	}
	if params.DestinationType == nil {
		api.Error(w, &internal.MissingParameterError{Parameter: "destination_type"})
		return
	}

	opts := CreateConfigOptions{
		DestinationType: (Destination)(*params.DestinationType),
		Enabled:         params.Enabled,
		Name:            params.Name,
		URL:             params.URL,
	}
	for _, t := range params.Triggers {
		opts.Triggers = append(opts.Triggers, Trigger(t))
	}

	nc, err := a.CreateNotificationConfiguration(r.Context(), workspaceID, opts)
	if err != nil {
		api.Error(w, err)
		return
	}

	a.Respond(w, r, a.convert(nc), http.StatusCreated)
}

func (a *tfe) listNotifications(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		api.Error(w, err)
		return
	}

	configs, err := a.ListNotificationConfigurations(r.Context(), workspaceID)
	if err != nil {
		api.Error(w, err)
		return
	}

	a.Respond(w, r, configs, http.StatusOK)
}

func (a *tfe) getNotification(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("id", r)
	if err != nil {
		api.Error(w, err)
		return
	}

	nc, err := a.GetNotificationConfiguration(r.Context(), id)
	if err != nil {
		api.Error(w, err)
		return
	}

	a.Respond(w, r, nc, http.StatusOK)
}

func (a *tfe) updateNotification(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("id", r)
	if err != nil {
		api.Error(w, err)
		return
	}
	var params types.NotificationConfigurationUpdateOptions
	if err := api.Unmarshal(r.Body, &params); err != nil {
		api.Error(w, err)
		return
	}

	opts := UpdateConfigOptions{
		Enabled: params.Enabled,
		Name:    params.Name,
		URL:     params.URL,
	}
	for _, t := range params.Triggers {
		opts.Triggers = append(opts.Triggers, Trigger(t))
	}

	updated, err := a.UpdateNotificationConfiguration(r.Context(), id, opts)
	if err != nil {
		api.Error(w, err)
		return
	}

	a.Respond(w, r, updated, http.StatusOK)
}

func (a *tfe) verifyNotification(w http.ResponseWriter, r *http.Request) {}

func (a *tfe) deleteNotification(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("id", r)
	if err != nil {
		api.Error(w, err)
		return
	}

	if err := a.DeleteNotificationConfiguration(r.Context(), id); err != nil {
		api.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *tfe) convert(from *Config) *types.NotificationConfiguration {
	to := &types.NotificationConfiguration{
		ID:              from.ID,
		CreatedAt:       from.CreatedAt,
		UpdatedAt:       from.UpdatedAt,
		Name:            from.Name,
		Enabled:         from.Enabled,
		DestinationType: types.NotificationDestinationType(from.DestinationType),
		Subscribable: &types.Workspace{
			ID: from.WorkspaceID,
		},
	}
	if from.URL != nil {
		to.URL = *from.URL
	}
	for _, t := range from.Triggers {
		to.Triggers = append(to.Triggers, string(t))
	}
	return to
}

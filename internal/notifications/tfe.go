package notifications

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/workspace"
)

type tfe struct {
	*Service
	*tfeapi.Responder
}

func (a *tfe) addHandlers(r *mux.Router) {
	r = r.PathPrefix(tfeapi.APIPrefixV2).Subrouter()

	r.HandleFunc("/workspaces/{workspace_id}/notification-configurations", a.createNotification).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/notification-configurations", a.listNotifications).Methods("GET")
	r.HandleFunc("/notification-configurations/{id}", a.getNotification).Methods("GET")
	r.HandleFunc("/notification-configurations/{id}", a.updateNotification).Methods("PATCH")
	r.HandleFunc("/notification-configurations/{id}", a.verifyNotification).Methods("POST")
	r.HandleFunc("/notification-configurations/{id}", a.deleteNotification).Methods("DELETE")
}

func (a *tfe) createNotification(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	var params TFENotificationConfigurationCreateOptions
	if err := tfeapi.Unmarshal(r.Body, &params); err != nil {
		tfeapi.Error(w, err)
		return
	}
	if params.DestinationType == nil {
		tfeapi.Error(w, &internal.ErrMissingParameter{Parameter: "destination_type"})
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

	nc, err := a.Create(r.Context(), workspaceID, opts)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, a.convert(nc), http.StatusCreated)
}

func (a *tfe) listNotifications(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	configs, err := a.List(r.Context(), workspaceID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	// convert items
	to := make([]*TFENotificationConfiguration, len(configs))
	for i, from := range configs {
		to[i] = a.convert(from)
	}
	a.Respond(w, r, to, http.StatusOK)
}

func (a *tfe) getNotification(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	nc, err := a.Get(r.Context(), id)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, a.convert(nc), http.StatusOK)
}

func (a *tfe) updateNotification(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	var params TFENotificationConfigurationUpdateOptions
	if err := tfeapi.Unmarshal(r.Body, &params); err != nil {
		tfeapi.Error(w, err)
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

	updated, err := a.Update(r.Context(), id, opts)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, a.convert(updated), http.StatusOK)
}

func (a *tfe) verifyNotification(w http.ResponseWriter, r *http.Request) {}

func (a *tfe) deleteNotification(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	if err := a.Delete(r.Context(), id); err != nil {
		tfeapi.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *tfe) convert(from *Config) *TFENotificationConfiguration {
	to := &TFENotificationConfiguration{
		ID:              from.ID,
		CreatedAt:       from.CreatedAt,
		UpdatedAt:       from.UpdatedAt,
		Name:            from.Name,
		Enabled:         from.Enabled,
		DestinationType: TFENotificationDestinationType(from.DestinationType),
		Subscribable: &workspace.TFEWorkspace{
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

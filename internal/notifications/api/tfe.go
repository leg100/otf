package api

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/notifications"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/workspace"
)

type TFEAPI struct {
	*tfeapi.Responder
	Client tfeClient
}

type tfeClient interface {
	CreateNotificationConfig(ctx context.Context, workspaceID resource.TfeID, opts notifications.CreateConfigOptions) (*notifications.Config, error)
	UpdateNotificationConfig(ctx context.Context, id resource.TfeID, opts notifications.UpdateConfigOptions) (*notifications.Config, error)
	GetNotificationConfig(ctx context.Context, id resource.TfeID) (*notifications.Config, error)
	ListNotificationConfigs(ctx context.Context, workspaceID resource.TfeID) ([]*notifications.Config, error)
	DeleteNotificationConfig(ctx context.Context, id resource.TfeID) error
}

func (a *TFEAPI) AddHandlers(r *mux.Router) {
	r.HandleFunc("/workspaces/{workspace_id}/notification-configurations", a.createNotification).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/notification-configurations", a.listNotifications).Methods("GET")
	r.HandleFunc("/notification-configurations/{id}", a.getNotification).Methods("GET")
	r.HandleFunc("/notification-configurations/{id}", a.updateNotification).Methods("PATCH")
	r.HandleFunc("/notification-configurations/{id}", a.verifyNotification).Methods("POST")
	r.HandleFunc("/notification-configurations/{id}", a.deleteNotification).Methods("DELETE")
}

func (a *TFEAPI) createNotification(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	var params notifications.TFENotificationConfigurationCreateOptions
	if err := tfeapi.Unmarshal(r.Body, &params); err != nil {
		tfeapi.Error(w, err)
		return
	}
	if params.DestinationType == nil {
		tfeapi.Error(w, &internal.ErrMissingParameter{Parameter: "destination_type"})
		return
	}

	opts := notifications.CreateConfigOptions{
		DestinationType: (notifications.Destination)(*params.DestinationType),
		Enabled:         params.Enabled,
		Name:            params.Name,
		URL:             params.URL,
	}
	for _, t := range params.Triggers {
		opts.Triggers = append(opts.Triggers, notifications.Trigger(t))
	}

	nc, err := a.Client.CreateNotificationConfig(r.Context(), workspaceID, opts)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, a.convert(nc), http.StatusCreated)
}

func (a *TFEAPI) listNotifications(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	configs, err := a.Client.ListNotificationConfigs(r.Context(), workspaceID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	// convert items
	to := make([]*notifications.TFENotificationConfiguration, len(configs))
	for i, from := range configs {
		to[i] = a.convert(from)
	}
	a.Respond(w, r, to, http.StatusOK)
}

func (a *TFEAPI) getNotification(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	nc, err := a.Client.GetNotificationConfig(r.Context(), id)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, a.convert(nc), http.StatusOK)
}

func (a *TFEAPI) updateNotification(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	var params notifications.TFENotificationConfigurationUpdateOptions
	if err := tfeapi.Unmarshal(r.Body, &params); err != nil {
		tfeapi.Error(w, err)
		return
	}

	opts := notifications.UpdateConfigOptions{
		Enabled: params.Enabled,
		Name:    params.Name,
		URL:     params.URL,
	}
	for _, t := range params.Triggers {
		opts.Triggers = append(opts.Triggers, notifications.Trigger(t))
	}

	updated, err := a.Client.UpdateNotificationConfig(r.Context(), id, opts)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, a.convert(updated), http.StatusOK)
}

func (a *TFEAPI) verifyNotification(w http.ResponseWriter, r *http.Request) {}

func (a *TFEAPI) deleteNotification(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	if err := a.Client.DeleteNotificationConfig(r.Context(), id); err != nil {
		tfeapi.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *TFEAPI) convert(from *notifications.Config) *notifications.TFENotificationConfiguration {
	to := &notifications.TFENotificationConfiguration{
		ID:              from.ID,
		CreatedAt:       from.CreatedAt,
		UpdatedAt:       from.UpdatedAt,
		Name:            from.Name,
		Enabled:         from.Enabled,
		DestinationType: notifications.TFENotificationDestinationType(from.DestinationType),
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

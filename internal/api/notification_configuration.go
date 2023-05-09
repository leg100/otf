package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/api/types"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/http/decode"
)

func (a *api) addNotificationHandlers(r *mux.Router) {
	r = otfhttp.APIRouter(r)

	r.HandleFunc("/workspaces/{workspace_id}/notification-configurations", a.createNotification).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/notification-configurations", a.listNotifications).Methods("GET")
	r.HandleFunc("/notification-configurations/{id}", a.getNotification).Methods("GET")
	r.HandleFunc("/notification-configurations/{id}", a.updateNotification).Methods("PATCH")
	r.HandleFunc("/notification-configurations/{id}", a.verifyNotification).Methods("POST")
	r.HandleFunc("/notification-configurations/{id}", a.deleteNotification).Methods("DELETE")
}

func (a *api) createNotification(w http.ResponseWriter, r *http.Request) {
	var params types.NotificationConfigurationCreateOptions
	if err := decode.All(&params, r); err != nil {
		Error(w, err)
		return
	}
}

func (a *api) listNotifications(w http.ResponseWriter, r *http.Request)  {}
func (a *api) getNotification(w http.ResponseWriter, r *http.Request)    {}
func (a *api) updateNotification(w http.ResponseWriter, r *http.Request) {}
func (a *api) verifyNotification(w http.ResponseWriter, r *http.Request) {}
func (a *api) deleteNotification(w http.ResponseWriter, r *http.Request) {}

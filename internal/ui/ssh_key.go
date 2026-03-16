package ui

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/sshkey"
	"github.com/leg100/otf/internal/ui/helpers"
	"github.com/leg100/otf/internal/ui/paths"
)

// addSSHKeyHandlers registers ssh key UI handlers with the router
func addSSHKeyHandlers(r *mux.Router, h *Handlers) {
	r.HandleFunc("/organizations/{organization_name}/ssh-keys", h.listSSHKeys).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/ssh-keys/create", h.createSSHKey).Methods("POST")
	r.HandleFunc("/ssh-keys/{ssh_key_id}/delete", h.deleteSSHKey).Methods("POST")
}

func (h *Handlers) createSSHKey(w http.ResponseWriter, r *http.Request) {
	var opts sshkey.CreateOptions
	if err := decode.All(&opts, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	key, err := h.SSHKeys.CreateSSHKey(r.Context(), opts)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	helpers.FlashSuccess(w, "created ssh key: "+key.Name)
	http.Redirect(w, r, paths.SSHKeys(opts.Organization), http.StatusFound)
}

func (h *Handlers) listSSHKeys(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name organization.Name `schema:"organization_name"`
	}
	if err := decode.All(&params, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	keys, err := h.SSHKeys.ListSSHKeys(r.Context(), params.Name)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	helpers.RenderPage(
		h.templates.listSSHKeys(listSSHKeysProps{
			organization: params.Name,
			keys:         keys,
		}),
		"ssh keys",
		w,
		r,
		helpers.WithOrganization(params.Name),
		helpers.WithBreadcrumbs(
			helpers.Breadcrumb{Name: "SSH Keys"},
		),
	)
}

func (h *Handlers) deleteSSHKey(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("ssh_key_id", r)
	if err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	key, err := h.SSHKeys.DeleteSSHKey(r.Context(), id)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	helpers.FlashSuccess(w, "deleted ssh key: "+key.Name)
	http.Redirect(w, r, paths.SSHKeys(key.Organization), http.StatusFound)
}

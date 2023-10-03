package repo

import (
	"context"
	"net/http"
	"path"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/vcs"
	"github.com/leg100/otf/internal/vcsprovider"
)

const (
	// handlerPrefix is the URL path prefix for endpoints receiving vcs events
	handlerPrefix = "/webhooks/vcs"
)

type (
	// handlers handle VCS events triggered by webhooks
	handlers struct {
		logr.Logger

		vcs.Publisher
		github.GithubAppService
		vcsprovider.VCSProviderService

		cloudHandlers *internal.SafeMap[cloud.Kind, CloudHandler]

		handlerDB
	}

	// CloudHandler does two things:
	// (a) handles incoming request (containing a VCS event) and sends appropriate response
	// (b) unmarshals event from the request; if the event is irrelevant or
	// invalid then nil is returned.
	//
	// TODO: rename to UnmarshalEvent (?)
	CloudHandler func(w http.ResponseWriter, r *http.Request, secret string) *vcs.EventPayload

	// handleDB is the database the handler interacts with
	handlerDB interface {
		getHookByID(context.Context, uuid.UUID) (*hook, error)
	}
)

func newHandler(logger logr.Logger, publisher vcs.Publisher, db handlerDB, githubAppService github.GithubAppService) *handlers {
	return &handlers{
		Logger:           logger,
		Publisher:        publisher,
		GithubAppService: githubAppService,
		handlerDB:        db,
		cloudHandlers:    internal.NewSafeMap[cloud.Kind, CloudHandler](),
	}
}

func (h *handlers) AddHandlers(r *mux.Router) {
	r.HandleFunc(github.AppEventsPath, h.githubAppHandler)
	r.HandleFunc(path.Join(handlerPrefix, "{webhook_id}"), h.repohookHandler)
}

func (h *handlers) repohookHandler(w http.ResponseWriter, r *http.Request) {
	var opts struct {
		ID uuid.UUID `schema:"webhook_id,required"`
	}
	if err := decode.All(&opts, r); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	hook, err := h.getHookByID(r.Context(), opts.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	h.V(2).Info("received vcs event", "repohook_id", opts.ID, "repo", hook.identifier, "cloud", hook.cloud)

	cloudHandler, ok := h.cloudHandlers.Get(hook.cloud)
	if !ok {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if payload := cloudHandler(w, r, hook.secret); payload != nil {
		h.Publish(vcs.Event{
			EventHeader: vcs.EventHeader{
				VCSProviderID: hook.vcsProviderID,
			},
			EventPayload: *payload,
		})
	}
}

func (h *handlers) githubAppHandler(w http.ResponseWriter, r *http.Request) {
	// retrieve github app config; if one hasn't been configured then return a
	// 400
	app, err := h.GetGithubApp(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// use github-specific handler to unmarshal event
	payload := github.HandleEvent(w, r, app.WebhookSecret)
	if payload == nil {
		return
	}
	h.V(2).Info("received vcs event", "type", "github-app", "repo", payload.RepoPath)
	// relay a copy of the event for each vcs provider configured with the
	// github app install that triggered the event.
	providers, err := h.ListVCSProvidersByGithubAppInstall(r.Context(), *payload.GithubAppInstallID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, prov := range providers {
		h.Publish(vcs.Event{
			EventHeader:  vcs.EventHeader{VCSProviderID: prov.ID},
			EventPayload: *payload,
		})
	}
}

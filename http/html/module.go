package html

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
)

// moduleRepoRequest provides metadata about a request for a module repo
type moduleRepoRequest struct {
	organizationRequest
}

func (r moduleRepoRequest) VCSProviderID() string {
	return param(r.r, "vcs_provider_id")
}

func (app *Application) listModules(w http.ResponseWriter, r *http.Request) {
	var opts otf.ListModulesOptions
	if err := decode.Route(&opts, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	mdoules, err := app.ListModules(r.Context(), opts)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.render("module_list.tmpl", w, r, struct {
		Items []*otf.Module
		organizationRoute
	}{
		Items:             mdoules,
		organizationRoute: organizationRequest{r},
	})
}

func (app *Application) listModuleVCSProviders(w http.ResponseWriter, r *http.Request) {
	providers, err := app.ListVCSProviders(r.Context(), mux.Vars(r)["organization_name"])
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.render("module_vcs_provider_list.tmpl", w, r, struct {
		Items []*otf.VCSProvider
		organizationRoute
	}{
		Items:             providers,
		organizationRoute: organizationRequest{r},
	})
}

type organizationParams struct {
	Organization string `schema:"organization_name,required"`
}

type vcsProviderParams struct {
	VCSProviderID string `schema:"vcs_provider_id,required"`

	organizationParams
}

func (app *Application) listModuleVCSRepos(w http.ResponseWriter, r *http.Request) {
	type options struct {
		vcsProviderParams
		otf.ListOptions // Pagination
		// TODO: filters, public/private, etc
	}
	var opts options
	if err := decode.All(&opts, r); err != nil {
		writeError(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	provider, err := app.GetVCSProvider(r.Context(), opts.VCSProviderID)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	repos, err := app.ListRepositories(r.Context(), opts.VCSProviderID, opts.ListOptions)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.render("module_vcs_repo_list.tmpl", w, r, struct {
		*otf.RepoList
		*otf.VCSProvider
	}{
		RepoList:    repos,
		VCSProvider: provider,
	})
}

package ui

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/policy"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/ui/paths"
	"github.com/leg100/otf/internal/vcs"
	"github.com/leg100/otf/internal/workspace"
)

// Handlers is the collection of UI handlers
type Handlers struct {
	Handlers     []internal.Handlers
	Authorizer   authz.Interface
	Policies     PolicyService
	VCSProviders VCSProviderService
	Workspaces   WorkspaceService
}

type PolicyService interface {
	CreatePolicySet(ctx context.Context, org organization.Name, opts policy.CreatePolicySetOptions) (*policy.PolicySet, error)
	CreateVCSPolicySet(ctx context.Context, org organization.Name, opts policy.CreateVCSPolicySetOptions) (*policy.PolicySet, []*policy.Policy, error)
	ListPolicySets(ctx context.Context, org organization.Name) ([]*policy.PolicySet, error)
	GetPolicySet(ctx context.Context, id resource.TfeID) (*policy.PolicySet, error)
	UpdatePolicySet(ctx context.Context, id resource.TfeID, opts policy.UpdatePolicySetOptions) (*policy.PolicySet, error)
	DeletePolicySet(ctx context.Context, id resource.TfeID) error
	ListImportablePolicies(ctx context.Context, providerID resource.TfeID, repo vcs.Repo, ref, subpath string) ([]policy.ImportablePolicy, error)
	SyncPolicySetFromVCS(ctx context.Context, setID resource.TfeID) (*policy.SyncResult, error)
	CreatePolicy(ctx context.Context, setID resource.TfeID, opts policy.CreatePolicyOptions) (*policy.Policy, error)
	ListPolicies(ctx context.Context, setID resource.TfeID) ([]*policy.Policy, error)
	GetPolicy(ctx context.Context, id resource.TfeID) (*policy.Policy, error)
	UpdatePolicy(ctx context.Context, id resource.TfeID, opts policy.UpdatePolicyOptions) (*policy.Policy, error)
	DeletePolicy(ctx context.Context, id resource.TfeID) error
	ListPolicyChecks(ctx context.Context, runID resource.TfeID) ([]*policy.PolicyCheck, error)
	GenerateWorkspaceMocks(ctx context.Context, workspaceID resource.TfeID, runID *resource.TfeID) ([]byte, error)
}

type VCSProviderService interface {
	ListVCSProviders(ctx context.Context, org organization.Name) ([]*vcs.Provider, error)
	GetVCSProvider(ctx context.Context, id resource.TfeID) (*vcs.Provider, error)
}

type WorkspaceService interface {
	GetWorkspace(context.Context, resource.TfeID) (*workspace.Workspace, error)
}

func (h *Handlers) AddHandlers(r *mux.Router) {
	// Root path redirects to the organization list
	r.Handle("/", http.RedirectHandler("/app/organizations", http.StatusFound))

	r = r.PathPrefix(paths.UIPrefix).Subrouter()
	if h.Policies != nil && h.VCSProviders != nil && h.Workspaces != nil {
		addPolicyHandlers(r, h)
	}
	for _, handler := range h.Handlers {
		handler.AddHandlers(r)
	}
}

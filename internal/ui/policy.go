package ui

import (
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"time"

	"github.com/a-h/templ"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/policy"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/ui/helpers"
	"github.com/leg100/otf/internal/ui/paths"
	"github.com/leg100/otf/internal/vcs"
)

func addPolicyHandlers(r *mux.Router, h *Handlers) {
	r.HandleFunc("/organizations/{organization_name}/policy-sets", h.listPolicySets).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/policy-sets/connect", h.connectPolicySet).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/policy-sets/connect/manual", h.newManualPolicySet).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/policy-sets/connect/vcs", h.newVCSPolicySet).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/policy-sets/connect/vcs/preview", h.previewVCSPolicySet).Methods("POST")
	r.HandleFunc("/organizations/{organization_name}/policy-sets/connect/vcs/create", h.createVCSPolicySet).Methods("POST")
	r.HandleFunc("/organizations/{organization_name}/policy-sets/create", h.createPolicySet).Methods("POST")
	r.HandleFunc("/policy-sets/{policy_set_id}", h.policySet).Methods("GET")
	r.HandleFunc("/policy-sets/{policy_set_id}/update", h.updatePolicySet).Methods("POST")
	r.HandleFunc("/policy-sets/{policy_set_id}/delete", h.deletePolicySet).Methods("POST")
	r.HandleFunc("/policy-sets/{policy_set_id}/sync", h.syncPolicySet).Methods("POST")
	r.HandleFunc("/policy-sets/{policy_set_id}/policies/create", h.createPolicy).Methods("POST")
	r.HandleFunc("/policies/{policy_id}/update", h.updatePolicy).Methods("POST")
	r.HandleFunc("/policies/{policy_id}/delete", h.deletePolicy).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/sentinel", h.workspaceSentinel).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/sentinel/mocks", h.downloadWorkspaceMocks).Methods("GET")
}

func (h *Handlers) listPolicySets(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Organization organization.Name `schema:"organization_name,required"`
	}
	if err := decode.All(&params, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	sets, err := h.Policies.ListPolicySets(r.Context(), params.Organization)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}
	component := templ.ComponentFunc(func(ctx context.Context, out io.Writer) error {
		esc := html.EscapeString
		fmt.Fprint(out, `<div class="flex flex-col gap-6">`)
		fmt.Fprint(out, `<div class="alert alert-info"><span>Policy sets are organization-wide. Every policy set on this page applies to every workspace in this organization.</span></div>`)
		fmt.Fprintf(out, `<div class="flex justify-end"><a class="btn" href="%s">Connect a new policy set</a></div>`, paths.ConnectPolicySet(params.Organization))
		fmt.Fprint(out, `<div class="grid grid-cols-1 gap-4">`)
		for _, set := range sets {
			policies, _ := h.Policies.ListPolicies(r.Context(), set.ID)
			fmt.Fprintf(out, `<a class="border border-base-300 rounded-box p-5 flex flex-col gap-3 hover:border-primary transition-colors" href="%s">`, paths.PolicySet(set.ID))
			fmt.Fprintf(out, `<div class="flex items-center gap-2 flex-wrap"><h3 class="font-bold text-lg">%s</h3><span class="badge badge-outline">%s</span><span class="text-sm opacity-70">%d policies</span></div>`, esc(set.Name), esc(string(set.Source)), len(policies))
			if set.Description != "" {
				fmt.Fprintf(out, `<p class="opacity-70 break-words">%s</p>`, esc(set.Description))
			}
			if set.Source == policy.VCSPolicySetSource {
				fmt.Fprintf(out, `<div class="text-sm opacity-80">Repo: <span class="font-mono">%s</span>`, esc(policySetRepoLabel(set)))
				if set.VCSPath != "" {
					fmt.Fprintf(out, ` | Path: <span class="font-mono">%s</span>`, esc(set.VCSPath))
				}
				if set.VCSRef != "" {
					fmt.Fprintf(out, ` | Ref: <span class="font-mono">%s</span>`, esc(set.VCSRef))
				}
				if set.LastSyncedAt != nil {
					fmt.Fprintf(out, ` | Last synced: %s`, set.LastSyncedAt.Format(time.RFC3339))
				}
				fmt.Fprint(out, `</div>`)
			} else {
				fmt.Fprint(out, `<div class="text-sm opacity-80">Applies to all workspaces in this organization.</div>`)
			}
			fmt.Fprint(out, `</a>`)
		}
		if len(sets) == 0 {
			fmt.Fprint(out, `<div class="border border-dashed border-base-300 rounded-box p-8 text-center opacity-70">No policy sets yet.</div>`)
		}
		fmt.Fprint(out, `</div></div>`)
		return nil
	})
	helpers.RenderPage(component, "policy sets", w, r, helpers.WithOrganization(params.Organization), helpers.WithBreadcrumbs(helpers.Breadcrumb{Name: "Policy Sets"}))
}

func policySetRepoLabel(set *policy.PolicySet) string {
	if set.VCSRepo == nil {
		return ""
	}
	return set.VCSRepo.String()
}

func (h *Handlers) policySet(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("policy_set_id", r)
	if err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	set, err := h.Policies.GetPolicySet(r.Context(), id)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}
	policies, err := h.Policies.ListPolicies(r.Context(), id)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}
	component := templ.ComponentFunc(func(ctx context.Context, out io.Writer) error {
		esc := html.EscapeString
		fmt.Fprint(out, `<div class="flex flex-col gap-6">`)
		fmt.Fprintf(out, `<div><h1 class="text-3xl font-bold">%s</h1><p class="opacity-70 mt-2">Configure this policy set and review the policies it contains.</p></div>`, esc(set.Name))
		fmt.Fprint(out, `<section class="border border-base-300 rounded-box p-4 flex flex-col gap-4">`)
		fmt.Fprintf(out, `<div class="flex items-center gap-2 flex-wrap"><span class="badge badge-outline">%s</span>`, esc(string(set.Source)))
		if set.LastSyncedAt != nil {
			fmt.Fprintf(out, `<span class="text-sm opacity-70">Last synced %s</span>`, set.LastSyncedAt.Format(time.RFC3339))
		}
		fmt.Fprint(out, `</div>`)
		fmt.Fprintf(out, `<form class="grid grid-cols-1 lg:grid-cols-2 gap-4" method="POST" action="%s">`, paths.UpdatePolicySet(set.ID))
		fmt.Fprintf(out, `<div class="flex flex-col gap-1"><label>Name</label><input class="input w-full" name="name" value="%s"></div>`, esc(set.Name))
		fmt.Fprintf(out, `<div class="flex flex-col gap-1"><label>Description</label><input class="input w-full" name="description" value="%s"></div>`, esc(set.Description))
		if set.Source == policy.VCSPolicySetSource {
			fmt.Fprintf(out, `<div class="text-sm opacity-80 lg:col-span-2">Repo: <span class="font-mono">%s</span>`, esc(policySetRepoLabel(set)))
			if set.VCSPath != "" {
				fmt.Fprintf(out, ` | Path: <span class="font-mono">%s</span>`, esc(set.VCSPath))
			}
			if set.VCSRef != "" {
				fmt.Fprintf(out, ` | Ref: <span class="font-mono">%s</span>`, esc(set.VCSRef))
			}
			fmt.Fprint(out, `</div>`)
		} else {
			fmt.Fprint(out, `<div class="text-sm opacity-80 lg:col-span-2">Applies to all workspaces in this organization.</div>`)
		}
		fmt.Fprint(out, `<div class="flex gap-2">`)
		fmt.Fprint(out, `<button class="btn" type="submit">Save</button>`)
		if set.Source == policy.VCSPolicySetSource {
			fmt.Fprintf(out, `<button class="btn" formaction="%s" type="submit">Sync from VCS</button>`, paths.SyncPolicySet(set.ID))
		}
		fmt.Fprintf(out, `<button class="btn btn-error" formaction="%s" type="submit">Delete</button>`, paths.DeletePolicySet(set.ID))
		fmt.Fprint(out, `</div></form></section>`)

		if set.Source != policy.VCSPolicySetSource {
			fmt.Fprintf(out, `<section class="border border-base-300 rounded-box p-4 flex flex-col gap-4"><div><h2 class="text-xl font-bold">Add policy</h2><p class="opacity-70">Create a Sentinel policy managed directly in OTF.</p></div><form class="grid grid-cols-1 lg:grid-cols-3 gap-3" method="POST" action="%s">`, paths.CreatePolicy(set.ID))
			fmt.Fprint(out, `<div class="flex flex-col gap-1"><label>Name</label><input class="input w-full" name="name" required></div>`)
			fmt.Fprint(out, `<div class="flex flex-col gap-1"><label>Level</label><select class="select w-full" name="enforcement_level"><option value="mandatory">mandatory</option><option value="advisory">advisory</option></select></div>`)
			fmt.Fprint(out, `<div class="flex flex-col gap-1"><label>Description</label><input class="input w-full" name="description"></div>`)
			fmt.Fprint(out, `<div class="flex flex-col gap-1 lg:col-span-3"><label>Source</label><textarea class="textarea min-h-48 w-full font-mono" name="source" required></textarea></div>`)
			fmt.Fprint(out, `<div><button class="btn" type="submit">Add policy</button></div></form></section>`)
		}

		fmt.Fprint(out, `<section class="border border-base-300 rounded-box p-4 flex flex-col gap-4"><div><h2 class="text-xl font-bold">Policies</h2></div><div class="flex flex-col gap-3">`)
		for _, p := range policies {
			fmt.Fprintf(out, `<form class="border border-base-300 rounded-box p-4 grid grid-cols-1 lg:grid-cols-3 gap-3 items-start" method="POST" action="%s">`, paths.UpdatePolicy(p.ID))
			readonly := ""
			if set.Source == policy.VCSPolicySetSource {
				readonly = " readonly"
			}
			fmt.Fprintf(out, `<div class="flex flex-col gap-1"><label>Name</label><input class="input input-sm w-full" name="name" value="%s"%s></div>`, esc(p.Name), readonly)
			fmt.Fprintf(out, `<div class="flex flex-col gap-1"><label>Level</label><select class="select select-sm w-full" name="enforcement_level"><option value="mandatory"%s>mandatory</option><option value="advisory"%s>advisory</option></select></div>`,
				selectedAttr(p.EnforcementLevel == policy.MandatoryEnforcement),
				selectedAttr(p.EnforcementLevel == policy.AdvisoryEnforcement),
			)
			fmt.Fprintf(out, `<div class="flex flex-col gap-1"><label>Description</label><input class="input input-sm w-full" name="description" value="%s"></div>`, esc(p.Description))
			if p.Path != "" {
				fmt.Fprintf(out, `<div class="text-sm opacity-70 lg:col-span-3">Path: <span class="font-mono">%s</span></div>`, esc(p.Path))
			}
			fmt.Fprintf(out, `<div class="flex flex-col gap-1 lg:col-span-3"><label>Source</label><textarea class="textarea textarea-sm min-h-48 w-full font-mono" name="source"%s>%s</textarea></div>`, readonly, esc(p.Source))
			fmt.Fprint(out, `<div class="flex gap-2">`)
			if set.Source != policy.VCSPolicySetSource {
				fmt.Fprint(out, `<button class="btn btn-sm" type="submit">Save</button>`)
				fmt.Fprintf(out, `<button class="btn btn-sm" formaction="%s" type="submit">Delete</button>`, paths.DeletePolicy(p.ID))
			} else {
				fmt.Fprint(out, `<span class="text-sm opacity-70">Managed from VCS</span>`)
			}
			fmt.Fprint(out, `</div></form>`)
		}
		if len(policies) == 0 {
			fmt.Fprint(out, `<div class="text-sm opacity-70">No policies yet.</div>`)
		}
		fmt.Fprint(out, `</div></section></div>`)
		return nil
	})
	helpers.RenderPage(component, "policy set", w, r, helpers.WithOrganization(set.Organization), helpers.WithBreadcrumbs(
		helpers.Breadcrumb{Name: "Policy Sets", Link: paths.PolicySets(set.Organization)},
		helpers.Breadcrumb{Name: set.Name},
	))
}

func (h *Handlers) newManualPolicySet(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Organization organization.Name `schema:"organization_name,required"`
	}
	if err := decode.All(&params, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	component := templ.ComponentFunc(func(ctx context.Context, out io.Writer) error {
		fmt.Fprint(out, `<div class="flex flex-col gap-6">`)
		fmt.Fprint(out, `<div><h1 class="text-3xl font-bold">Create a manual policy set</h1><p class="opacity-70 mt-2">Create a policy set managed directly in OTF.</p></div>`)
		fmt.Fprintf(out, `<form class="border border-base-300 rounded-box p-4 grid grid-cols-1 lg:grid-cols-2 gap-4" method="POST" action="%s">`, paths.CreatePolicySet(params.Organization))
		fmt.Fprint(out, `<div class="flex flex-col gap-1"><label>Name</label><input class="input w-full" name="name" required></div>`)
		fmt.Fprint(out, `<div class="flex flex-col gap-1"><label>Description</label><input class="input w-full" name="description"></div>`)
		fmt.Fprint(out, `<div><button class="btn" type="submit">Create policy set</button></div></form></div>`)
		return nil
	})
	helpers.RenderPage(component, "new manual policy set", w, r, helpers.WithOrganization(params.Organization), helpers.WithBreadcrumbs(
		helpers.Breadcrumb{Name: "Policy Sets", Link: paths.PolicySets(params.Organization)},
		helpers.Breadcrumb{Name: "Manual"},
	))
}

func (h *Handlers) connectPolicySet(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Organization organization.Name `schema:"organization_name,required"`
	}
	if err := decode.All(&params, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	component := templ.ComponentFunc(func(ctx context.Context, out io.Writer) error {
		fmt.Fprint(out, `<div class="flex flex-col gap-6">`)
		fmt.Fprint(out, `<div><h1 class="text-3xl font-bold">Connect a new policy set</h1><p class="opacity-70 mt-2">Choose how this policy set will be managed.</p></div>`)
		fmt.Fprint(out, `<div class="grid grid-cols-1 lg:grid-cols-2 gap-4">`)
		fmt.Fprintf(out, `<a class="border border-base-300 rounded-box p-6 flex flex-col gap-3 hover:border-primary" href="%s"><h2 class="text-xl font-bold">Version control system (VCS)</h2><p class="opacity-70">Import policies from a repository and sync them manually when you want updates.</p></a>`, paths.ConnectVCSPolicySet(params.Organization))
		fmt.Fprintf(out, `<a class="border border-base-300 rounded-box p-6 flex flex-col gap-3 hover:border-primary" href="%s"><h2 class="text-xl font-bold">Individually managed policies</h2><p class="opacity-70">Create and edit Sentinel policies directly in OTF.</p></a>`, paths.NewManualPolicySet(params.Organization))
		fmt.Fprint(out, `</div></div>`)
		return nil
	})
	helpers.RenderPage(component, "connect policy set", w, r, helpers.WithOrganization(params.Organization), helpers.WithBreadcrumbs(
		helpers.Breadcrumb{Name: "Policy Sets", Link: paths.PolicySets(params.Organization)},
		helpers.Breadcrumb{Name: "Connect"},
	))
}

func (h *Handlers) newVCSPolicySet(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Organization organization.Name `schema:"organization_name,required"`
	}
	if err := decode.All(&params, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	providers, err := h.VCSProviders.ListVCSProviders(r.Context(), params.Organization)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}
	component := templ.ComponentFunc(func(ctx context.Context, out io.Writer) error {
		fmt.Fprint(out, `<div class="flex flex-col gap-6">`)
		fmt.Fprint(out, `<div><h1 class="text-3xl font-bold">Connect a new policy set</h1><p class="opacity-70 mt-2">Import Sentinel policies from a repository using <span class="font-mono">sentinel.hcl</span> as the manifest, then choose which declared policies belong to this policy set.</p></div>`)
		fmt.Fprint(out, `<div class="steps"><div class="step step-primary">Choose source</div><div class="step step-primary">Configure settings</div><div class="step">Select policies</div></div>`)
		fmt.Fprintf(out, `<form class="border border-base-300 rounded-box p-4 grid grid-cols-1 lg:grid-cols-2 gap-4" method="POST" action="%s">`, paths.PreviewVCSPolicySet(params.Organization))
		fmt.Fprint(out, `<div class="flex flex-col gap-1"><label>Name</label><input class="input w-full" name="name" required></div>`)
		fmt.Fprint(out, `<div class="flex flex-col gap-1"><label>Description</label><input class="input w-full" name="description"></div>`)
		fmt.Fprint(out, `<div class="flex flex-col gap-1"><label>VCS Provider</label><select class="select w-full" name="vcs_provider_id" required>`)
		fmt.Fprint(out, `<option value="">Select a provider</option>`)
		for _, provider := range providers {
			fmt.Fprintf(out, `<option value="%s">%s</option>`, provider.ID.String(), html.EscapeString(provider.String()))
		}
		fmt.Fprint(out, `</select></div>`)
		fmt.Fprint(out, `<div class="flex flex-col gap-1"><label>Repository</label><input class="input w-full font-mono" name="identifier" placeholder="owner/repo" required></div>`)
		fmt.Fprint(out, `<div class="flex flex-col gap-1"><label>Ref</label><input class="input w-full font-mono" name="vcs_ref" placeholder="main"></div>`)
		fmt.Fprint(out, `<div class="flex flex-col gap-1"><label>Policy path</label><input class="input w-full font-mono" name="vcs_path" placeholder="policies"></div>`)
		fmt.Fprint(out, `<div class="lg:col-span-2 text-sm opacity-70">Use <span class="font-mono">owner/repo</span> and an optional subdirectory. OTF expects a <span class="font-mono">sentinel.hcl</span> file in that path and imports only the <span class="font-mono">policy</span> blocks declared there. Helper modules stay attached to the policy set automatically.</div>`)
		fmt.Fprint(out, `<div><button class="btn" type="submit">Preview policies</button></div></form></div>`)
		return nil
	})
	helpers.RenderPage(component, "connect VCS policy set", w, r, helpers.WithOrganization(params.Organization), helpers.WithBreadcrumbs(
		helpers.Breadcrumb{Name: "Policy Sets", Link: paths.PolicySets(params.Organization)},
		helpers.Breadcrumb{Name: "Connect"},
	))
}

func (h *Handlers) previewVCSPolicySet(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Organization  organization.Name `schema:"organization_name,required"`
		Name          string            `schema:"name,required"`
		Description   string            `schema:"description"`
		VCSProviderID resource.TfeID    `schema:"vcs_provider_id,required"`
		Identifier    vcs.Repo          `schema:"identifier,required"`
		VCSRef        string            `schema:"vcs_ref"`
		VCSPath       string            `schema:"vcs_path"`
	}
	if err := decode.All(&params, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	provider, err := h.VCSProviders.GetVCSProvider(r.Context(), params.VCSProviderID)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}
	items, err := h.Policies.ListImportablePolicies(r.Context(), params.VCSProviderID, params.Identifier, params.VCSRef, params.VCSPath)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}
	component := templ.ComponentFunc(func(ctx context.Context, out io.Writer) error {
		fmt.Fprint(out, `<div class="flex flex-col gap-6">`)
		fmt.Fprint(out, `<div><h1 class="text-3xl font-bold">Connect a new policy set</h1><p class="opacity-70 mt-2">Select the Sentinel policies declared in <span class="font-mono">sentinel.hcl</span>. Imported helper modules are synced automatically.</p></div>`)
		fmt.Fprint(out, `<div class="steps"><div class="step step-primary">Choose source</div><div class="step step-primary">Configure settings</div><div class="step step-primary">Select policies</div></div>`)
		fmt.Fprintf(out, `<div class="border border-base-300 rounded-box p-4 text-sm opacity-80">Provider: <span class="font-mono">%s</span> | Repo: <span class="font-mono">%s</span>`, html.EscapeString(provider.String()), html.EscapeString(params.Identifier.String()))
		if params.VCSRef != "" {
			fmt.Fprintf(out, ` | Ref: <span class="font-mono">%s</span>`, html.EscapeString(params.VCSRef))
		}
		if params.VCSPath != "" {
			fmt.Fprintf(out, ` | Path: <span class="font-mono">%s</span>`, html.EscapeString(params.VCSPath))
		}
		fmt.Fprint(out, `</div>`)
		fmt.Fprintf(out, `<form class="flex flex-col gap-4" method="POST" action="%s">`, paths.CreateVCSPolicySet(params.Organization))
		fmt.Fprintf(out, `<input type="hidden" name="name" value="%s">`, html.EscapeString(params.Name))
		fmt.Fprintf(out, `<input type="hidden" name="description" value="%s">`, html.EscapeString(params.Description))
		fmt.Fprintf(out, `<input type="hidden" name="vcs_provider_id" value="%s">`, params.VCSProviderID.String())
		fmt.Fprintf(out, `<input type="hidden" name="identifier" value="%s">`, html.EscapeString(params.Identifier.String()))
		fmt.Fprintf(out, `<input type="hidden" name="vcs_ref" value="%s">`, html.EscapeString(params.VCSRef))
		fmt.Fprintf(out, `<input type="hidden" name="vcs_path" value="%s">`, html.EscapeString(params.VCSPath))
		if len(items) == 0 {
			fmt.Fprint(out, `<div class="alert alert-warning"><span>No Sentinel policy blocks were found in <span class="font-mono">sentinel.hcl</span> at that repository path.</span></div>`)
		} else {
			fmt.Fprint(out, `<div class="flex flex-col gap-3">`)
			for _, item := range items {
				fmt.Fprint(out, `<label class="border border-base-300 rounded-box p-4 flex flex-col gap-3">`)
				fmt.Fprint(out, `<div class="flex items-center gap-3">`)
				fmt.Fprintf(out, `<input class="checkbox" type="checkbox" name="selected_policy_paths" value="%s" checked>`, html.EscapeString(item.Path))
				fmt.Fprintf(out, `<div class="font-semibold">%s</div>`, html.EscapeString(item.Name))
				fmt.Fprintf(out, `<div class="text-sm opacity-70 font-mono">%s</div>`, html.EscapeString(item.Path))
				fmt.Fprint(out, `</div>`)
				fmt.Fprintf(out, `<textarea class="textarea textarea-sm w-full min-h-40 font-mono" readonly>%s</textarea>`, html.EscapeString(item.Source))
				fmt.Fprint(out, `</label>`)
			}
			fmt.Fprint(out, `</div>`)
			fmt.Fprint(out, `<div><button class="btn" type="submit">Create policy set</button></div>`)
		}
		fmt.Fprint(out, `</form></div>`)
		return nil
	})
	helpers.RenderPage(component, "preview VCS policies", w, r, helpers.WithOrganization(params.Organization), helpers.WithBreadcrumbs(
		helpers.Breadcrumb{Name: "Policy Sets", Link: paths.PolicySets(params.Organization)},
		helpers.Breadcrumb{Name: "Connect"},
	))
}

func (h *Handlers) createPolicySet(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Organization organization.Name `schema:"organization_name,required"`
		Name         string            `schema:"name,required"`
		Description  string            `schema:"description"`
	}
	if err := decode.All(&params, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	set, err := h.Policies.CreatePolicySet(r.Context(), params.Organization, policy.CreatePolicySetOptions{Name: &params.Name, Description: &params.Description})
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}
	http.Redirect(w, r, paths.PolicySet(set.ID), http.StatusFound)
}

func (h *Handlers) createVCSPolicySet(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Organization        organization.Name `schema:"organization_name,required"`
		Name                string            `schema:"name,required"`
		Description         string            `schema:"description"`
		VCSProviderID       resource.TfeID    `schema:"vcs_provider_id,required"`
		Identifier          vcs.Repo          `schema:"identifier,required"`
		VCSRef              string            `schema:"vcs_ref"`
		VCSPath             string            `schema:"vcs_path"`
		SelectedPolicyPaths []string          `schema:"selected_policy_paths"`
	}
	if err := decode.All(&params, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	set, _, err := h.Policies.CreateVCSPolicySet(r.Context(), params.Organization, policy.CreateVCSPolicySetOptions{
		Name:                &params.Name,
		Description:         &params.Description,
		VCSProviderID:       &params.VCSProviderID,
		VCSRepo:             &params.Identifier,
		VCSRef:              stringPtr(params.VCSRef),
		VCSPath:             stringPtr(params.VCSPath),
		SelectedPolicyPaths: params.SelectedPolicyPaths,
	})
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}
	http.Redirect(w, r, paths.PolicySet(set.ID), http.StatusFound)
}

func (h *Handlers) updatePolicySet(w http.ResponseWriter, r *http.Request) {
	var params struct {
		ID          resource.TfeID `schema:"policy_set_id,required"`
		Name        string         `schema:"name,required"`
		Description string         `schema:"description"`
	}
	if err := decode.All(&params, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	set, err := h.Policies.UpdatePolicySet(r.Context(), params.ID, policy.UpdatePolicySetOptions{Name: &params.Name, Description: &params.Description})
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}
	http.Redirect(w, r, paths.PolicySet(set.ID), http.StatusFound)
}

func (h *Handlers) deletePolicySet(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("policy_set_id", r)
	if err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	set, err := h.Policies.GetPolicySet(r.Context(), id)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}
	if err := h.Policies.DeletePolicySet(r.Context(), id); err != nil {
		helpers.Error(r, w, err.Error())
		return
	}
	http.Redirect(w, r, paths.PolicySets(set.Organization), http.StatusFound)
}

func (h *Handlers) syncPolicySet(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("policy_set_id", r)
	if err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	result, err := h.Policies.SyncPolicySetFromVCS(r.Context(), id)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}
	helpers.FlashSuccess(w, fmt.Sprintf("synced %d policies from VCS", len(result.Imported)))
	http.Redirect(w, r, paths.PolicySet(result.Set.ID), http.StatusFound)
}

func (h *Handlers) createPolicy(w http.ResponseWriter, r *http.Request) {
	var params struct {
		SetID            resource.TfeID          `schema:"policy_set_id,required"`
		Name             string                  `schema:"name,required"`
		Description      string                  `schema:"description"`
		EnforcementLevel policy.EnforcementLevel `schema:"enforcement_level,required"`
		Source           string                  `schema:"source,required"`
	}
	if err := decode.All(&params, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	set, err := h.Policies.GetPolicySet(r.Context(), params.SetID)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}
	_, err = h.Policies.CreatePolicy(r.Context(), params.SetID, policy.CreatePolicyOptions{
		Name: &params.Name, Description: &params.Description, EnforcementLevel: &params.EnforcementLevel, Source: &params.Source,
	})
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}
	http.Redirect(w, r, paths.PolicySet(set.ID), http.StatusFound)
}

func (h *Handlers) updatePolicy(w http.ResponseWriter, r *http.Request) {
	var params struct {
		ID               resource.TfeID          `schema:"policy_id,required"`
		Name             string                  `schema:"name,required"`
		Description      string                  `schema:"description"`
		EnforcementLevel policy.EnforcementLevel `schema:"enforcement_level,required"`
		Source           string                  `schema:"source,required"`
	}
	if err := decode.All(&params, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	updated, err := h.Policies.UpdatePolicy(r.Context(), params.ID, policy.UpdatePolicyOptions{
		Name: &params.Name, Description: &params.Description, EnforcementLevel: &params.EnforcementLevel, Source: &params.Source,
	})
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}
	http.Redirect(w, r, paths.PolicySet(updated.PolicySetID), http.StatusFound)
}

func (h *Handlers) deletePolicy(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("policy_id", r)
	if err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	p, err := h.Policies.GetPolicy(r.Context(), id)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}
	if err := h.Policies.DeletePolicy(r.Context(), id); err != nil {
		helpers.Error(r, w, err.Error())
		return
	}
	http.Redirect(w, r, paths.PolicySet(p.PolicySetID), http.StatusFound)
}

func (h *Handlers) workspaceSentinel(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	ws, err := h.Workspaces.GetWorkspace(r.Context(), workspaceID)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}
	sets, err := h.Policies.ListPolicySets(r.Context(), ws.Organization)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}
	component := templ.ComponentFunc(func(ctx context.Context, out io.Writer) error {
		fmt.Fprint(out, `<div class="flex flex-col gap-4">`)
		fmt.Fprint(out, `<div class="alert alert-info"><span>Sentinel policies are inherited from the organization. Any policy set defined for this organization applies to this workspace.</span></div>`)
		fmt.Fprintf(out, `<div><a class="btn" href="%s">Download Mock Bundle</a></div>`, paths.DownloadWorkspaceMocks(ws.ID))
		fmt.Fprint(out, `<div><h3 class="font-semibold">Organization Policy Sets</h3><ul class="list-disc pl-6">`)
		for _, set := range sets {
			fmt.Fprintf(out, `<li>%s</li>`, set.Name)
		}
		if len(sets) == 0 {
			fmt.Fprint(out, `<li>No policy sets are defined for this organization.</li>`)
		}
		fmt.Fprint(out, `</ul></div></div>`)
		return nil
	})
	helpers.RenderPage(component, "workspace sentinel", w, r, helpers.WithWorkspace(ws, h.Authorizer), helpers.WithBreadcrumbs(helpers.Breadcrumb{Name: "Sentinel"}))
}

func (h *Handlers) downloadWorkspaceMocks(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	var runID *resource.TfeID
	if runIDParam := r.URL.Query().Get("run_id"); runIDParam != "" {
		parsed, err := resource.ParseTfeID(runIDParam)
		if err != nil {
			helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
			return
		}
		runID = &parsed
	}
	if runID == nil {
		ws, err := h.Workspaces.GetWorkspace(r.Context(), workspaceID)
		if err != nil {
			helpers.Error(r, w, err.Error())
			return
		}
		if ws.LatestRun != nil {
			runID = &ws.LatestRun.ID
		}
	}
	bundle, err := h.Policies.GenerateWorkspaceMocks(r.Context(), workspaceID, runID)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s-sentinel-mocks.zip"`, workspaceID))
	_, _ = w.Write(bundle)
}

func selectedAttr(ok bool) string {
	if ok {
		return " selected"
	}
	return ""
}

func stringPtr(s string) *string {
	return &s
}

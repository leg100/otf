package policy

import (
	"encoding/json"
	"net/http"

	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/tfeapi"
)

func (a *api) createPolicySet(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Organization organization.Name `schema:"organization_name,required"`
	}
	if err := decode.All(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}
	var opts CreatePolicySetOptions
	if err := json.NewDecoder(r.Body).Decode(&opts); err != nil {
		tfeapi.Error(w, err)
		return
	}
	set, err := a.CreatePolicySet(r.Context(), params.Organization, opts)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	json.NewEncoder(w).Encode(set)
}

func (a *api) listPolicySets(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Organization organization.Name `schema:"organization_name,required"`
	}
	if err := decode.All(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}
	sets, err := a.ListPolicySets(r.Context(), params.Organization)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	json.NewEncoder(w).Encode(sets)
}

func (a *api) updatePolicySet(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("policy_set_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	var opts UpdatePolicySetOptions
	if err := json.NewDecoder(r.Body).Decode(&opts); err != nil {
		tfeapi.Error(w, err)
		return
	}
	set, err := a.UpdatePolicySet(r.Context(), id, opts)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	json.NewEncoder(w).Encode(set)
}

func (a *api) deletePolicySet(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("policy_set_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	if err := a.DeletePolicySet(r.Context(), id); err != nil {
		tfeapi.Error(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *api) createPolicy(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("policy_set_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	var opts CreatePolicyOptions
	if err := json.NewDecoder(r.Body).Decode(&opts); err != nil {
		tfeapi.Error(w, err)
		return
	}
	policy, err := a.CreatePolicy(r.Context(), id, opts)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	json.NewEncoder(w).Encode(policy)
}

func (a *api) listPolicies(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("policy_set_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	policies, err := a.ListPolicies(r.Context(), id)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	json.NewEncoder(w).Encode(policies)
}

func (a *api) updatePolicy(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("policy_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	var opts UpdatePolicyOptions
	if err := json.NewDecoder(r.Body).Decode(&opts); err != nil {
		tfeapi.Error(w, err)
		return
	}
	policy, err := a.UpdatePolicy(r.Context(), id, opts)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	json.NewEncoder(w).Encode(policy)
}

func (a *api) deletePolicy(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("policy_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	if err := a.DeletePolicy(r.Context(), id); err != nil {
		tfeapi.Error(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *api) setPolicySetWorkspaces(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("policy_set_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	var payload struct {
		WorkspaceIDs []resource.TfeID `json:"workspace_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		tfeapi.Error(w, err)
		return
	}
	if err := a.SetPolicySetWorkspaces(r.Context(), id, payload.WorkspaceIDs); err != nil {
		tfeapi.Error(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *api) listPolicyChecks(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("run_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	checks, err := a.ListPolicyChecks(r.Context(), id)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	json.NewEncoder(w).Encode(checks)
}

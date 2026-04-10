package policy

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gorilla/mux"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/repohooks"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/vcs"
)

type Service struct {
	logr.Logger
	*authz.Authorizer

	db         *pgdb
	api        *api
	runs       runClient
	workspaces workspaceClient
	configs    configClient
	vcs        vcsProviderClient
	repohooks  repoHookClient
	states     stateClient
	variables  variableClient
	plans      planClient
	orgs       organizationClient
	evaluators map[Kind]Evaluator
}

type Options struct {
	Logger              logr.Logger
	Authorizer          *authz.Authorizer
	DB                  *sql.DB
	Runs                runClient
	Organizations       organizationClient
	Workspaces          workspaceClient
	Configs             configClient
	VCSProviders        vcsProviderClient
	RepoHooks           repoHookClient
	States              stateClient
	Variables           variableClient
	Plans               planClient
	VCSEventSub         vcs.Subscriber
	PolicyEngineBinDir  string
	PolicyEngineWorkDir string
}

type organizationClient interface {
	GetOrganization(context.Context, organization.Name) (*organization.Organization, error)
}

func NewService(opts Options) *Service {
	resolver, err := newCLIResolver(opts.Logger, opts.PolicyEngineBinDir)
	if err != nil {
		panic(err)
	}
	svc := &Service{
		Logger:     opts.Logger,
		Authorizer: opts.Authorizer,
		db:         &pgdb{opts.DB},
		runs:       opts.Runs,
		workspaces: opts.Workspaces,
		configs:    opts.Configs,
		vcs:        opts.VCSProviders,
		repohooks:  opts.RepoHooks,
		states:     opts.States,
		variables:  opts.Variables,
		plans:      opts.Plans,
		orgs:       opts.Organizations,
		evaluators: map[Kind]Evaluator{
			SentinelKind: &sentinelEvaluator{
				logger:   opts.Logger,
				resolver: resolver,
				workDir:  opts.PolicyEngineWorkDir,
			},
			OPAKind: &opaEvaluator{},
		},
	}
	svc.api = &api{Service: svc}
	opts.Authorizer.RegisterParentResolver(resource.PolicySetKind, func(ctx context.Context, id resource.ID) (resource.ID, error) {
		return svc.db.getPolicySetOrganization(ctx, id.(resource.TfeID))
	})
	opts.Authorizer.RegisterParentResolver(resource.PolicyKind, func(ctx context.Context, id resource.ID) (resource.ID, error) {
		return svc.db.getPolicySetIDByPolicy(ctx, id.(resource.TfeID))
	})
	opts.Authorizer.RegisterParentResolver(resource.PolicyCheckKind, func(ctx context.Context, id resource.ID) (resource.ID, error) {
		return svc.db.getRunIDByPolicyCheck(ctx, id.(resource.TfeID))
	})
	if opts.VCSEventSub != nil {
		opts.VCSEventSub.Subscribe(svc.handleVCSEvent)
	}
	return svc
}

func (s *Service) AddHandlers(r *mux.Router) {
	s.api.addHandlers(r)
}

func (s *Service) CreatePolicySet(ctx context.Context, org organization.Name, opts CreatePolicySetOptions) (*PolicySet, error) {
	subject, err := s.Authorize(ctx, authz.CreatePolicySetAction, org)
	if err != nil {
		return nil, err
	}
	set, err := newPolicySet(org, opts)
	if err != nil {
		return nil, err
	}
	if err := s.db.createPolicySet(ctx, set); err != nil {
		s.Error(err, "creating policy set", "subject", subject, "organization", org)
		return nil, err
	}
	return set, nil
}

func (s *Service) ListPolicySets(ctx context.Context, org organization.Name) ([]*PolicySet, error) {
	if _, err := s.Authorize(ctx, authz.ListPolicySetsAction, org); err != nil {
		return nil, err
	}
	return s.db.listPolicySets(ctx, org)
}

func (s *Service) GetPolicySet(ctx context.Context, id resource.TfeID) (*PolicySet, error) {
	if _, err := s.Authorize(ctx, authz.GetPolicySetAction, id); err != nil {
		return nil, err
	}
	return s.db.getPolicySet(ctx, id)
}

func (s *Service) UpdatePolicySet(ctx context.Context, id resource.TfeID, opts UpdatePolicySetOptions) (*PolicySet, error) {
	if _, err := s.Authorize(ctx, authz.UpdatePolicySetAction, id); err != nil {
		return nil, err
	}
	return s.db.updatePolicySet(ctx, id, func(ctx context.Context, set *PolicySet) error {
		if opts.Name != nil {
			if err := resource.ValidateName(opts.Name); err != nil {
				return err
			}
			set.Name = *opts.Name
		}
		if opts.Description != nil {
			set.Description = *opts.Description
		}
		if opts.Kind != nil {
			if !opts.Kind.Valid() {
				return fmt.Errorf("invalid policy engine kind: %s", *opts.Kind)
			}
			if set.Source == VCSPolicySetSource && *opts.Kind != SentinelKind {
				return fmt.Errorf("policy engine %q does not support VCS imports yet", *opts.Kind)
			}
			set.Kind = *opts.Kind
		}
		if opts.EngineVersion != nil {
			set.EngineVersion = *opts.EngineVersion
		}
		set.UpdatedAt = internal.CurrentTimestamp(nil)
		return nil
	})
}

func (s *Service) DeletePolicySet(ctx context.Context, id resource.TfeID) error {
	if _, err := s.Authorize(ctx, authz.DeletePolicySetAction, id); err != nil {
		return err
	}
	return s.db.Tx(ctx, func(ctx context.Context) error {
		if err := s.db.deletePolicySet(ctx, id); err != nil {
			return err
		}
		if s.repohooks != nil {
			if err := s.repohooks.DeleteUnreferencedRepohooks(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}

type repoPolicyBundle struct {
	Policies []ImportablePolicy
	Modules  []ImportableModule
}

func (s *Service) ListImportablePolicies(ctx context.Context, providerID resource.TfeID, repo vcs.Repo, ref, subpath string) ([]ImportablePolicy, error) {
	bundle, err := s.loadRepoPolicyBundle(ctx, SentinelKind, providerID, repo, ref, subpath)
	if err != nil {
		return nil, err
	}
	return bundle.Policies, nil
}

func (s *Service) CreateVCSPolicySet(ctx context.Context, org organization.Name, opts CreateVCSPolicySetOptions) (*PolicySet, []*Policy, error) {
	subject, err := s.Authorize(ctx, authz.CreatePolicySetAction, org)
	if err != nil {
		return nil, nil, err
	}
	if opts.VCSProviderID == nil || opts.VCSRepo == nil {
		return nil, nil, fmt.Errorf("missing VCS provider or repository")
	}
	kind := SentinelKind
	if opts.Kind != nil {
		kind = *opts.Kind
	}
	if !kind.Valid() {
		return nil, nil, fmt.Errorf("invalid policy engine kind: %s", kind)
	}
	bundle, err := s.loadRepoPolicyBundle(ctx, kind, *opts.VCSProviderID, *opts.VCSRepo, derefString(opts.VCSRef), derefString(opts.VCSPath))
	if err != nil {
		return nil, nil, err
	}
	selected := map[string]ImportablePolicy{}
	for _, item := range bundle.Policies {
		selected[item.Path] = item
	}
	if len(opts.SelectedPolicyPaths) == 0 {
		return nil, nil, fmt.Errorf("select at least one policy")
	}
	now := internal.CurrentTimestamp(nil)
	set, err := newPolicySet(org, CreatePolicySetOptions{
		Name:           opts.Name,
		Description:    opts.Description,
		Kind:           opts.Kind,
		EngineVersion:  opts.EngineVersion,
		Source:         VCSPolicySetSource,
		VCSProviderID:  opts.VCSProviderID,
		VCSRepo:        opts.VCSRepo,
		VCSRef:         opts.VCSRef,
		VCSPath:        opts.VCSPath,
		VCSPolicyPaths: append([]string(nil), opts.SelectedPolicyPaths...),
		LastSyncedAt:   &now,
	})
	if err != nil {
		return nil, nil, err
	}
	policies := make([]*Policy, 0, len(opts.SelectedPolicyPaths))
	modules := make([]*PolicyModule, 0, len(bundle.Modules))
	for _, mod := range bundle.Modules {
		modules = append(modules, &PolicyModule{
			ID:          resource.NewTfeID(resource.PolicyKind),
			PolicySetID: set.ID,
			CreatedAt:   now,
			UpdatedAt:   now,
			Name:        mod.Name,
			Path:        mod.Path,
			Source:      mod.Source,
		})
	}
	for _, path := range opts.SelectedPolicyPaths {
		item, ok := selected[path]
		if !ok {
			return nil, nil, fmt.Errorf("selected policy not found in repo: %s", path)
		}
		level := item.EnforcementLevel
		if level == "" {
			level = MandatoryEnforcement
		}
		src := item.Source
		p, err := newPolicy(set, CreatePolicyOptions{
			Name:             stringPtr(item.Name),
			Description:      stringPtr("Imported from " + path),
			EnforcementLevel: &level,
			Source:           &src,
			Path:             &item.Path,
		})
		if err != nil {
			return nil, nil, err
		}
		policies = append(policies, p)
	}
	err = s.db.Tx(ctx, func(ctx context.Context) error {
		if s.repohooks != nil {
			if _, err := s.repohooks.CreateRepohook(ctx, repohooks.CreateRepohookOptions{
				VCSProviderID: *opts.VCSProviderID,
				RepoPath:      *opts.VCSRepo,
			}); err != nil {
				return fmt.Errorf("creating webhook: %w", err)
			}
		}
		if err := s.db.createPolicySet(ctx, set); err != nil {
			return err
		}
		for _, mod := range modules {
			if err := s.db.createPolicyModule(ctx, mod); err != nil {
				return err
			}
		}
		for _, p := range policies {
			if err := s.db.createPolicy(ctx, p); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		s.Error(err, "creating VCS policy set", "subject", subject, "organization", org)
		return nil, nil, err
	}
	return set, policies, nil
}

func (s *Service) SyncPolicySetFromVCS(ctx context.Context, setID resource.TfeID) (*SyncResult, error) {
	if _, err := s.Authorize(ctx, authz.UpdatePolicySetAction, setID); err != nil {
		return nil, err
	}
	set, err := s.db.getPolicySet(ctx, setID)
	if err != nil {
		return nil, err
	}
	if set.Source != VCSPolicySetSource || set.VCSProviderID == nil || set.VCSRepo == nil {
		return nil, fmt.Errorf("policy set is not backed by VCS")
	}
	bundle, err := s.loadRepoPolicyBundle(ctx, set.Kind, *set.VCSProviderID, *set.VCSRepo, set.VCSRef, set.VCSPath)
	if err != nil {
		return nil, err
	}
	importedByPath := map[string]ImportablePolicy{}
	for _, item := range bundle.Policies {
		importedByPath[item.Path] = item
	}
	existing, err := s.db.listPolicies(ctx, setID)
	if err != nil {
		return nil, err
	}
	existingModules, err := s.db.listPolicyModules(ctx, setID)
	if err != nil {
		return nil, err
	}
	existingByPath := map[string]*Policy{}
	for _, p := range existing {
		existingByPath[p.Path] = p
	}
	existingModulesByPath := map[string]*PolicyModule{}
	for _, mod := range existingModules {
		existingModulesByPath[mod.Path] = mod
	}
	var imported []*Policy
	err = s.db.Tx(ctx, func(ctx context.Context) error {
		seenModules := map[string]struct{}{}
		for _, item := range bundle.Modules {
			seenModules[item.Path] = struct{}{}
			if current, ok := existingModulesByPath[item.Path]; ok {
				_, err := s.db.updatePolicyModule(ctx, current.ID, func(ctx context.Context, mod *PolicyModule) error {
					mod.Name = item.Name
					mod.Path = item.Path
					mod.Source = item.Source
					mod.UpdatedAt = internal.CurrentTimestamp(nil)
					return nil
				})
				if err != nil {
					return err
				}
				continue
			}
			now := internal.CurrentTimestamp(nil)
			if err := s.db.createPolicyModule(ctx, &PolicyModule{
				ID:          resource.NewTfeID(resource.PolicyKind),
				PolicySetID: setID,
				CreatedAt:   now,
				UpdatedAt:   now,
				Name:        item.Name,
				Path:        item.Path,
				Source:      item.Source,
			}); err != nil {
				return err
			}
		}
		for _, mod := range existingModules {
			if _, keep := seenModules[mod.Path]; keep {
				continue
			}
			if err := s.db.deletePolicyModule(ctx, mod.ID); err != nil {
				return err
			}
		}
		seen := map[string]struct{}{}
		for _, selectedPath := range set.VCSPolicyPaths {
			item, ok := importedByPath[selectedPath]
			if !ok {
				continue
			}
			seen[selectedPath] = struct{}{}
			if current, ok := existingByPath[selectedPath]; ok {
				_, err := s.db.updatePolicy(ctx, current.ID, func(ctx context.Context, p *Policy) error {
					p.Name = item.Name
					p.Source = item.Source
					p.Path = item.Path
					if p.Description == "" || strings.HasPrefix(p.Description, "Imported from ") {
						p.Description = "Imported from " + item.Path
					}
					p.UpdatedAt = internal.CurrentTimestamp(nil)
					return nil
				})
				if err != nil {
					return err
				}
				updated, err := s.db.getPolicy(ctx, current.ID)
				if err != nil {
					return err
				}
				imported = append(imported, updated)
				continue
			}
			level := item.EnforcementLevel
			if level == "" {
				level = MandatoryEnforcement
			}
			src := item.Source
			p, err := newPolicy(set, CreatePolicyOptions{
				Name:             stringPtr(item.Name),
				Description:      stringPtr("Imported from " + item.Path),
				EnforcementLevel: &level,
				Source:           &src,
				Path:             &item.Path,
			})
			if err != nil {
				return err
			}
			if err := s.db.createPolicy(ctx, p); err != nil {
				return err
			}
			imported = append(imported, p)
		}
		for _, p := range existing {
			if _, keep := seen[p.Path]; keep {
				continue
			}
			if err := s.db.deletePolicy(ctx, p.ID); err != nil {
				return err
			}
		}
		now := internal.CurrentTimestamp(nil)
		_, err := s.db.updatePolicySet(ctx, setID, func(ctx context.Context, pset *PolicySet) error {
			pset.LastSyncedAt = &now
			pset.UpdatedAt = now
			return nil
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	updatedSet, err := s.db.getPolicySet(ctx, setID)
	if err != nil {
		return nil, err
	}
	return &SyncResult{Set: updatedSet, Imported: imported}, nil
}

func (s *Service) loadRepoPolicyBundle(ctx context.Context, kind Kind, providerID resource.TfeID, repo vcs.Repo, ref, subpath string) (*repoPolicyBundle, error) {
	if kind != SentinelKind {
		return nil, fmt.Errorf("policy engine %q does not support VCS imports yet", kind)
	}
	provider, err := s.vcs.GetVCSProvider(ctx, providerID)
	if err != nil {
		return nil, err
	}
	tarball, _, err := provider.GetRepoTarball(ctx, vcs.GetRepoTarballOptions{
		Repo: repo,
		Ref:  optionalStringPtr(ref),
	})
	if err != nil {
		return nil, err
	}
	root, err := os.MkdirTemp("", "otf-policy-import-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(root)
	if err := internal.Unpack(bytes.NewReader(tarball), root); err != nil {
		return nil, err
	}
	searchRoot := root
	if subpath != "" {
		searchRoot = filepath.Join(root, filepath.Clean(subpath))
	}
	info, err := os.Stat(searchRoot)
	if err != nil {
		return nil, fmt.Errorf("opening policy path: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("policy path must be a directory")
	}
	configPath := filepath.Join(searchRoot, "sentinel.hcl")
	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("reading sentinel.hcl: %w", err)
	}
	parser := hclparse.NewParser()
	file, diags := parser.ParseHCL(configBytes, configPath)
	if diags.HasErrors() {
		return nil, fmt.Errorf("parsing sentinel.hcl: %s", diags.Error())
	}
	schema := &hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{Type: "import", LabelNames: []string{"kind", "name"}},
			{Type: "policy", LabelNames: []string{"name"}},
		},
	}
	content, diags := file.Body.Content(schema)
	if diags.HasErrors() {
		return nil, fmt.Errorf("decoding sentinel.hcl: %s", diags.Error())
	}
	bundle := &repoPolicyBundle{}
	for _, block := range content.Blocks {
		switch block.Type {
		case "import":
			if len(block.Labels) != 2 || block.Labels[0] != "module" {
				continue
			}
			sourcePath, err := sentinelStringAttribute(block.Body, "source")
			if err != nil {
				return nil, fmt.Errorf("import %q: %w", block.Labels[1], err)
			}
			source, rel, err := readRepoFile(searchRoot, sourcePath)
			if err != nil {
				return nil, err
			}
			bundle.Modules = append(bundle.Modules, ImportableModule{
				Name:   block.Labels[1],
				Path:   rel,
				Source: source,
			})
		case "policy":
			if len(block.Labels) != 1 {
				continue
			}
			sourcePath, err := sentinelStringAttribute(block.Body, "source")
			if err != nil {
				return nil, fmt.Errorf("policy %q: %w", block.Labels[0], err)
			}
			source, rel, err := readRepoFile(searchRoot, sourcePath)
			if err != nil {
				return nil, err
			}
			levelText, err := sentinelOptionalStringAttribute(block.Body, "enforcement_level")
			if err != nil {
				return nil, fmt.Errorf("policy %q: %w", block.Labels[0], err)
			}
			bundle.Policies = append(bundle.Policies, ImportablePolicy{
				Name:             block.Labels[0],
				Path:             rel,
				Source:           source,
				EnforcementLevel: enforcementLevelFromSentinel(levelText),
			})
		}
	}
	sort.Slice(bundle.Policies, func(i, j int) bool { return bundle.Policies[i].Path < bundle.Policies[j].Path })
	sort.Slice(bundle.Modules, func(i, j int) bool { return bundle.Modules[i].Path < bundle.Modules[j].Path })
	return bundle, nil
}

func (s *Service) CreatePolicy(ctx context.Context, setID resource.TfeID, opts CreatePolicyOptions) (*Policy, error) {
	set, err := s.GetPolicySet(ctx, setID)
	if err != nil {
		return nil, err
	}
	if set.Source == VCSPolicySetSource {
		return nil, fmt.Errorf("cannot manually add policies to a VCS-backed policy set")
	}
	if _, err := s.Authorize(ctx, authz.CreatePolicyAction, setID); err != nil {
		return nil, err
	}
	policy, err := newPolicy(set, opts)
	if err != nil {
		return nil, err
	}
	if err := s.db.createPolicy(ctx, policy); err != nil {
		return nil, err
	}
	return policy, nil
}

func (s *Service) ListPolicies(ctx context.Context, setID resource.TfeID) ([]*Policy, error) {
	if _, err := s.Authorize(ctx, authz.GetPolicySetAction, setID); err != nil {
		return nil, err
	}
	return s.db.listPolicies(ctx, setID)
}

func (s *Service) GetPolicy(ctx context.Context, id resource.TfeID) (*Policy, error) {
	if _, err := s.Authorize(ctx, authz.GetPolicyAction, id); err != nil {
		return nil, err
	}
	return s.db.getPolicy(ctx, id)
}

func (s *Service) UpdatePolicy(ctx context.Context, id resource.TfeID, opts UpdatePolicyOptions) (*Policy, error) {
	if _, err := s.Authorize(ctx, authz.UpdatePolicyAction, id); err != nil {
		return nil, err
	}
	return s.db.updatePolicy(ctx, id, func(ctx context.Context, policy *Policy) error {
		if opts.Name != nil {
			if err := resource.ValidateName(opts.Name); err != nil {
				return err
			}
			policy.Name = *opts.Name
		}
		if opts.Description != nil {
			policy.Description = *opts.Description
		}
		if opts.EnforcementLevel != nil {
			policy.EnforcementLevel = *opts.EnforcementLevel
		}
		if opts.Source != nil {
			policy.Source = *opts.Source
		}
		policy.UpdatedAt = internal.CurrentTimestamp(nil)
		return nil
	})
}

func (s *Service) DeletePolicy(ctx context.Context, id resource.TfeID) error {
	if _, err := s.Authorize(ctx, authz.DeletePolicyAction, id); err != nil {
		return err
	}
	return s.db.deletePolicy(ctx, id)
}

func (s *Service) SetPolicySetWorkspaces(ctx context.Context, setID resource.TfeID, workspaceIDs []resource.TfeID) error {
	// Deprecated: policy sets are org-wide by default in v1. Keep this method
	// as a harmless no-op for compatibility with any older callers.
	if _, err := s.Authorize(ctx, authz.AttachPolicySetAction, setID); err != nil {
		return err
	}
	return nil
}

func (s *Service) ListPolicySetWorkspaces(ctx context.Context, setID resource.TfeID) ([]resource.TfeID, error) {
	// Deprecated: policy sets are org-wide by default in v1.
	if _, err := s.Authorize(ctx, authz.GetPolicySetAction, setID); err != nil {
		return nil, err
	}
	return []resource.TfeID{}, nil
}

func (s *Service) ListPolicyChecks(ctx context.Context, runID resource.TfeID) ([]*PolicyCheck, error) {
	if _, err := s.Authorize(ctx, authz.ListPolicyChecksAction, runID); err != nil {
		return nil, err
	}
	return s.db.listPolicyChecks(ctx, runID)
}

func (s *Service) CountWorkspaceFailures(ctx context.Context, workspaceID resource.TfeID) (int, error) {
	if _, err := s.Authorize(ctx, authz.GetWorkspaceAction, workspaceID); err != nil {
		return 0, err
	}
	return s.db.countWorkspaceFailures(ctx, workspaceID)
}

func (s *Service) BuildMockBundle(ctx context.Context, workspaceID resource.TfeID, runID *resource.TfeID) (*MockBundle, error) {
	// Policy evaluation runs as an internal system workflow after planning, so
	// it must not depend on an end-user subject being present on the context.
	internalCtx := authz.AddSkipAuthz(authz.AddSubjectToContext(ctx, &authz.Superuser{Username: "sentinel-policy"}))

	files := map[string][]byte{}
	missing := []string{}

	var (
		runInfo  *RunInfo
		planData []byte
	)

	state, err := s.states.DownloadCurrentState(internalCtx, workspaceID)
	if err != nil {
		if errors.Is(err, internal.ErrResourceNotFound) {
			missing = append(missing, "state.json")
			files["state.json"] = marshalJSON(map[string]any{})
		} else {
			return nil, fmt.Errorf("building state mock: %w", err)
		}
	} else {
		files["state.json"] = state
	}

	if runID != nil {
		if s.runs != nil {
			runInfo, err = s.runs.GetRunInfo(internalCtx, *runID)
			if err != nil {
				return nil, fmt.Errorf("building tfrun mock: %w", err)
			}
			files["tfrun.json"] = marshalJSON(runInfo)
		} else {
			missing = append(missing, "tfrun.json")
		}

		vars, err := s.variables.ListRunVariables(internalCtx, *runID)
		if err != nil {
			if errors.Is(err, internal.ErrResourceNotFound) {
				missing = append(missing, "variables.json")
				files["variables.json"] = marshalJSON([]Variable{})
			} else {
				return nil, fmt.Errorf("building variables mock: %w", err)
			}
		} else {
			files["variables.json"] = marshalJSON(vars)
			if runInfo != nil {
				runInfo.Variables = make(map[string]RunVariableInfo, len(vars))
				for _, v := range vars {
					runInfo.Variables[v.Key] = RunVariableInfo{
						Category:  v.Category,
						Sensitive: v.Sensitive,
					}
				}
				files["tfrun.json"] = marshalJSON(runInfo)
			}
		}

		planData, err = s.plans.GetRunPlanJSON(internalCtx, *runID)
		if err != nil {
			if errors.Is(err, internal.ErrResourceNotFound) {
				missing = append(missing, "plan.json")
				files["plan.json"] = marshalJSON(map[string]any{})
			} else {
				return nil, fmt.Errorf("building plan mock: %w", err)
			}
		} else {
			files["plan.json"] = planData
		}
	} else {
		missing = append(missing, "tfrun.json", "variables.json", "plan.json")
		files["variables.json"] = marshalJSON([]Variable{})
		files["plan.json"] = marshalJSON(map[string]any{})
	}

	mockFiles, err := s.buildCanonicalMockFiles(internalCtx, runInfo, files["variables.json"], files["plan.json"], files["state.json"], missing)
	if err != nil {
		return nil, fmt.Errorf("building sentinel mock bundle: %w", err)
	}
	for name, data := range mockFiles {
		files[name] = data
	}

	files["README.json"] = marshalJSON(map[string]any{
		"generated_at": internal.CurrentTimestamp(nil),
		"workspace_id": workspaceID,
		"run_id":       runID,
		"missing":      missing,
	})
	return &MockBundle{Files: files}, nil
}

func (s *Service) GenerateWorkspaceMocks(ctx context.Context, workspaceID resource.TfeID, runID *resource.TfeID) ([]byte, error) {
	if _, err := s.Authorize(ctx, authz.DownloadWorkspaceMocksAction, workspaceID); err != nil {
		return nil, err
	}
	bundle, err := s.BuildMockBundle(ctx, workspaceID, runID)
	if err != nil {
		return nil, err
	}
	return bundle.Zip()
}

func (s *Service) buildCanonicalMockFiles(ctx context.Context, runInfo *RunInfo, variablesJSON, planJSON, stateJSON []byte, missing []string) (map[string][]byte, error) {
	files := map[string][]byte{
		"sentinel.hcl": []byte(sentinelConfig()),
	}

	tfrunSource := ""
	if runInfo != nil {
		var err error
		tfrunSource, err = sentinelModuleFromJSON(marshalJSON(runInfo))
		if err != nil {
			return nil, fmt.Errorf("converting tfrun mock: %w", err)
		}
	}
	files["mock-tfrun.sentinel"] = []byte(tfrunSource)

	tfplanSource, err := sentinelModuleFromJSON(planJSON)
	if err != nil {
		return nil, fmt.Errorf("converting tfplan mock: %w", err)
	}
	files["mock-tfplan.sentinel"] = []byte(tfplanSource)
	files["mock-tfplan-v2.sentinel"] = []byte(tfplanSource)

	tfstateValue, err := buildTFStateImport(planJSON, stateJSON)
	if err != nil {
		return nil, fmt.Errorf("converting tfstate mock: %w", err)
	}
	tfstateSource, err := sentinelModuleFromValue(tfstateValue)
	if err != nil {
		return nil, fmt.Errorf("rendering tfstate mock: %w", err)
	}
	files["mock-tfstate.sentinel"] = []byte(tfstateSource)
	files["mock-tfstate-v2.sentinel"] = []byte(tfstateSource)

	tfconfigValue, err := s.buildTFConfigImport(ctx, runInfo)
	if err != nil {
		return nil, fmt.Errorf("building tfconfig mock: %w", err)
	}
	tfconfigSource, err := sentinelModuleFromValue(tfconfigValue)
	if err != nil {
		return nil, fmt.Errorf("rendering tfconfig mock: %w", err)
	}
	files["mock-tfconfig.sentinel"] = []byte(tfconfigSource)
	files["mock-tfconfig-v2.sentinel"] = []byte(tfconfigSource)

	files["README.json"] = marshalJSON(map[string]any{
		"format": "hashicorp-sentinel-test-bundle",
		"includes": []string{
			"mock-tfrun.sentinel",
			"mock-tfconfig.sentinel",
			"mock-tfconfig-v2.sentinel",
			"mock-tfplan.sentinel",
			"mock-tfplan-v2.sentinel",
			"mock-tfstate.sentinel",
			"mock-tfstate-v2.sentinel",
			"sentinel.hcl",
		},
		"missing_inputs": missing,
	})

	_ = variablesJSON

	return files, nil
}

func (s *Service) buildTFConfigImport(ctx context.Context, runInfo *RunInfo) (any, error) {
	if runInfo == nil || runInfo.ConfigurationVersionID.IsZero() || s.configs == nil {
		return map[string]any{}, nil
	}
	tarball, err := s.configs.DownloadConfig(ctx, runInfo.ConfigurationVersionID)
	if err != nil {
		if errors.Is(err, internal.ErrResourceNotFound) {
			return map[string]any{}, nil
		}
		return nil, err
	}

	dir, err := os.MkdirTemp("", "otf-tfconfig-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)

	if err := internal.Unpack(bytes.NewReader(tarball), dir); err != nil {
		return nil, err
	}

	moduleDir := dir
	if wd := runInfo.Workspace.WorkingDirectory; wd != "" {
		moduleDir = filepath.Join(dir, wd)
	}
	mod, diags := tfconfig.LoadModule(moduleDir)
	if diags.HasErrors() {
		return nil, fmt.Errorf("parsing terraform configuration: %s", diags.Error())
	}

	var v any
	raw := marshalJSON(mod)
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, err
	}
	return v, nil
}

func buildTFStateImport(planJSON, stateJSON []byte) (any, error) {
	var plan map[string]any
	if len(planJSON) > 0 {
		if err := json.Unmarshal(planJSON, &plan); err == nil {
			if prior, ok := plan["prior_state"]; ok {
				return prior, nil
			}
		}
	}

	var state any
	if err := json.Unmarshal(stateJSON, &state); err != nil {
		return map[string]any{}, nil
	}
	return state, nil
}

func stringPtr(s string) *string {
	return &s
}

func optionalStringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func sentinelStringAttribute(body hcl.Body, name string) (string, error) {
	attrs, diags := body.JustAttributes()
	if diags.HasErrors() {
		return "", fmt.Errorf("reading attributes: %s", diags.Error())
	}
	attr, ok := attrs[name]
	if !ok {
		return "", fmt.Errorf("missing %q", name)
	}
	value, diags := attr.Expr.Value(nil)
	if diags.HasErrors() {
		return "", fmt.Errorf("evaluating %q: %s", name, diags.Error())
	}
	return value.AsString(), nil
}

func sentinelOptionalStringAttribute(body hcl.Body, name string) (string, error) {
	attrs, diags := body.JustAttributes()
	if diags.HasErrors() {
		return "", fmt.Errorf("reading attributes: %s", diags.Error())
	}
	attr, ok := attrs[name]
	if !ok {
		return "", nil
	}
	value, diags := attr.Expr.Value(nil)
	if diags.HasErrors() {
		return "", fmt.Errorf("evaluating %q: %s", name, diags.Error())
	}
	return value.AsString(), nil
}

func readRepoFile(root, sourcePath string) (string, string, error) {
	clean := filepath.Clean(sourcePath)
	path := filepath.Join(root, clean)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", "", fmt.Errorf("reading %s: %w", sourcePath, err)
	}
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return "", "", err
	}
	return string(data), filepath.ToSlash(rel), nil
}

func enforcementLevelFromSentinel(level string) EnforcementLevel {
	switch level {
	case "advisory":
		return AdvisoryEnforcement
	case "soft-mandatory", "hard-mandatory", "mandatory":
		return MandatoryEnforcement
	default:
		return MandatoryEnforcement
	}
}

func (s *Service) EvaluateRun(ctx context.Context, runID, workspaceID resource.TfeID, org organization.Name) (EvaluationResult, error) {
	policies, err := s.db.listPoliciesByOrganization(ctx, org)
	if err != nil {
		return EvaluationResult{}, fmt.Errorf("listing organization policies: %w", err)
	}
	if len(policies) == 0 {
		return EvaluationResult{}, nil
	}
	orgConfig, err := s.orgs.GetOrganization(ctx, org)
	if err != nil {
		return EvaluationResult{}, fmt.Errorf("retrieving organization policy settings: %w", err)
	}
	sets, err := s.db.listPolicySets(ctx, org)
	if err != nil {
		return EvaluationResult{}, fmt.Errorf("listing organization policy sets: %w", err)
	}
	setsByID := make(map[resource.TfeID]*PolicySet, len(sets))
	for _, set := range sets {
		setsByID[set.ID] = set
	}
	modules, err := s.db.listPolicyModulesByOrganization(ctx, org)
	if err != nil {
		return EvaluationResult{}, fmt.Errorf("listing organization policy modules: %w", err)
	}

	policiesBySet := make(map[resource.TfeID][]*Policy)
	for _, policy := range policies {
		policiesBySet[policy.PolicySetID] = append(policiesBySet[policy.PolicySetID], policy)
	}
	modulesBySet := make(map[resource.TfeID][]*PolicyModule)
	for _, module := range modules {
		modulesBySet[module.PolicySetID] = append(modulesBySet[module.PolicySetID], module)
	}

	var (
		bundle *MockBundle
		checks []*PolicyCheck
	)
	for setID, setPolicies := range policiesBySet {
		set, ok := setsByID[setID]
		if !ok {
			return EvaluationResult{}, fmt.Errorf("policy set %s not found", setID)
		}
		evaluator, ok := s.evaluators[set.Kind]
		if !ok {
			return EvaluationResult{}, fmt.Errorf("policy engine %q is not implemented", set.Kind)
		}
		if bundle == nil {
			bundle, err = s.BuildMockBundle(ctx, workspaceID, &runID)
			if err != nil {
				return EvaluationResult{}, fmt.Errorf("building sentinel mock bundle: %w", err)
			}
		}
		evalSet, err := applyOrganizationPolicySettings(set, orgConfig)
		if err != nil {
			return EvaluationResult{}, err
		}
		got, err := evaluator.Evaluate(ctx, evalSet, bundle, setPolicies, modulesBySet[setID])
		if err != nil {
			return EvaluationResult{}, fmt.Errorf("executing %s policies: %w", set.Kind, err)
		}
		checks = append(checks, got...)
	}
	now := internal.CurrentTimestamp(nil)
	result := EvaluationResult{
		Evaluated: true,
		Checks:    checks,
	}
	for _, check := range checks {
		check.RunID = runID
		check.WorkspaceID = workspaceID
		check.Organization = org
		check.CreatedAt = now
		if !check.Passed {
			if check.EnforcementLevel == MandatoryEnforcement {
				result.HasMandatoryFailure = true
			} else {
				result.HasAdvisoryFailure = true
			}
		}
	}
	if err := s.db.replacePolicyChecks(ctx, runID, checks); err != nil {
		return EvaluationResult{}, fmt.Errorf("persisting policy checks: %w", err)
	}
	return result, nil
}

func applyOrganizationPolicySettings(set *PolicySet, org *organization.Organization) (*PolicySet, error) {
	evalSet := *set
	if evalSet.Kind != SentinelKind {
		return &evalSet, nil
	}
	if err := evalSet.EngineVersion.UnmarshalText([]byte(org.SentinelVersion)); err != nil {
		return nil, fmt.Errorf("parsing organization sentinel version: %w", err)
	}
	return &evalSet, nil
}

func (s *Service) handleVCSEvent(event vcs.Event) {
	if event.Type != vcs.EventTypePush {
		return
	}
	switch event.Action {
	case vcs.ActionCreated, vcs.ActionUpdated:
	default:
		return
	}

	ctx := context.Background()
	ctx = authz.AddSkipAuthz(authz.AddSubjectToContext(ctx, &authz.Superuser{Username: "policy-set-sync"}))

	sets, err := s.db.listVCSPolicySetsByRepo(ctx, event.VCSProviderID, event.Repo)
	if err != nil {
		s.Error(err, "listing VCS-backed policy sets for sync", "provider", event.VCSProviderID, "repo", event.Repo)
		return
	}
	for _, set := range sets {
		if !matchesPolicySetRef(set.VCSRef, event) {
			continue
		}
		if _, err := s.SyncPolicySetFromVCS(ctx, set.ID); err != nil {
			s.Error(err, "syncing policy set from VCS event", "policy_set_id", set.ID, "repo", event.Repo, "branch", event.Branch, "sha", event.CommitSHA)
			continue
		}
		s.V(0).Info("synced policy set from VCS event", "policy_set_id", set.ID, "repo", event.Repo, "branch", event.Branch, "sha", event.CommitSHA)
	}
}

func matchesPolicySetRef(ref string, event vcs.Event) bool {
	if ref == "" {
		return event.Branch == event.DefaultBranch
	}
	return ref == event.CommitSHA ||
		ref == event.Branch ||
		ref == event.Tag ||
		ref == "refs/heads/"+event.Branch ||
		ref == "refs/tags/"+event.Tag
}

type api struct {
	*Service
}

func (a *api) addHandlers(r *mux.Router) {
	r = r.PathPrefix("/api").Subrouter()
	r.HandleFunc("/organizations/{organization_name}/policy-sets", a.listPolicySets).Methods(http.MethodGet)
	r.HandleFunc("/organizations/{organization_name}/policy-sets", a.createPolicySet).Methods(http.MethodPost)
	r.HandleFunc("/policy-sets/{policy_set_id}", a.updatePolicySet).Methods(http.MethodPatch)
	r.HandleFunc("/policy-sets/{policy_set_id}", a.deletePolicySet).Methods(http.MethodDelete)
	r.HandleFunc("/policy-sets/{policy_set_id}/policies", a.listPolicies).Methods(http.MethodGet)
	r.HandleFunc("/policy-sets/{policy_set_id}/policies", a.createPolicy).Methods(http.MethodPost)
	r.HandleFunc("/policies/{policy_id}", a.updatePolicy).Methods(http.MethodPatch)
	r.HandleFunc("/policies/{policy_id}", a.deletePolicy).Methods(http.MethodDelete)
	r.HandleFunc("/policy-sets/{policy_set_id}/workspaces", a.setPolicySetWorkspaces).Methods(http.MethodPut)
	r.HandleFunc("/runs/{run_id}/policy-checks", a.listPolicyChecks).Methods(http.MethodGet)
}

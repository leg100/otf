package run

import (
	"context"
	"fmt"
	"regexp"
	"slices"

	"github.com/gobwas/glob"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/vcs"
	"github.com/leg100/otf/internal/workspace"
)

type (
	// Spawner spawns new runs in response to vcs events
	Spawner struct {
		logr.Logger

		configs    spawnerConfigClient
		workspaces spawnerWorkspaceClient
		vcs        spawnerVCSClient
		runs       spawnerRunClient
	}

	spawnerWorkspaceClient interface {
		ListConnectedWorkspaces(ctx context.Context, vcsProviderID resource.TfeID, repoPath vcs.Repo) ([]*workspace.Workspace, error)
	}

	spawnerConfigClient interface {
		Create(ctx context.Context, workspaceID resource.TfeID, opts configversion.CreateOptions) (*configversion.ConfigurationVersion, error)
		Get(ctx context.Context, id resource.TfeID) (*configversion.ConfigurationVersion, error)
		GetLatest(ctx context.Context, workspaceID resource.TfeID) (*configversion.ConfigurationVersion, error)
		UploadConfig(ctx context.Context, id resource.TfeID, config []byte) error
	}

	spawnerVCSClient interface {
		Get(ctx context.Context, providerID resource.TfeID) (*vcs.Provider, error)
	}

	spawnerRunClient interface {
		Create(ctx context.Context, workspaceID resource.TfeID, opts CreateOptions) (*Run, error)
	}
)

func (s *Spawner) handle(event vcs.Event) {
	// TODO: vcs.Event should implement slog.LogValue
	logger := s.Logger.WithValues(
		"sha", event.CommitSHA,
		"type", event.Type,
		"action", event.Action,
		"branch", event.Branch,
		"tag", event.Tag,
	)

	if err := s.handleWithError(logger, event); err != nil {
		s.Error(err, "handling event")
	}
}

func (s *Spawner) handleWithError(logger logr.Logger, event vcs.Event) error {
	// no parent context; handler is called asynchronously
	ctx := context.Background()
	// give spawner unlimited powers
	ctx = authz.AddSubjectToContext(ctx, &authz.Superuser{Username: "run-spawner"})

	// skip events other than those that create or update a ref or pull request
	switch event.Action {
	case vcs.ActionCreated, vcs.ActionUpdated:
	default:
		logger.V(4).Info("ignoring vcs event: non-applicable action")
		return nil
	}

	workspaces, err := s.workspaces.ListConnectedWorkspaces(ctx, event.VCSProviderID, event.Repo)
	if err != nil {
		return err
	}
	if len(workspaces) == 0 {
		logger.V(4).Info("ignoring vcs event: no connected workspaces found")
		return nil
	}

	// filter out workspaces based on info contained in the event
	var n int
	for _, ws := range workspaces {
		switch event.Type {
		case vcs.EventTypeTag:
			// skip workspaces with a non-nil tag regex that doesn't match the
			// tag event
			if ws.Connection.TagsRegex != "" {
				re := regexp.MustCompile(ws.Connection.TagsRegex)
				if !re.MatchString(event.Tag) {
					continue
				}
			}
		case vcs.EventTypePush:
			if ws.Connection.Branch != "" {
				// skip workspaces with a user-specified branch that doesn't match the
				// event branch
				if ws.Connection.Branch != event.Branch {
					continue
				}
			} else {
				// skip workspaces with default branch for events on non-default
				// branches
				if event.Branch != event.DefaultBranch {
					continue
				}
			}
			if ws.Connection.TagsRegex != "" {
				// skip workspaces which specify a tags regex
				continue
			}
		case vcs.EventTypePull:
			if !ws.SpeculativeEnabled {
				// skip workspaces configured to not trigger speculative plans
				// (plan-only runs) in response to pull request events.
				continue
			}
		}

		// only tag and push events contain a list of changed files
		switch event.Type {
		case vcs.EventTypeTag, vcs.EventTypePush:
			// filter workspaces with trigger pattern that doesn't match any of the
			// files in the event
			if ws.TriggerPatterns != nil {
				if !globMatch(event.Paths, ws.TriggerPatterns) {
					continue
				}
			}
		}
		workspaces[n] = ws
		n++
	}
	if n == 0 {
		// no workspaces survived the filter
		logger.V(4).Info("ignoring vcs event: no matching triggers found")
		return nil
	}
	workspaces = workspaces[:n]

	// fetch tarball
	client, err := s.vcs.Get(ctx, event.VCSProviderID)
	if err != nil {
		return err
	}
	tarball, _, err := client.GetRepoTarball(ctx, vcs.GetRepoTarballOptions{
		Repo: event.Repo,
		Ref:  &event.CommitSHA,
	})
	if err != nil {
		return fmt.Errorf("retrieving repo tarball: %w", err)
	}

	// pull request events don't contain a list of changed files; instead an API
	// call is necessary to retrieve the list of changed files
	if event.Type == vcs.EventTypePull {
		// only perform API call if at least one workspace has file triggers
		// enabled.
		var listFiles bool
		for _, ws := range workspaces {
			if ws.TriggerPatterns != nil {
				listFiles = true
				break
			}
		}
		if listFiles {
			paths, err := client.ListPullRequestFiles(ctx, event.Repo, event.PullRequestNumber)
			if err != nil {
				return fmt.Errorf("retrieving list of files in pull request from cloud provider: %w", err)
			}
			n := 0
			for _, ws := range workspaces {
				if ws.TriggerPatterns != nil && !globMatch(paths, ws.TriggerPatterns) {
					// skip workspace
					continue
				}
				workspaces[n] = ws
				n++
			}
			workspaces = workspaces[:n]
		}
	}

	// create a config version for each workspace and spawn run.
	for _, ws := range workspaces {
		cvOpts := configversion.CreateOptions{
			// pull request events trigger speculative runs
			Speculative: new(event.Type == vcs.EventTypePull),
			IngressAttributes: &configversion.IngressAttributes{
				// ID     resource.TfeID
				Branch: event.Branch,
				// CloneURL          string
				// CommitMessage     string
				CommitSHA: event.CommitSHA,
				CommitURL: event.CommitURL,
				// CompareURL        string
				Repo:              ws.Connection.Repo,
				IsPullRequest:     event.Type == vcs.EventTypePull,
				OnDefaultBranch:   event.Branch == event.DefaultBranch,
				PullRequestNumber: event.PullRequestNumber,
				PullRequestTitle:  event.PullRequestTitle,
				PullRequestURL:    event.PullRequestURL,
				SenderUsername:    event.SenderUsername,
				SenderAvatarURL:   event.SenderAvatarURL,
				SenderHTMLURL:     event.SenderHTMLURL,
				Tag:               event.Tag,
			},
			Source: event.EventHeader.Source,
		}
		runOpts := CreateOptions{
			Source: event.EventHeader.Source,
		}
		cv, err := s.configs.Create(ctx, ws.ID, cvOpts)
		if err != nil {
			return err
		}
		if err := s.configs.UploadConfig(ctx, cv.ID, tarball); err != nil {
			return err
		}
		runOpts.ConfigurationVersionID = &cv.ID
		_, err = s.runs.Create(ctx, ws.ID, runOpts)
		if err != nil {
			return err
		}
	}
	return nil
}

// globMatch returns true if any of the paths match any of the glob patterns.
func globMatch(paths []string, patterns []string) bool {
	if len(paths) == 0 || len(patterns) == 0 {
		return false
	}
	for _, pattern := range patterns {
		g := glob.MustCompile(pattern)
		if slices.ContainsFunc(paths, g.Match) {
			// only one match is necessary
			return true
		}
	}
	return false
}

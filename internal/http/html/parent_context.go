package html

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/leg100/otf/internal/resource"
)

var (
	workspaceNamePath  = regexp.MustCompile(`/organizations/([^/]+)/workspaces/([^/]+)`)
	organizationPathRe = regexp.MustCompile(`/organizations/([^/]+)`)
)

type parentContextResolver interface {
	GetParentOrganizationID(ctx context.Context, id resource.ID) (resource.ID, error)
	GetParentWorkspaceID(ctx context.Context, id resource.ID) (resource.ID, error)
}

type parentContextWorkspaceClient interface {
	GetName(workspaceID resource.TfeID) (string, error)
}

type ParentContext struct {
	workspaces parentContextWorkspaceClient
	resolver   parentContextResolver
}

// addParents is middleware that adds a resource's parent resources to the
// request context, where parents are its workspace and/or its organization.
func (p *ParentContext) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Ignore requests other than for html pages
		if r.Header.Get("Content-type") != "text/html" {
			next.ServeHTTP(w, r)
		}
		ctx := r.Context()
		if matches := workspaceNamePath.FindStringSubmatch(r.URL.Path); matches != nil {
			// Both organization and workspace name are already in the path.
			ctx = addOrganizationNameToContext(ctx, matches[1])
			ctx = addWorkspaceNameToContext(ctx, matches[2])
		} else if matches := organizationPathRe.FindStringSubmatch(r.URL.Path); matches != nil {
			// Only organization name is in path; paths that specify an
			// organization name are only ever for a resource that does not
			// belong to a workspace, e.g. the organization page, or a list of
			// vcs providers for an org.
			ctx = addOrganizationNameToContext(ctx, matches[1])
		} else {
			// Organization name not in path. This means there is no workspace
			// name either (it only appears in tandem with organization name).
			// Now we check for any IDs that appear in the path, which could be
			// a workspace ID or the ID of any other resource (apart from
			// organizations: only the organization *name* is used in paths).
			//
			// TODO: fix this comment above
			for comp := range strings.SplitSeq(r.URL.Path, "/") {
				id, err := resource.ParseTfeID(comp)
				if err != nil {
					// Not an ID
					continue
				}
				wsID, err := p.resolver.GetParentWorkspaceID(ctx, id)
				if err == nil {
					// workspace ID must be cast to a TFE ID before its name can
					// be looked up.
					wsTfeID, ok := wsID.(resource.TfeID)
					if !ok {
						continue
					}
					name, err := p.workspaces.GetName(wsTfeID)
					if err == nil {
						ctx = addWorkspaceNameToContext(ctx, name)
					}
				}
				orgID, err := p.resolver.GetParentOrganizationID(ctx, id)
				if err == nil {
					ctx = addOrganizationNameToContext(ctx, orgID.String())
				}
			}
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// unexported key types prevents collisions
type (
	orgCtxKey string
	wsCtxKey  string
)

const (
	orgCtx orgCtxKey = "organization"
	wsCtx  wsCtxKey  = "workspace"
)

// addOrganizationNameToContext adds a subject to a context
func addOrganizationNameToContext(ctx context.Context, organization string) context.Context {
	return context.WithValue(ctx, orgCtx, organization)
}

// OrganizationFromContext retrieves a subject from a context
func OrganizationFromContext(ctx context.Context) (string, error) {
	name, ok := ctx.Value(orgCtx).(string)
	if !ok {
		return "", fmt.Errorf("no organization in context")
	}
	return name, nil
}

// addWorkspaceNameToContext adds the workspace name to the context
func addWorkspaceNameToContext(ctx context.Context, workspace string) context.Context {
	return context.WithValue(ctx, wsCtx, workspace)
}

// WorkspaceFromContext retrieves a workspace name from the context
func WorkspaceFromContext(ctx context.Context) (string, error) {
	name, ok := ctx.Value(wsCtx).(string)
	if !ok {
		return "", fmt.Errorf("no workspace in context")
	}
	return name, nil
}

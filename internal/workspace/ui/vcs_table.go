package ui

import (
	"context"
	"io"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/ui/helpers"
	"github.com/leg100/otf/internal/ui/paths"
	"github.com/leg100/otf/internal/vcs"
	"github.com/templ-go/x/urlbuilder"
)

// latestRunUpdate is the SSE event name used by workspace_view.templ to receive
// updates for the latest run table. It must match the value used in run/ui.
const latestRunUpdate = "LatestRunUpdate"

// vcsTable implements helpers.TablePopulator[*vcs.Provider] for displaying a
// list of VCS providers. It is defined here (in workspace/ui) because
// workspace_view.templ references it and that templ file lives in this package.
type vcsTable struct {
	Actions func(vcsProviderID resource.TfeID) templ.Component
}

func (t vcsTable) Header() templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := io.WriteString(w, "<th>Name</th><th>ID</th><th>Created</th><th>Action</th>")
		return err
	})
}

func (t vcsTable) Row(provider *vcs.Provider) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := io.WriteString(w, `<tr id="item-vcsprovider-`+templ.EscapeString(provider.String())+`"><td><div class="flex gap-2">`)
		if err != nil {
			return err
		}
		if err := provider.Icon.Render(ctx, w); err != nil {
			return err
		}
		_, err = io.WriteString(w, `<a class="link" href="`+templ.EscapeString(string(paths.EditVCSProvider(provider.ID)))+`">`+templ.EscapeString(provider.String())+`</a></div></td><td>`)
		if err != nil {
			return err
		}
		if err := helpers.Identifier(provider.ID).Render(ctx, w); err != nil {
			return err
		}
		_, err = io.WriteString(w, "</td><td>"+templ.EscapeString(internal.Ago(time.Now(), provider.CreatedAt))+"</td><td>")
		if err != nil {
			return err
		}
		if err := t.Actions(provider.ID).Render(ctx, w); err != nil {
			return err
		}
		_, err = io.WriteString(w, "</td></tr>")
		return err
	})
}

// RepoURL builds a URL to the repository homepage on the VCS provider.
func RepoURL(provider *vcs.Provider, repo vcs.Repo) templ.SafeURL {
	b := urlbuilder.New(provider.BaseURL.Scheme, provider.BaseURL.Host)
	for segment := range strings.SplitSeq(repo.Owner(), "/") {
		b.Path(segment)
	}
	b.Path(repo.Name())
	return b.Build()
}

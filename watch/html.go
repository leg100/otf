package watch

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/gin-contrib/sse"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/html"
)

type htmlApp struct {
	otf.Renderer
}

func (app *htmlApp) watchWorkspace(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		WorkspaceID string `schema:"workspace_id,required"`
		StreamID    string `schema:"stream,required"`
	}
	var params parameters
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := app.GetWorkspace(r.Context(), params.WorkspaceID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	events, err := app.Watch(r.Context(), otf.WatchOptions{
		WorkspaceID: otf.String(ws.ID()),
	})
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	go func() {
		for {
			select {
			case <-r.Context().Done():
				return
			case event, ok := <-events:
				if !ok {
					return
				}
				run, ok := event.Payload.(*otf.Run)
				if !ok {
					// skip non-run events
					continue
				}

				// Handle query parameters which filter run events:
				// - 'latest' specifies that the client is only interest in events
				// relating to the latest run for the workspace
				// - 'run-id' (mutually exclusive with 'latest') - specifies
				// that the client is only interested in events relating to that
				// run.
				// - otherwise, if neither of those parameters are specified
				// then events for all runs are relayed.
				if r.URL.Query().Get("latest") != "" {
					if !run.Latest() {
						// skip: run is not the latest run for a workspace
						continue
					}
				} else if runID := r.URL.Query().Get("run-id"); runID != "" {
					if runID != run.ID() {
						// skip: event is for a run which does not match the
						// filter
						continue
					}
				}

				// render HTML snippets and send as payload in SSE event
				itemHTML := new(bytes.Buffer)
				if err := app.RenderTemplate("run_item.tmpl", itemHTML, run); err != nil {
					app.Error(err, "rendering template for run item")
					continue
				}
				runStatusHTML := new(bytes.Buffer)
				if err := app.RenderTemplate("run_status.tmpl", runStatusHTML, run); err != nil {
					app.Error(err, "rendering run status template")
					continue
				}
				planStatusHTML := new(bytes.Buffer)
				if err := app.RenderTemplate("phase_status.tmpl", planStatusHTML, run.Plan()); err != nil {
					app.Error(err, "rendering plan status template")
					continue
				}
				applyStatusHTML := new(bytes.Buffer)
				if err := app.RenderTemplate("phase_status.tmpl", applyStatusHTML, run.Apply()); err != nil {
					app.Error(err, "rendering apply status template")
					continue
				}
				js, err := json.Marshal(struct {
					ID              string        `json:"id"`
					RunStatus       otf.RunStatus `json:"run-status"`
					RunItemHTML     string        `json:"run-item-html"`
					RunStatusHTML   string        `json:"run-status-html"`
					PlanStatusHTML  string        `json:"plan-status-html"`
					ApplyStatusHTML string        `json:"apply-status-html"`
				}{
					ID:              run.ID(),
					RunStatus:       run.Status(),
					RunItemHTML:     itemHTML.String(),
					RunStatusHTML:   runStatusHTML.String(),
					PlanStatusHTML:  planStatusHTML.String(),
					ApplyStatusHTML: applyStatusHTML.String(),
				})
				if err != nil {
					app.Error(err, "marshalling watched run", "run", run.ID())
					continue
				}
				app.Server.Publish(params.StreamID, &sse.Event{
					Data:  js,
					Event: []byte(event.Type),
				})
			}
		}
	}()
	app.Server.ServeHTTP(w, r)
}

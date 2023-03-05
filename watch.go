package otf

import "context"

// DefaultWatchPath is the default relative path to the watch endpoint
const DefaultWatchPath = "/watch"

type (
	WatchService interface {
		// Watch provides access to a stream of events. The WatchOptions filters
		// events. Context must be cancelled to close stream.
		//
		// TODO(@leg100): it would be clearer to the caller if the stream is closed by
		// returning a stream object with a Close() method. The calling code would
		// call Watch(), and then defer a Close(), which is more readable IMO.
		Watch(ctx context.Context, opts WatchOptions) (<-chan Event, error)
	}

	// WatchOptions filters events returned by the Watch endpoint.
	WatchOptions struct {
		// Name to uniquely describe the watcher. If not provided then a
		// name will be auto generated.
		Name         *string
		Organization *string `schema:"organization_name"` // filter by organization name
		WorkspaceID  *string `schema:"workspace_id"`      // filter by workspace ID; mutually exclusive with organization filter
	}
)

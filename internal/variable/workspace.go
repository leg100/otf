package variable

import "log/slog"

type (
	// WorkspaceVariable is a workspace-scoped variable.
	WorkspaceVariable struct {
		*Variable

		WorkspaceID string
	}
)

func (v *WorkspaceVariable) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("workspace_id", v.WorkspaceID),
		slog.Any("variable", v.Variable.LogValue().Any()),
	}
	return slog.GroupValue(attrs...)
}

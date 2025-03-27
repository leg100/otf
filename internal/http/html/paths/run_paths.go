// Code generated by "go generate"; DO NOT EDIT.

package paths

import (
	"fmt"

	"github.com/leg100/otf/internal/resource"
)

func Runs(workspace resource.ID) string {
	return fmt.Sprintf("/app/workspaces/%s/runs", workspace)
}

func CreateRun(workspace resource.ID) string {
	return fmt.Sprintf("/app/workspaces/%s/runs/create", workspace)
}

func NewRun(workspace resource.ID) string {
	return fmt.Sprintf("/app/workspaces/%s/runs/new", workspace)
}

func Run(run resource.ID) string {
	return fmt.Sprintf("/app/runs/%s", run)
}

func EditRun(run resource.ID) string {
	return fmt.Sprintf("/app/runs/%s/edit", run)
}

func UpdateRun(run resource.ID) string {
	return fmt.Sprintf("/app/runs/%s/update", run)
}

func DeleteRun(run resource.ID) string {
	return fmt.Sprintf("/app/runs/%s/delete", run)
}

func ApplyRun(run resource.ID) string {
	return fmt.Sprintf("/app/runs/%s/apply", run)
}

func DiscardRun(run resource.ID) string {
	return fmt.Sprintf("/app/runs/%s/discard", run)
}

func CancelRun(run resource.ID) string {
	return fmt.Sprintf("/app/runs/%s/cancel", run)
}

func ForceCancelRun(run resource.ID) string {
	return fmt.Sprintf("/app/runs/%s/force-cancel", run)
}

func RetryRun(run resource.ID) string {
	return fmt.Sprintf("/app/runs/%s/retry", run)
}

func TailRun(run resource.ID) string {
	return fmt.Sprintf("/app/runs/%s/tail", run)
}

func WidgetRun(run resource.ID) string {
	return fmt.Sprintf("/app/runs/%s/widget", run)
}

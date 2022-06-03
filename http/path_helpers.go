package http

import (
	"fmt"

	"github.com/leg100/otf"
)

func uploadConfigurationVersionPath(cv *otf.ConfigurationVersion) string {
	return fmt.Sprintf("/configuration-versions/%s/upload", cv.ID())
}

func getPlanLogsPath(plan *otf.Plan) string {
	return fmt.Sprintf("plans/%s/logs", plan.ID())
}

func getApplyLogsPath(apply *otf.Apply) string {
	return fmt.Sprintf("applies/%s/logs", apply.ID())
}

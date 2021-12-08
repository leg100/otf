package otf

import "strconv"

var (
	// Build-time parameters set -ldflags
	Version = "unknown"
	Commit  = "unknown"
	Built   = "unknown"

	// BuildInt is an integer representation of Built
	BuiltInt int
)

func init() {
	// Convert Built into BuiltTime
	var err error
	BuiltInt, err = strconv.Atoi(Built)
	if err != nil {
		// On error just set to 0 (so we can run can continue to run go build
		// without -ldflags)
		BuiltInt = 0
	}
}

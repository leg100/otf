//go:build ignore

package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/leg100/otf/internal/runstatus"
)

//go:generate go run gen.go

var themeMappings = map[runstatus.Status]string{
	runstatus.Applied:            "success",
	runstatus.ApplyQueued:        "secondary",
	runstatus.Applying:           "accent",
	runstatus.Canceled:           "warning",
	runstatus.Confirmed:          "info",
	runstatus.Discarded:          "warning",
	runstatus.Errored:            "error",
	runstatus.ForceCanceled:      "warning",
	runstatus.Pending:            "primary",
	runstatus.PlanQueued:         "secondary",
	runstatus.Planned:            "info",
	runstatus.PlannedAndFinished: "success",
	runstatus.Planning:           "primary",
}

var components = []string{
	"badge",
	"bg",
	"checkbox",
	"text",
}

func main() {
	var b strings.Builder

	for status, semantic := range themeMappings {
		for _, component := range components {
			b.WriteString(fmt.Sprintf(".%s-%s {\n", component, status.String()))
			b.WriteString(fmt.Sprintf("\t@apply %s-%s;\n", component, semantic))
			b.WriteString("}\n")
		}
	}

	if err := os.WriteFile("../http/html/static/css/runstatus.css", []byte(b.String()), 0o755); err != nil {
		log.Fatal("Error: ", err.Error())
	}
}

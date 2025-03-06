package runstatus

import (
	"fmt"
	"log"
	"os"
)

var components = []string{
	"badge",
	"bg",
	"checkbox",
	"text",
}

func main() {
	f, err := os.Create("class_names.txt")
	if err != nil {
		log.Fatal("Error: ", err.Error())
	}
	defer f.Close()

	for status, semantic := range ThemeMappings {
		for _, component := range components {
			f.WriteString(fmt.Sprintf("%s-%s", status.String(), semantic))
		}
	}
}

module github.com/leg100/ots

go 1.16

require (
	github.com/google/jsonapi v1.0.0
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/schema v1.2.0
	github.com/hashicorp/jsonapi v0.0.0-20210518035559-1e50d74c8db3
	github.com/leg100/go-tfe v0.17.1-0.20210702211340-dae2585de0d6
	github.com/mattn/go-sqlite3 v1.14.7 // indirect
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.3.0
	github.com/urfave/negroni v1.0.0
	gorm.io/driver/sqlite v1.1.4
	gorm.io/gorm v1.21.10
)

//replace github.com/leg100/go-tfe => ../go-tfe

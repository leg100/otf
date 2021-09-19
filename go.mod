module github.com/leg100/otf

go 1.16

require (
	github.com/go-logr/logr v1.0.0
	github.com/google/uuid v1.3.0
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/schema v1.2.0
	github.com/hashicorp/hcl/v2 v2.10.0
	github.com/leg100/go-tfe v0.17.1-0.20210804135856-d05f7109100e
	github.com/leg100/gorm-zerolog v0.1.1-0.20210718123649-2348997004e6
	github.com/leg100/jsonapi v1.0.1-0.20210703183827-d0513d61dc3f
	github.com/leg100/zerologr v0.0.0-20210805173127-2e0b118333c5
	github.com/mattn/go-sqlite3 v1.14.7 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/rs/zerolog v1.23.0
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.3.0
	github.com/urfave/negroni v1.0.0
	gorm.io/driver/sqlite v1.1.4
	gorm.io/gorm v1.21.10
)

//replace github.com/leg100/go-tfe => ../go-tfe

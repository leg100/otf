module github.com/leg100/otf

go 1.16

require (
	github.com/Masterminds/squirrel v1.5.0
	github.com/go-logr/logr v1.0.0
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/google/uuid v1.3.0
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/schema v1.2.0
	github.com/gorilla/websocket v1.4.2
	github.com/hashicorp/go-cleanhttp v0.5.1
	github.com/hashicorp/go-multierror v1.0.0
	github.com/hashicorp/go-retryablehttp v0.5.2
	github.com/hashicorp/hcl/v2 v2.10.0
	github.com/iancoleman/strcase v0.2.0
	github.com/jmoiron/sqlx v1.3.1
	github.com/leg100/jsonapi v1.0.1-0.20210703183827-d0513d61dc3f
	github.com/leg100/zerologr v0.0.0-20210805173127-2e0b118333c5
	github.com/lib/pq v1.10.3
	github.com/mattn/go-isatty v0.0.3
	github.com/mitchellh/copystructure v1.2.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/pressly/goose/v3 v3.1.0
	github.com/rs/zerolog v1.23.0
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.6.1
	github.com/urfave/negroni v1.0.0
	golang.org/x/text v0.3.6 // indirect
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
)

//replace github.com/leg100/go-tfe => ../go-tfe

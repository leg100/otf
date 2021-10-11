module github.com/leg100/otf

go 1.16

require (
	github.com/ClickHouse/clickhouse-go v1.5.0 // indirect
	github.com/Masterminds/squirrel v1.5.0
	github.com/denisenkom/go-mssqldb v0.11.0 // indirect
	github.com/go-logr/logr v1.0.0
	github.com/golang-migrate/migrate/v4 v4.15.0
	github.com/google/uuid v1.3.0
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/schema v1.2.0
	github.com/gorilla/websocket v1.4.2
	github.com/hashicorp/go-cleanhttp v0.5.1
	github.com/hashicorp/go-retryablehttp v0.5.2
	github.com/hashicorp/hcl/v2 v2.10.0
	github.com/iancoleman/strcase v0.2.0
	github.com/jinzhu/copier v0.3.2
	github.com/jmoiron/sqlx v1.3.1
	github.com/leg100/gorm-zerolog v0.1.1-0.20210718123649-2348997004e6
	github.com/leg100/jsonapi v1.0.1-0.20210703183827-d0513d61dc3f
	github.com/leg100/zerologr v0.0.0-20210805173127-2e0b118333c5
	github.com/lib/pq v1.10.3 // indirect
	github.com/mattn/go-sqlite3 v1.14.8
	github.com/mitchellh/copystructure v1.2.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/pressly/goose/v3 v3.1.0
	github.com/rs/zerolog v1.23.0
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.6.1
	github.com/urfave/negroni v1.0.0
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519 // indirect
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e
	gorm.io/driver/sqlite v1.1.4
	gorm.io/gorm v1.21.10
)

//replace github.com/leg100/go-tfe => ../go-tfe

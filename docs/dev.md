# Development

Code contributions are welcome.

## Setup your machine

* Clone the repo:

```
git clone git@github.com:leg100/otf.git
```

* Install [Go](https://go.dev/doc/install).
* Install [PostgreSQL](https://www.postgresql.org/download/) (optional).

## SQL migrations

The database schema is migrated using [goose](https://github.com/pressly/goose). The SQL migration files are kept in the repo in `./sql/migrations`. Upon startup `otfd` automatically migrates the DB to the latest version.

If you're developing a SQL migration you may want to migrate the database manually. Use the `make` tasks to assist you:

* `make migrate`
* `make migrate-redo`
* `make migrate-rollback`
* `make migrate-status`

## SQL queries

SQL queries are handwritten in `./sql/queries` and turned into Go using [pggen](https://github.com/jschaf/pggen).

After you make changes to the queries run the following make task to invoke `pggen`:

* `make sql`

## HTML path helpers

Rails-style path helpers are generated using `go generate`. The path specifications are maintaining in `./http/html/paths/gen.go`. After making changes to the specs run the following make task to generate the helpers:

* `make paths`

## Web development

If you're making changes to web templates then you may want to enable [developer mode](../config/flags/#-dev-mode). Once enabled you will be able to see the changes without restarting `otfd`: while `otfd` is running, you can make a change to a template and then reload the page in your browser and you should see the change.

To auto-reload the browser, check out the recommended [developer tooling](#developer-tooling).

## Developer tooling

Both [modd](https://github.com/cortesi/modd) and [devd](https://github.com/cortesi/devd) are recommended to automate development tasks:

* Automatically restart `otfd` whenever changes are made to code.
* Automatically reload the browser whenever changes are made to templates, CSS, etc.
* Automatically generate Go code whenever SQL queries are updated.
* Automatically generate path helpers whenever path specifications are updated.

A `modd.conf` is included in the OTF repository. Once you've installed `devd` and `modd`, run `modd` in the root of the repository and it'll perform the above tasks.

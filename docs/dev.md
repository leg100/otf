# Development

Code contributions are welcome.

## Setup your machine

* Clone the repo:

```
git clone git@github.com:leg100/otf.git
```

* Install [Go](https://go.dev/doc/install).
* Install [PostgreSQL](https://www.postgresql.org/download/) (optional).

## Documentation

The documentation uses [Material for MkDocs](https://squidfunk.github.io/mkdocs-material/). A [fork](https://github.com/leg100/mkdocs-material) is maintained which makes a couple of minor aesthetic changes.

The documentation pages are maintained in the `./docs` directory of the repository. To make small edits it is recommended you click on the `Edit this page` icon (see top right of this page), which'll take you to Github and prompt you to make a pull request.

For larger changes, you can use the following make task:

```
make serve-docs
```

That builds and runs the documentation site on your workstation at `http://localhost:9999`. Any changes you make to the documentation are reflected in real-time in the browser.

Screenshots in the documentation are largely automated. The browser-based integration tests produce screenshots at various steps. If the environment variable `OTF_DOC_SCREENSHOTS=true` is present then such a test also writes the screenshot into the documentation directory. The following make task runs the tests along with the aforementioned environment variable:

```
make doc-screenshots
```

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

Rails-style path helpers are generated using `go generate`. The path specifications are maintained in `./http/html/paths/gen.go`. After making changes to the specs run the following make task to generate the helpers:

* `make paths`

## Web development

If you're making changes to web templates then you may want to enable [developer mode](../config/flags/#-dev-mode). Once enabled you will be able to see the changes without restarting `otfd`: while `otfd` is running, you can make a change to a template and then reload the page in your browser and you should see the change.

To auto-reload the browser, check out the recommended [developer tooling](#developer-tooling).

## Developer tooling

[modd](https://github.com/cortesi/modd) is recommended to automate development tasks:

* Automatically restart `otfd` whenever changes are made to code.
* Automatically generate Go code whenever SQL queries are updated.
* Automatically generate path helpers whenever path specifications are updated.

A `modd.conf` is included in the OTF repository. Once you've installed `modd`, run it from the root of the repository and it'll perform the above tasks.

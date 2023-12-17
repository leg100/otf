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

Before running the make task you'll need to install some python packages:

```
pip install -r docs/requirements.txt
```

Screenshots in the documentation are largely automated. The browser-based integration tests produce screenshots at various steps. If the environment variable `OTF_DOC_SCREENSHOTS=true` is present then such a test also writes the screenshot into the documentation directory. The following make task runs the tests along with the aforementioned environment variable:

```
make doc-screenshots
```

## SQL migrations

The database schema is migrated using [tern](https://github.com/jackc/tern). The SQL migration files are kept in the repo in `./sql/migrations`. Upon startup `otfd` automatically migrates the DB to the latest version.

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

If you're making changes to web templates then you may want to enable [developer mode](config/flags.md/#-dev-mode). Once enabled you will be able to see changes without restarting `otfd`.

OTF uses [Tailwind CSS](https://tailwindcss.com/) to generate CSS classes. Run the following make task to generate the CSS:

* `make tailwind`

!!! note
    To install tailwind first ensure you've installed `npm` and then run `npm install -D tailwindcss`

## Developer tooling

[modd](https://github.com/cortesi/modd) is recommended to automate development tasks:

* Automatically restart `otfd` whenever changes are made to code.
* Automatically generate Go code whenever SQL queries are updated.
* Automatically generate path helpers whenever path specifications are updated.

A `modd.conf` is included in the OTF repository. Once you've installed `modd`, run it from the root of the repository and it'll perform the above tasks.

The following make task runs not only `modd` but watches for changes to Tailwind CSS classes (see above) and generates CSS:

* `make watch -j`

The `-j` flag permits both `modd` and the tailwind watcher to run in parallel.

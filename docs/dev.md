# Development

Code contributions are welcome.

## Setup your machine

* Clone the repo:

```
git clone git@github.com:leg100/otf.git
```

* Install [Go](https://go.dev/doc/install).
* Install [PostgreSQL](https://www.postgresql.org/download/) (optional).

## Tests

### Unit tests

Change into the repo directory and run unit tests:

```
go test ./...
```

### Integration tests

Integration tests require:

* PostgreSQL
* Terraform >= 1.2.0
* Chrome

Set the environment variable `OTF_TEST_DATABASE_URL` to a valid connection string. For example, if you have installed postgres on your local machine with the default database `postgres`:

```
export OTF_TEST_DATABASE_URL=postgres:///postgres
```

Then run the both unit and integration tests:

```
go test ./...
```

!!! note
	Tests check for the presence of `OTF_TEST_DATABASE_URL`. If it absent then only unit tests are run; otherwise both unit and integration tests are run.

#### Postgres too many connections error

When running integration tests you may run into this error:

```
FATAL: sorry, too many clients already
```

The tests make potentially hundreds of connections to postgres at a time and you'll need to increase the maximum number of connections in postgres. Increase the `max_connections` parameter on your installation:

```
psql postgres:///otf
alter system set max_connections = 999;
```

Then reload/restart postgres. On Ubuntu, for example:

```
sudo systemctl reload postgres
```

!!! note
    This is only recommended for a postgres installation dedicated to development and testing purposes. Too many connections in production most likely indicates a problem with the client application.


#### Postgres performance optimisation

To further improve the speed of your integration tests turn off `fsync`:

```
psql postgres:///otf
alter system set fsync = off;
```

Then reload/restart postgres. On Ubuntu, for example:

```
sudo systemctl reload postgres
```

!!! note
    This is only recommended for a postgres installation dedicated to development and testing purposes. Disabling `fsync` can lead to data loss.

#### Database cleanup

A dedicated logical database is created in postgres for each individual
integration test; as a result the above command creates 100s of databases. Upon
test completion the databases are removed. In certain situtations upon a test
failure the test may fail to remove a database, in which case you will have to
manually remove the database (although this is often not a concern other than
consuming a small amount of disk space).

#### Disable headless mode

Browser-based tests spawn a headless Chrome process. In certain situations it
can be useful to disable headless mode, e.g. if a test is stuck on a certain
page and you want to know which page. To disable headless mode:

```
export OTF_E2E_HEADLESS=false
```

#### Cache provider requests

Terraform-based tests spawn `terraform`. These tests retrieve providers from
the internet which can consume quite a lot of bandwidth and slow down the tests
significantly. To cache these providers it is recommended to use a caching
proxy. The following make task runs [squid](http://www.squid-cache.org/) in a
docker container:

```
make squid
```

It is configured to use
[SSL-bumping](https://wiki.squid-cache.org/Features/SslBump), which permits
caching content transported via SSL (`terraform` retrieves providers only via
SSL).

You then need to instruct the tests to use the proxy:

```
export HTTPS_PROXY=localhost:3128
```

You should now find the tests consume a lot less bandwidth and run several times
faster.

### go-tfe tests

Tests from the [go-tfe](https://github.com/hashicorp/go-tfe) project are invoked to test adherence to the documented Terraform Cloud API. A subset of the tests are invoked from a [fork](https://github.com/leg100/go-tfe) using the following make task:

* `make go-tfe-tests`

It requires postgres and otfd to be running, both of which can be started using docker compose:

* `docker compose up -d`

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

If you're making changes to web templates then you may want to enable [developer mode](/config/flags#-dev-mode). Once enabled you will be able to see the changes without restarting `otfd`: while `otfd` is running, you can make a change to a template and then reload the page in your browser and you should see the change.

To auto-reload the browser, check out the recommended [developer tooling](#developer-tooling).

## Developer tooling

Both [modd](https://github.com/cortesi/modd) and [devd](https://github.com/cortesi/devd) are recommended to automate development tasks:

* Automatically restart `otfd` whenever changes are made to code.
* Automatically reload the browser whenever changes are made to templates, CSS, etc.
* Automatically generate Go code whenever SQL queries are updated.
* Automatically generate path helpers whenever path specifications are updated.

A `modd.conf` is included in the OTF repository. Once you've installed `devd` and `modd`, run `modd` in the root of the repository and it'll perform the above tasks.

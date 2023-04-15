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

# Tests

### Unit tests

Change into the repo directory and run unit tests:

```
go test -short ./...
```

### Integration tests

Integration tests require:

* docker compose
* terraform >= 1.2.0
* Chrome

The following runs integration tests:

```
go test ./internal/integration...
```

The tests first stands up external services using docker compose:

* postgres (integration tests require a real database)
* squid cache (speeds up tests by caching terraform providers)
* GCP pub/sub emulator (necessary for the GCP pub/sub integration test)

#### Disable headless mode

Browser-based tests spawn a headless Chrome process. In certain situations it
can be useful to disable headless mode, e.g. if a test is stuck on a certain
page and you want to know which page. To disable headless mode:

```
export OTF_E2E_HEADLESS=false
```

<figure markdown>
![headless mode disabled](images/integration_tests_headless_mode_disabled.png){.screenshot}
<figcaption>Integration tests with headless mode disabled</figcaption>
</figure>

### API tests

Tests from the [go-tfe](https://github.com/hashicorp/go-tfe) project are routinely run to ensure OTF correctly implements the documented Terraform Cloud API. However, OTF only implements a subset of the API endpoints, and there are some differences (e.g. an OTF organization has no email address where as a TFC organization does). Therefore a [fork](https://github.com/leg100/go-tfe) of the go-tfe repo is maintained.

The make task:

```
make go-tfe-tests
```

performs the following steps:

* Starts a docker compose stack of `otfd`, postgres, and squid
* Runs a subset of `go-tfe` tests using the **forked** repo
* Runs a subset of `go-tfe` tests using the **upstream** repo

!!! note
    You can instead manually invoke API tests using the scripts in `./hack`. The tests first require `otfd` to be running at `https://localhost:8080`, with a [site token](../config/flags/#-site-token) set to `site-token`. These settings can be overridden with the environment variables `TFE_ADDRESS` and `TFE_TOKEN`.

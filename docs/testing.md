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

#### More verbose logging

By default, the integration tests don't print the logs from the OTF daemons they spawn. To enable logging with a verbosity of 1, set the following environment variable:

```
export OTF_INTEGRATION_TEST_ENABLE_LOGGER=yes go test -v ./internal/integration
```

Because the tests run in parallel and each test runs its own daemon, you'll see the logs from multiple daemons intermingled. You'll instead probably want to run one test at a time, and to stop at the first failing test:

```
export OTF_INTEGRATION_TEST_ENABLE_LOGGER=yes go test -v ./internal/integration -parallel 1 -failfast
```

This can be helpful for diagnosing the cause of a failing test.

### API tests

Tests from the [go-tfe](https://github.com/hashicorp/go-tfe) project are routinely run to ensure OTF correctly implements the documented Terraform Cloud API.

The make task:

```
make go-tfe-tests
```

performs the following steps:

* Starts a docker compose stack of `otfd`, postgres, and squid
* Runs a subset of `go-tfe` tests against that stack

The tests require the following environment variables:

* `GITHUB_POLICY_SET_IDENTIFIER`: set to a github repo on which the tests can create webhooks.
* `OAUTH_CLIENT_GITHUB_TOKEN`: a personal access token with permissions to create webhooks on the above repo.

!!! note
    You can instead manually invoke API tests using the scripts in `./hack`. The tests first require `otfd` to be running at `https://localhost:8080`, with a [site token](../config/flags/#-site-token) set to `site-token`. These settings can be overridden with the environment variables `TFE_ADDRESS` and `TFE_TOKEN`.

!!! note
    The tests create webhooks on the github repository specified in `GITHUB_POLICY_SET_IDENTIFIER`. The tests should delete the webhooks once they're finished. However, should the tests fail and/or panic, then the webhooks won't be deleted and you'll quickly run into the maximum limit of 20 webhooks Github imposes and you'll need to delete them manually.

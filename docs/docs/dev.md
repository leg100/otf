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

The documentation uses [Material for MkDocs](https://squidfunk.github.io/mkdocs-material/).

The documentation pages are maintained in the `./docs` directory of the repository. To make small edits it is recommended you click on the `Edit this page` icon (see top right of this page), which'll take you to Github and prompt you to make a pull request.

For larger changes, make the edits locally. Change into the `./docs` directory in your terminal. First you need to install the python package dependencies:

```
make install

```

Then serve the docs:

```
make serve
```

That builds and runs the documentation site on your workstation at `http://localhost:8000`. Any changes you make to the documentation are reflected in real-time in the browser.

Screenshots in the documentation are largely automated. The browser-based integration tests produce screenshots at various steps. If the environment variable `OTF_DOC_SCREENSHOTS=true` is present then such a test also writes the screenshot into the documentation directory. Run the following make task from the **root** of the repo to generate the screenshots:

```
OTF_DOC_SCREENSHOTS=true make test
```

## SQL migrations

The database schema is migrated using [tern](https://github.com/jackc/tern). The SQL migration files are kept in the repo in `./internal/sql/migrations`. Upon startup `otfd` automatically migrates the DB to the latest version.

## HTML path helpers

Rails-style path helpers are generated using `go generate`. The path specifications are maintained in `./ui/paths/paths.yaml`. After making changes to the specs run the following make task to generate the helpers:

* `make paths`

## Web development

If you're making changes to web templates then you may want to enable developer mode: set the environment variable `DEV_MODE=1`. Once enabled you will be able to see changes without restarting `otfd`.

OTF uses [Tailwind CSS](https://tailwindcss.com/) to generate CSS classes. Run the following make task to generate the CSS:

* `make live/tailwind`

!!! note
    To install tailwind first ensure you've installed `npm` and then run `npm install -D tailwindcss`

For templates, OTF uses [Templ](https://templ.guide/), an alternative to go's built in templates that generates go code from a proprietary syntax. Templ provides a handy command to watch for changes and generate go code before reloading your browser to show changes in real-time. A make task is provided to run this command:

* `make live/templ`

The two commands above are combined into the following make task, to watch for all changes to templates, tailwind CSS classes and static assets such as images:

* `make live`

## Helm charts

The helm charts are maintained in the `./charts` directory of the repo. 

### Bumping the chart version

If you make any changes to a chart you need to bump its chart version. You can either do that by hand in `Chart.yaml`, or using `make`:

```bash
# requires `yq`
#
# To update the otfd chart version
CHART=otfd make bump-chart-version
#
# To update the otf-agent chart version
CHART=otf-agent make bump-chart-version
```

### Generating README.md's

Each chart's `README.md` is generated from a template, `README.md.gotmpl` in the same directory, using [helm-docs](https://github.com/norwoodj/helm-docs). Therefore any changes must be made to `README.md.gotmpl` and not `README.md`. To update all templated README.md's, run the following from the root of the repo:

```bash
make helm-docs
```
Any changes to the version or to the `values.yaml` file are automatically reflected in the generated `README.md`. 

### Linting

To lint the charts to check for any errors run `helm lint`:

```bash
# lint the otfd chart
helm lint ./charts/otfd
# lint the otf-agent chart
helm lint ./charts/otf-agent
```

### Deploy and test otfd chart

To deploy the `./charts/otfd` chart to a cluster to the namespace `otfd-test` with pre-configured defaults along with PostgreSQL:

```bash
make deploy-otfd
```

To test the chart (assumes release is named `otfd`):

```bash
make test-otfd
```

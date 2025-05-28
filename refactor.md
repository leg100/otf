* replace kind field constant on vcs providers with an interface.
* interface is implemented by respective kinds
* interface is populated in vcs provider constructor and in database row scanner.
* enable kinds to hook into database row scanner to populate the interface
    * e.g. kinds that rely on a PAT, can read token from vcs provider row
    * e.g. more sophisticated kinds like the github app can perform further database query to retrieve private key etc
* remove github_app_id foreign key from vcs_providers (github_app_installs already has foreign key to vcs_providers)
* move web UI handlers and views to respective implementations:
    * new page
    * edit page
* these pages operate on the vcs_providers config field
* add registration hook(s):
    * register a vcs kind
    * register handlers for new/edit pages
    * register tfe api service_provider
    * register a client constructor that takes round tripper (which can be configured to skip tls verification)
* make existing event handler registration a package-level function.
* don't disable tls verification in (github.service).newClient()

# otf

**otf** is an open-source alternative to Terraform Enterprise:

* Full Terraform CLI integration
* Remote execution mode: plans and applies run on servers
* Agent execution mode: plans and applies run on agents
* Remote state backend: state stored in PostgreSQL
* SSO signin: github and gitlab supported
* Team-based authorization: syncs your github teams / gitlab roles
* Compatible with much of the Terraform Enterprise/Cloud API
* Minimal dependencies: requires only PostgreSQL
* Stateless: horizontally scale servers in pods on Kubernetes, etc

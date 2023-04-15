# Introduction

OTF is an open-source alternative to Terraform Enterprise, sharing many of its features:

* Full Terraform CLI integration
* Remote execution mode: plans and applies run on server
* Agent execution mode: plans and applies run on agents
* Remote state backend: state stored in PostgreSQL
* SSO: sign in using Github and Gitlab
* Organization and team synchronisation from Github and gitlab
* Module registry (provider registry coming soon)
* Authorization: control team access to workspaces
* VCS integration: trigger runs and publish modules from git commits
* Compatible with much of the Terraform Enterprise/Cloud API
* Minimal dependencies: requires only PostgreSQL
* Stateless: horizontally scale servers in pods on Kubernetes, etc

Feel free to trial it using the demo deployment: [https://demo.otf.ninja](https://demo.otf.ninja)

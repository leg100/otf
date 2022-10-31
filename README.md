# otf

An open source alternative to Terraform Enterprise:

* Full Terraform CLI integration
* Remote execution mode: plans and applies run remotely on server
* Agent execution mode: plans and applies run remotely on agents
* Remote state backend: state stored in PostgreSQL
* SSO signin: github and gitlab supported
* Team-based authorization: syncs your github teams / gitlab roles
* Compatible with much of the Terraform Enterprise/Cloud API
* Minimal dependencies: requires only PostgreSQL
* Stateless: horizontally scale servers in pods on Kubernetes, etc

Docs: https://otf-project.readthedocs.io/

## Quickstart Demo

To quickly try out otf you can sign into the demo server using your github account:

https://demo.otf.ninja

Once signed in you'll notice any github organization and team memberships are synchronised across automatically. Additionally, an organization matching your username is created.

Now we'll demonstrate terraform CLI usage. First create a token:

```bash
terraform login demo.otf.ninja
```

Confirm with `yes` to proceed and it'll open a browser window where you can create a token:

<img src="https://user-images.githubusercontent.com/75728/198881088-bb6f83f5-68ce-4a1d-966b-badd6b5340b7.png" width="600">


Click `New token` and give it a description and click `Create token`. The token will be revealed. Click on the token to copy it to your clipboard:

<img src="https://user-images.githubusercontent.com/75728/198881193-591ae9f5-8446-4db4-9861-62682bc69d15.png" width="600">

Now return to your terminal and paste the token into the prompt.

You should then receive successful confirmation:

```
Success! Logged in to Terraform Enterprise (demo.otf.ninja)
```

Now write some terraform configuration to a file:

```terraform
terraform {
  backend "remote" {
    hostname     = "demo.otf.ninja"
    organization = "<your organization>"

    workspaces {
      name = "dev"
    }
  }
}

resource "null_resource" "demo" {}
```

Initialize and run a plan:

```bash
terraform init
terraform plan
```

That starts a run on the server. Click on the link to the run to view status and logs:

<img src="https://user-images.githubusercontent.com/75728/198881848-0d7f42f9-18f7-418d-9474-a828da6982fe.png" width="600">

You optionally run `terraform apply` to apply the changes:

```bash
terraform apply
```

## Getting Started


## TODO

* VCS integration
* Provider and module registry
* Policies (OPA)

## Building

You'll need [Go](https://golang.org/doc/install) installed.

Clone the repo, and then build and install the binary using the make task:

```bash
git clone https://github.com/leg100/otf
cd otf
make install
```

That'll create a binary inside your go bins directory (defaults to `$HOME/go/bin`).

## Intellectual Property

Note: otf is in no way affiliated with Hashicorp. Terraform and Terraform Enterprise are trademarks of Hashicorp.


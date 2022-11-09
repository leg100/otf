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

Docs: https://docs.otf.ninja/

## Quickstart Demo

To quickly try out otf you can sign into the demo server using your github account:

https://demo.otf.ninja

Once signed in you'll notice any github organization and team memberships are synchronised across automatically. Additionally, an organization matching your username is created.

Setup local credentials:

```bash
terraform login demo.otf.ninja
```

Confirm with `yes` to proceed and it'll open a browser window where you can create a token:

Click `New token` and give it a description and click `Create token`. The token will be revealed. Click on the token to copy it to your clipboard.

Return to your terminal and paste the token into the prompt.

You should then receive successful confirmation:

```
Success! Logged in to Terraform Enterprise (demo.otf.ninja)
```

Write some terraform configuration to a file, setting the organization to your username:

```terraform
terraform {
  backend "remote" {
    hostname     = "demo.otf.ninja"
    organization = "<your username>"

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

You can optionally run `terraform apply` to apply the changes:

```bash
terraform apply
```

## Legal

otf is in no way affiliated with Hashicorp. Terraform and Terraform Enterprise are trademarks of Hashicorp.


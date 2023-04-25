![OTF logo](/http/html/static/images/logo_large.svg)

[![Slack](https://img.shields.io/badge/join-slack%20community-brightgreen)](https://join.slack.com/t/otf-pg29376/shared_invite/zt-1jga4k1cl-bzmJg71f4uUB9fJhxdT~gQ)

An open source alternative to Terraform Enterprise. Includes SSO, team management, agents, and as many applies as you can throw hardware at.

Docs: https://docs.otf.ninja/

## Quickstart Demo

To quickly try out otf you can sign into the demo server using your github account:

https://demo.otf.ninja

Once signed in you'll notice any github organization and team memberships are synchronised across automatically. Additionally, an organization matching your username is created.

Now we'll login to the account in your terminal. You'll need terraform installed.

NOTE: only terraform version 1.2.0 and later is supported.

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

Initialize terraform:

```bash
terraform init
```

> NOTE: if you're using Mac or Windows, you may receive an error similar to the following error:
>
> > Error: Failed to install provider
> >
> > Error while installing hashicorp/null v3.2.1: the current package for registry.terraform.io/hashicorp/null 3.2.1
> > doesn't match any of the checksums previously recorded in the dependency lock file; for more information:
> > https://www.terraform.io/language/provider-checksum-verification
>
> If so, you need to update your lockfile (`.terraform.lock.hcl`) to include hashes for the platform that the OTF demo is hosted on (linux):
>
> ```
> terraform providers lock -platform=linux_amd64
> ```
>
> Then re-run `terraform init`

Now create a plan:

```bash
terraform plan
```

After you've invoked `terraform plan`, you'll see the plan output along with a link. Click on the link to the run to view the status and logs:

<img src="https://user-images.githubusercontent.com/75728/198881848-0d7f42f9-18f7-418d-9474-a828da6982fe.png" width="600">

You can optionally run `terraform apply` to apply the changes:

```bash
terraform apply
```

You've reached the end of this quickstart demo. See the [docs](https://docs.otf.ninja) for instructions on deploying OTF.

## Legal

otf is in no way affiliated with Hashicorp. Terraform and Terraform Enterprise are trademarks of Hashicorp.

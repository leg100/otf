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

These steps will get you started with running everything on your local system. You'll setup the server, configure SSL so that terraform trusts the server, and then configure terraform. You'll then be able to run terraform commands using the server as a remote backend.

![demo](https://user-images.githubusercontent.com/75728/133922405-b8474369-28ea-4772-b4bf-9131dc366f1d.gif)

1. Download a [release](https://github.com/leg100/otf/releases). The zip file contains two binaries: a daemon and a client, `otfd` and `otf`. Extract them to a directory in your `PATH`, e.g. `/usr/local/bin`.
1. Generate SSL cert and key. For example, to generate a self-signed cert and key for localhost using `openssl`:

    ```bash
    openssl req -x509 -newkey rsa:4096 -sha256 -keyout key.pem -out cert.crt -days 365 -nodes -subj '/CN=localhost' -addext 'subjectAltName=DNS:localhost'
    ```
    
1. Ensure your system trusts the generated cert. For example, on Linux:

    ```bash
    sudo cp cert.crt /usr/local/share/ca-certificates
    sudo update-ca-certificates
    ```
    
1. Ensure you have access to a postgresql server. otf assumes it's running locally on a unix domain socket in `/var/run/postgresql`. Create a database named `otf`:

    ```bash
    createdb otfd
    ```

1. Run the otf daemon:

    ```bash
    otfd --ssl --cert-file=cert.crt --key-file=key.pem
    ```
   
   The daemon runs in the foreground and can be left to run.

   Note: you can customise the postgres connection string by passing it via the flag `--database`.
      
1. In another terminal, login to your OTF server (this merely adds some dummy credentials to `~/.terraform.d/credentials.tfrc.json`):

   ```bash
   otf login
   ```

1. Configure the terraform backend and define a resource:

    ```bash
    cat > main.tf <<EOF
    terraform {
      backend "remote" {
        hostname = "localhost:8080"
        organization = "default"

        workspaces {
          name = "dev"
        }
      }
    }
    
    resource "null_resource" "e2e" {}
    EOF
    ```
    
1. Run terraform!:

   ```bash
   terraform init
   terraform plan
   terraform apply
   ```

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


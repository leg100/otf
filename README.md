# OTF: Open Terraforming Framework

An open source alternative to terraform enterprise.

Functionality is currently limited:

* Remote execution mode (plans and applies run remotely)
* State backend (state stored on disk)
* Workspace management (supports `terraform workspace` commands)
* No web frontend; CLI/API support only.

## Getting Started

These steps will get you started with running everything on your local system. You'll setup the server, configure SSL so that terraform trusts the server, and then configure terraform. You'll then be able to run terraform commands using the server as a remote backend.

![demo](https://user-images.githubusercontent.com/75728/133922405-b8474369-28ea-4772-b4bf-9131dc366f1d.gif)

1. Download a [release](https://github.com/leg100/otf/releases). The zip file contains two binaries: a daemon and a client, `otfd` and `otf`. Extract them to a directory in your `PATH`, e.g. `/usr/local/bin`.
1. Generate SSL cert and key. For example, to generate a self-signed cert and key for localhost:

    ```bash
    openssl req -x509 -newkey rsa:4096 -sha256 -keyout key.pem -out cert.crt -days 365 -nodes -subj '/CN=localhost' -addext 'subjectAltName=DNS:localhost'
    ```
    
1. Ensure your system trusts the generated cert. For example, on Linux:

    ```bash
    sudo cp cert.crt /usr/local/share/ca-certificates
    sudo update-ca-certificates
    ```
    
1. Run the OTF daemon:

    ```bash
    otfd --ssl --cert-file=cert.crt --key-file=key.pem
    ```
   
   The daemon runs in the foreground and can be left to run.
      
1. In another terminal, login to your OTF server (this merely adds some dummy credentials to `~/.terraform.d/credentials.tfrc.json`):

   ```bash
   otf login
   ```
   
1. Create an organization:

   ```bash
   otf organizations new mycorp --email=sysadmin@mycorp.co
   ```

1. Configure the terraform backend and define a resource:

    ```bash
    cat > main.tf <<EOF
    terraform {
      backend "remote" {
        hostname = "localhost:8080"
        organization = "mycorp"

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

## Next Steps

OTF is a mere prototype but a roadmap of further features is planned:

* User AuthN/Z
* Agents
* Terminal application
* Github integration
* Policies (OPA?)
* Web frontend

## Building

You'll need [Go](https://golang.org/doc/install) installed.

Clone the repo, and then build and install the binary using the make task:

```bash
git clone https://github.com/leg100/otf
cd otf
make install
```

That'll create a binary inside your go bins directory (defaults to `$HOME/go/bin`).


# OTS: Open Terraforming Server

A prototype open source alternative to terraform enterprise.

Functionality is currently limited:

* State backend (state stored in a sqlite database)
* Workspace management (supports `terraform workspace` commands)
* Local execution mode (plans and applies run locally)

## Getting Started

Note to mac users: the darwin release is broken. You'll need to clone this repo and build locally.

These steps will get you started with running everything on your local system. You'll setup the server, configure SSL so that terraform trusts the server, and then configure terraform. You'll then be able to run terraform commands using the server as a remote backend.

![demo](https://user-images.githubusercontent.com/75728/122572684-e21ffc80-d045-11eb-91a7-927d18eb7e62.gif)

1. Download and extract a [release](https://github.com/leg100/ots/releases).
1. Generate SSL cert and key. For example, to generate a self-signed cert and key for localhost:

    ```bash
    openssl req -x509 -newkey rsa:4096 -sha256 -keyout key.pem -out cert.crt -days 365 -nodes -subj '/CN=localhost' -addext 'subjectAltName=DNS:localhost'
    ```
    
1. Ensure your system trusts the generated cert. For example, on Linux:

    ```bash
    sudo cp cert.crt /usr/local/share/ca-certificates
    sudo update-ca-certificates
    
1. Run the OTS daemon:

    ```bash
    ./otsd -ssl -cert-file cert.crt -key-file key.pem
    ```
   
   The daemon runs in the foreground and can be left to run.
   
1. In another terminal create an organization:

   ```bash
   curl -H"Accept: application/vnd.api+json" https://localhost:8080/api/v2/organizations -d'{
     "data": {
       "type": "organizations",
       "attributes": {
         "name": "mycorp",
         "email": "sysadmin@mycorp.co"
       }
     }
   }'
   ```   
1. Enter some dummy credentials (this is necessary otherwise terraform will complain):

   ```bash
   cat > ~/.terraform.d/credentials.tfrc.json <<EOF
   {
     "credentials": {
       "localhost:8080": {
         "token": "dummy"
       }
     }
   }
   EOF
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

OTS is a mere prototype but a roadmap of further features could be:

* User AuthN/Z
* Remote execution mode
* Agents
* Github integration
* Policies (OPA?)
* Web frontend

# OTS: Open Terraforming Server

A terraform server compatible with the terraform cloud API.

## Getting Started

1. Download and extract a [release](https://github.com/leg100/ots/releases).
1. Generate SSL cert and key. For example, to generate a self-signed cert and key for localhost:

    ```bash
    openssl req -x509 -newkey rsa:4096 -sha256 -keyout key.pem -out cert.crt -days 365 -nodes -subj '/CN=localhost' -addext 'subjectAltName=DNS:localhost'
    ```
    
1. Ensure your systems trusts the generated cert. For example, on Linux:

    ```bash
    sudo cp cert.crt /usr/local/share/ca-certificates
    sudo update-ca-certificates
    
1. Run the OTS daemon:

    ```bash
    ./ots -ssl -cert-file cert.crt -key-file key.pem
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
    
1. Configure the terraform backend and define a resource:

    ```bash
    cat > main.tf <EOF
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

    

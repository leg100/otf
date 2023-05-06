terraform {
  backend "remote" {
    hostname = "app.terraform.io"
    organization = "automatize"

    workspaces {
      prefix = "etok-example-"
    }
  }
}

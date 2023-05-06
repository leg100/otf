terraform {
  backend "remote" {
    hostname     = "otf"
    organization = "automatize"

    workspaces {
      prefix = "etok-example-"
    }
  }
}

resource "null_resource" "test" {}

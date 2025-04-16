# State Migration

This is a guide for migrating your existing terraform state into OTF.

## Migrating from Terraform Cloud / Enterprise

If you're currently using Terraform Cloud or Terraform Enterprise, you are
either using the [`remote` backend](https://developer.hashicorp.com/terraform/language/settings/backends/remote) or
the newer [`cloud` block](https://developer.hashicorp.com/terraform/cli/cloud/settings). See the relevant instructions below.

### Cloud block migration

1. If you're using the the newer `cloud` block, your existing configuration will look something like this:

		terraform {
		  cloud {
			hostname = "app.terraform.io"
			organization = "automatize"

			workspaces {
			  name = "my-workspace"
			}
		  }
		}

1. Temporarily update the configuration to use `remote` instead:

		terraform {
		  backend "remote" {
			hostname = "app.terraform.io"
			organization = "automatize"

			workspaces {
			  name = "my-workspace"
			}
		  }
		}

	!!! note
		This step is necessary because `terraform` does not allow state to be
		migrated when using the cloud block configuration. Once you've migrated
		the state you can re-introduce the cloud block (see below).

1. Remove the `.terraform` directory:

		rm -r .terraform

1. Then follow the [Remote backend migration](#remote-backend-migration) instructions below.

### Remote backend migration

1. If you're using the `remote` backend, your existing configuration will look something like this:

		terraform {
		  backend "remote" {
			hostname = "app.terraform.io"
			organization = "automatize"

			workspaces {
			  name = "my-workspace"
			}
		  }
		}

1. To migrate to OTF you only need to update the hostname:

		terraform {
		  backend "remote" {
			hostname = "otf.example.com"
			organization = "automatize"

			workspaces {
			  name = "my-workspace"
			}
		  }
		}

1. Ensure you have credentials for your hostname:

		terraform login otf.example.com

1. And then migrate the state:

		terraform init -migrate-state

	You should see output similar to the following:

		Initializing the backend...
		Backend configuration changed!

		Terraform has detected that the configuration specified for the backend
		has changed. Terraform will now check for existing state in the backends.

		Do you want to copy existing state to the new backend?
		  Pre-existing state was found while migrating the previous "remote" backend to the
		  newly configured "remote" backend. No existing state was found in the newly
		  configured "remote" backend. Do you want to copy this state to the new "remote"
		  backend? Enter "yes" to copy and "no" to start with an empty state.

		  Enter a value: yes


		Successfully configured the backend "remote"! Terraform will automatically
		use this backend unless the backend configuration changes.

1. Optional: you can update the configuration to use the `cloud` block. Doing so allows you to use newer features such as [workspace tags](https://developer.hashicorp.com/terraform/cli/cloud/settings#tags):

		terraform {
		  cloud {
			hostname = "otf.example.com"
			organization = "automatize"

			workspaces {
			  name = "my-workspace"
			}
		  }
		}

1. You'll then need to reinitialize:

		terraform init

1. You'll be prompted to enter `yes` or `no`. Enter `yes` to complete the switch to using the `cloud` block:

		Initializing Terraform Cloud...
		Migrating from backend "remote" to Terraform Cloud.
		Do you wish to proceed?
		  As part of migrating to Terraform Cloud, Terraform can optionally copy your
		  current workspace state to the configured Terraform Cloud workspace.

		  Answer "yes" to copy the latest state snapshot to the configured
		  Terraform Cloud workspace.

		  Answer "no" to ignore the existing state and just activate the configured
		  Terraform Cloud workspace with its existing state, if any.

		  Should Terraform migrate your existing state?

		  Enter a value: yes

	  You should then be informed the migration was successful:

		Initializing provider plugins...
		- Reusing previous version of hashicorp/null from the dependency lock file
		- Using previously-installed hashicorp/null v3.2.1

		Terraform Cloud has been successfully initialized!

	!!! note
		Despite what the output says, `terraform` does not actually copy any state across; your state has already been uploaded to the relevant OTF workspace in a previous step.


## Migrating from other state backends

If you're currently using a configuration other the `remote` backend or the
`cloud` block, e.g.
[s3](https://developer.hashicorp.com/terraform/language/settings/backends/s3) or
[local](https://developer.hashicorp.com/terraform/language/settings/backends/local),
etc., then follow these steps:

1. Replace your existing backend configuration, e.g. `s3`:

		terraform {
		  backend "s3" {
			bucket = "mybucket"
			key    = "path/to/my/key"
			region = "us-east-1"
		  }
		}

	with a `cloud` block for OTF:

		terraform {
		  cloud {
			hostname = "otf.example.com"
			organization = "automatize"

			workspaces {
			  name = "my-workspace"
			}
		  }
		}

	See the [cloud settings documentation](https://developer.hashicorp.com/terraform/cli/cloud/settings) for help with configuration of the `cloud` block.

1. And then reinitialize:

		terraform init

1. You'll be prompted to enter `yes` or `no`. Enter `yes` to complete the migration:

		Initializing Terraform Cloud...
		Do you wish to proceed?
		  As part of migrating to Terraform Cloud, Terraform can optionally copy your
		  current workspace state to the configured Terraform Cloud workspace.

		  Answer "yes" to copy the latest state snapshot to the configured
		  Terraform Cloud workspace.

		  Answer "no" to ignore the existing state and just activate the configured
		  Terraform Cloud workspace with its existing state, if any.

		  Should Terraform migrate your existing state?

		  Enter a value: yes

	  You should then be informed the migration was successful:

		Initializing provider plugins...
		- Reusing previous version of hashicorp/null from the dependency lock file
		- Using previously-installed hashicorp/null v3.2.1

		Terraform Cloud has been successfully initialized!

		You may now begin working with Terraform Cloud. Try running "terraform plan" to
		see any changes that are required for your infrastructure.

		If you ever set or change modules or Terraform Settings, run "terraform init"
		again to reinitialize your working directory.

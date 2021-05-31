# Terraform Provider for Hopsworks.ai

![Unit Tests](https://github.com/logicalclocks/terraform-provider-hopsworksai/actions/workflows/unit-test.yml/badge.svg) 

The Terraform Hopsworks.ai provider is a plugin for Terraform that allows for creating and managing Hopsworks clusters on [Hopsworks.ai](http://managed.hopsworks.ai/)

## Example Usage 

```hcl
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "3.42.0"
    }
    hopsworksai = {
      source  = "logicalclocks/hopsworksai"
      version = "0.1.0"
    }
  }
}

locals {
  region      = "us-east-2"
  bucket_name = "tf-hopsworks-bucket"
}

provider "aws" {
  region = local.region
}

provider "hopsworksai" {
  # Highly recommeneded to use the HOPSWORKSAI_API_KEY environment variable instead
  api_key = "YOUR HOPSWORKS API KEY"
}

# Step 1: create an instance profile to allow hopsworks cluster 
data "hopsworksai_aws_instance_profile_policy" "policy" {
  bucket_name = local.bucket_name
}

resource "aws_iam_role" "role" {
  name = "tf-hopsworksai-instance-profile-role"
  assume_role_policy = jsonencode(
    {
      Version = "2012-10-17"
      Statement = [
        {
          Action = "sts:AssumeRole"
          Effect = "Allow"
          Principal = {
            Service = "ec2.amazonaws.com"
          }
        },
      ]
    }
  )

  inline_policy {
    name   = "hopsworksai"
    policy = data.hopsworksai_aws_instance_profile_policy.policy.json
  }
}

resource "aws_iam_instance_profile" "profile" {
  name = "hopsworksai-instance-profile"
  role = aws_iam_role.role.name
}

# Step 2: create s3 bucket to be used by your hopsworks cluster to store your data
resource "aws_s3_bucket" "bucket" {
  bucket        = local.bucket_name
  acl           = "private"
  force_destroy = true
}

# Step 3: create an ssh key pair 
resource "aws_key_pair" "key" {
  key_name   = "tf-hopsworksai-key"
  public_key = file("~/.ssh/id_rsa.pub")
}

# Step 4: create the cluster
resource "hopsworksai_cluster" "cluster" {
  name    = "tf-hopsworks-cluster"
  ssh_key = aws_key_pair.key.key_name

  head {
  }

  workers {
    count = 1
  }

  aws_attributes {
    region               = local.region
    bucket_name          = local.bucket_name
    instance_profile_arn = aws_iam_instance_profile.profile.arn
  }

  open_ports {
    ssh = true
  }

  tags = {
    Purpose = "testing"
  }
}

```
## Requirements

-	[Terraform](https://www.terraform.io/downloads.html) >= 0.13.x
-	[Go](https://golang.org/doc/install) >= 1.15

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the `make install` command: 
```sh
$ make install
```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `make install`. This will build the provider and put the provider binary in the terraform plugin directory.

To generate or update documentation, run `make generate`.

### Acceptance tests
**Note:** Acceptance tests create real resources, and cost money to run.

In order to run the full suite of Acceptance tests, you need do the following:
* Configure your AWS and Azure credentials locally 
* Export the following environment variables

```sh
   export HOPSWORKSAI_API_KEY=<YOUR HOPSWORKS API KEY>
   export TF_VAR_skip_aws=false # Setting it to true will not run any acceptance tests on AWS
   export TF_VAR_skip_azure=false # Setting it to true will not run any acceptance tests on Azure
   export TF_VAR_azure_resource_group=<YOUR AZURE RESOURCE GROUP> # no need to set if you skip tests on Azure
```
* Run all the acceptance tests using the following command 

```sh
$ make testacc 
```

You can also run only a single test or a some tests following some name pattern as follows
```sh
$ make testacc TESTARGS='-run=TestAcc*'
```

Acceptance tests provision real resources, and ideally these resources should be destroyed at the end of each test, however, it can happen that resources are leaked due to different reasons. For that, you can run the sweeper to clean up all resources created during acceptance testing.

```sh
$ make sweep 
```

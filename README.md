# Terraform Provider for Hopsworks.ai

![Unit Tests](https://github.com/logicalclocks/terraform-provider-hopsworksai/actions/workflows/unit-test.yml/badge.svg) 

- Website: [managed.hopsworks.ai](https://managed.hopsworks.ai/)
- Community: [community.hopsworks.ai](https://community.hopsworks.ai/)

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
  # Highly recommended to use the HOPSWORKSAI_API_KEY environment variable instead
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


## Quick Starts

- [Using the provider](https://registry.terraform.io/providers/logicalclocks/hopsworksai/latest/docs)
- [Provider development](DEVELOPMENT.md)


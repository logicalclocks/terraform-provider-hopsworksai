# Terraform Provider for Hopsworks.ai

![Build and Unit Tests](https://github.com/logicalclocks/terraform-provider-hopsworksai/actions/workflows/unit-test.yml/badge.svg) [![Code Coverage](https://codecov.io/gh/logicalclocks/terraform-provider-hopsworksai/branch/main/graph/badge.svg)](https://codecov.io/gh/logicalclocks/terraform-provider-hopsworksai)

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
      source = "logicalclocks/hopsworksai"
    }
  }
}

variable "region" {
  type    = string
  default = "us-east-2"
}

provider "aws" {
  region = var.region
}

provider "hopsworksai" {
  # Highly recommended to use the HOPSWORKSAI_API_KEY environment variable instead
  api_key = "YOUR HOPSWORKS API KEY"
}

# Step 1: create the required aws resources, an ssh key, an s3 bucket, and an instance profile with the required hopsworks permissions
module "aws" {
  source = "logicalclocks/helpers/hopsworksai//modules/aws"
  region = var.region
}

# Step 2: create a cluster with 1 worker
data "hopsworksai_instance_type" "smallest_worker" {
  cloud_provider = "AWS"
  node_type      = "worker"
}

resource "hopsworksai_cluster" "cluster" {
  name    = "tf-hopsworks-cluster"
  ssh_key = module.aws.ssh_key_pair_name

  head {
  }

  workers {
    instance_type = data.hopsworksai_instance_type.smallest_worker.id
    count         = 1
  }

  aws_attributes {
    region               = var.region
    bucket_name          = module.aws.bucket_name
    instance_profile_arn = module.aws.instance_profile_arn
  }

  open_ports {
    ssh = true
  }
}

# Outputs the url of the newly created cluster 
output "hopsworks_cluster_url" {
  value = hopsworksai_cluster.cluster.url
}
```

## Requirements

-	[Terraform](https://www.terraform.io/downloads.html) >= 0.13.x
-	[Go](https://golang.org/doc/install) >= 1.17
-	[golangci-lint](https://golangci-lint.run/usage/install/)


## Quick Starts

- [Using the provider](https://registry.terraform.io/providers/logicalclocks/hopsworksai/latest/docs)
- [Provider development](DEVELOPMENT.md)


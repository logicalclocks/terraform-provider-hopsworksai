provider "aws" {
  region = var.region
}

provider "hopsworksai" {
}

# Create required aws resources, an ssh key, an s3 bucket, and an instance profile with the required hopsworks permissions
module "aws" {
  source = "logicalclocks/helpers/hopsworksai//modules/aws"
  region = var.region
}

# Create a cluster with rondb default configuration, 1 management node, 2 data nodes, and 1 mysql node
resource "hopsworksai_cluster" "cluster" {
  name    = "tf-hopsworks-cluster"
  ssh_key = module.aws.ssh_key_pair_name

  head {
  }

  aws_attributes {
    region               = var.region
    bucket_name          = module.aws.bucket_name
    instance_profile_arn = module.aws.instance_profile_arn
  }

  rondb {
  }

  open_ports {
    ssh = true
  }
}

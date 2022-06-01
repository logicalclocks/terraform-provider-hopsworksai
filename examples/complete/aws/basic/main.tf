provider "aws" {
  region  = var.region
  profile = var.profile
}

provider "hopsworksai" {
}

# Create required aws resources, an ssh key, an s3 bucket, and an instance profile with the required hopsworks permissions
module "aws" {
  source  = "/Volumes/Code/terraform-hopsworksai-helpers/modules/aws"
  region  = var.region
  version = "2.0.0"
}

# Create a simple cluster with two workers with two different configuration

data "hopsworksai_instance_type" "small_worker" {
  cloud_provider = "AWS"
  node_type      = "worker"
  min_memory_gb  = 16
  min_cpus       = 4
}

data "hopsworksai_instance_type" "large_worker" {
  cloud_provider = "AWS"
  node_type      = "worker"
  min_memory_gb  = 32
  min_cpus       = 4
}

resource "hopsworksai_cluster" "cluster" {
  name    = "tf-hopsworks-cluster"
  ssh_key = module.aws.ssh_key_pair_name

  head {
  }

  workers {
    instance_type = data.hopsworksai_instance_type.small_worker.id
    disk_size     = 256
    count         = 1
  }

  workers {
    instance_type = data.hopsworksai_instance_type.large_worker.id
    disk_size     = 512
    count         = 1
  }

  aws_attributes {
    region = var.region
    bucket {
      name = module.aws.bucket_name
    }
    instance_profile_arn = module.aws.instance_profile_arn
  }

  rondb {
  }

  open_ports {
    ssh = true
  }
}

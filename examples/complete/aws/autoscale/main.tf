provider "aws" {
  region  = var.region
  profile = var.profile
}

provider "hopsworksai" {
}

# Create required aws resources, an ssh key, an s3 bucket, and an instance profile with the required hopsworks permissions
module "aws" {
  source  = "logicalclocks/helpers/hopsworksai//modules/aws"
  region  = var.region
  version = "2.1.0"
}

# Create a simple cluster with autoscale
data "hopsworksai_instance_type" "head" {
  cloud_provider = "AWS"
  node_type      = "head"
  region         = var.region
}

data "hopsworksai_instance_type" "rondb_mgm" {
  cloud_provider = "AWS"
  node_type      = "rondb_management"
  region         = var.region
}

data "hopsworksai_instance_type" "rondb_data" {
  cloud_provider = "AWS"
  node_type      = "rondb_data"
  region         = var.region
}

data "hopsworksai_instance_type" "rondb_mysql" {
  cloud_provider = "AWS"
  node_type      = "rondb_mysql"
  region         = var.region
}

data "hopsworksai_instance_type" "small_worker" {
  cloud_provider = "AWS"
  node_type      = "worker"
  region         = var.region
  min_memory_gb  = 16
  min_cpus       = 4
}

resource "hopsworksai_cluster" "cluster" {
  name    = "tf-hopsworks-cluster"
  ssh_key = module.aws.ssh_key_pair_name

  head {
    instance_type = data.hopsworksai_instance_type.head.id
  }

  autoscale {
    non_gpu_workers {
      instance_type       = data.hopsworksai_instance_type.small_worker.id
      disk_size           = 256
      min_workers         = 0
      max_workers         = 10
      standby_workers     = 0.5
      downscale_wait_time = 300
    }
  }

  aws_attributes {
    region = var.region
    bucket {
      name = module.aws.bucket_name
    }
    instance_profile_arn = module.aws.instance_profile_arn
  }

  rondb {
    management_nodes {
      instance_type = data.hopsworksai_instance_type.rondb_mgm.id
    }
    data_nodes {
      instance_type = data.hopsworksai_instance_type.rondb_data.id
    }
    mysql_nodes {
      instance_type = data.hopsworksai_instance_type.rondb_mysql.id
    }
  }

  open_ports {
    ssh = true
  }
}

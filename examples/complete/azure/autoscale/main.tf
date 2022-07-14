provider "azurerm" {
  features {}
  skip_provider_registration = true
}

provider "hopsworksai" {
}

# Create required azure resources, an ssh key, a storage account, and an user assigned managed identity with the required hopsworks permissions
module "azure" {
  source         = "logicalclocks/helpers/hopsworksai//modules/azure"
  resource_group = var.resource_group
  version        = "2.0.0"
}

# Create a simple cluster with autoscale

data "hopsworksai_instance_type" "small_worker" {
  cloud_provider = "AZURE"
  node_type      = "worker"
  region         = module.azure.location
  min_memory_gb  = 16
  min_cpus       = 4
}

resource "hopsworksai_cluster" "cluster" {
  name    = "tf-hopsworks-cluster"
  ssh_key = module.azure.ssh_key_pair_name

  head {
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

  azure_attributes {
    location                       = module.azure.location
    resource_group                 = module.azure.resource_group
    user_assigned_managed_identity = module.azure.user_assigned_identity_name
    container {
      storage_account = module.azure.storage_account_name
    }
  }

  rondb {

  }

  open_ports {
    ssh = true
  }
}

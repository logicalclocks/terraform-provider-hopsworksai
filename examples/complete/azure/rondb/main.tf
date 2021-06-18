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
}

# Create a cluster with rondb default configuration, 1 management node, 2 data nodes, and 1 mysql node
resource "hopsworksai_cluster" "cluster" {
  name    = "tf-hopsworks-cluster"
  ssh_key = module.azure.ssh_key_pair_name

  head {
  }

  azure_attributes {
    location                       = module.azure.location
    resource_group                 = module.azure.resource_group
    storage_account                = module.azure.storage_account_name
    user_assigned_managed_identity = module.azure.user_assigned_identity_name
  }

  rondb {
  }

  open_ports {
    ssh = true
  }
}

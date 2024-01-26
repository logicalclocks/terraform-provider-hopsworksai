terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "3.8.0"
    }
    hopsworksai = {
      source = "logicalclocks/hopsworksai"
    }
  }
}

variable "resource_group" {
  type    = string
  default = "YOUR AZURE RESOURCE GROUP"
}

provider "azurerm" {
  features {}
  skip_provider_registration = true
}

provider "hopsworksai" {
  # Highly recommended to use the HOPSWORKSAI_API_KEY environment variable instead
  api_key = "YOUR HOPSWORKS API KEY"
}

data "azurerm_resource_group" "rg" {
  name = var.resource_group
}

# Step 1: create the required azure resources, an ssh key, a storage account, and an user assigned managed identity with the required hopsworks permissions
module "azure" {
  source         = "logicalclocks/helpers/hopsworksai//modules/azure"
  resource_group = var.resource_group
  version        = "2.3.0"
}

# Step 2: create a cluster with 1 worker

data "hopsworksai_instance_type" "head" {
  cloud_provider = "AZURE"
  node_type      = "head"
  region         = module.azure.location
}

data "hopsworksai_instance_type" "rondb_data" {
  cloud_provider = "AZURE"
  node_type      = "rondb_data"
  region         = module.azure.location
}

data "hopsworksai_instance_type" "smallest_worker" {
  cloud_provider = "AZURE"
  node_type      = "worker"
  region         = module.azure.location
}

resource "azurerm_container_registry" "acr" {
  name                = "tfhopsworksbasic"
  resource_group_name = module.azure.resource_group
  location            = module.azure.location
  sku                 = "Premium"
  admin_enabled       = false
  retention_policy {
    enabled = true
    days    = 7
  }
}

resource "hopsworksai_cluster" "cluster" {
  name    = "tf-hopsworks-cluster"
  ssh_key = module.azure.ssh_key_pair_name

  head {
    instance_type = data.hopsworksai_instance_type.head.id
  }

  workers {
    instance_type = data.hopsworksai_instance_type.smallest_worker.id
    count         = 1
  }

  azure_attributes {
    location                       = module.azure.location
    resource_group                 = module.azure.resource_group
    user_assigned_managed_identity = module.azure.user_assigned_identity_name
    container {
      storage_account = module.azure.storage_account_name
    }
    acr_registry_name = azurerm_container_registry.acr.name
  }

  rondb {
    single_node {
      instance_type = data.hopsworksai_instance_type.rondb_data.id
    }
  }

  open_ports {
    ssh = true
  }
}

# Outputs the url of the newly created cluster 
output "hopsworks_cluster_url" {
  value = hopsworksai_cluster.cluster.url
}
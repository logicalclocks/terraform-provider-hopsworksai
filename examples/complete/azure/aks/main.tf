provider "azurerm" {
  features {}
  skip_provider_registration = true
}

provider "hopsworksai" {
}

data "azurerm_resource_group" "rg" {
  name = var.resource_group
}

# Step 1: create required azure resources, an ssh key, a storage account, and an user assigned managed identity with the required hopsworks permissions
module "azure" {
  source         = "logicalclocks/helpers/hopsworksai//modules/azure"
  resource_group = var.resource_group
  version        = "2.0.0"
}

# Step 2: create virtual network and two subnets, one for AKS and one for Hopsworks
resource "azurerm_virtual_network" "vnet" {
  name                = "hopsworksai-vnet"
  location            = data.azurerm_resource_group.rg.location
  resource_group_name = data.azurerm_resource_group.rg.name
  address_space       = ["10.240.0.0/16"]
}

resource "azurerm_subnet" "aks_subnet" {
  name                                           = "hopsworksai-aks-subnet"
  resource_group_name                            = data.azurerm_resource_group.rg.name
  virtual_network_name                           = azurerm_virtual_network.vnet.name
  address_prefixes                               = ["10.240.1.0/24"]
  enforce_private_link_endpoint_network_policies = true
}

resource "azurerm_subnet" "hopsworksai_subnet" {
  name                 = "hopsworksai-subnet"
  resource_group_name  = data.azurerm_resource_group.rg.name
  virtual_network_name = azurerm_virtual_network.vnet.name
  address_prefixes     = ["10.240.2.0/24"]
}

# Step 3: create an AKS cluster
resource "azurerm_kubernetes_cluster" "aks" {
  name                = "hopsworksai-aks"
  location            = data.azurerm_resource_group.rg.location
  resource_group_name = data.azurerm_resource_group.rg.name
  dns_prefix          = "hopsworksai-aks-dns"

  default_node_pool {
    name            = "default"
    node_count      = 2
    vm_size         = "Standard_DS2_v2"
    os_disk_size_gb = 100
    vnet_subnet_id  = azurerm_subnet.aks_subnet.id
    zones           = [1, 2, 3]
  }

  identity {
    type = "SystemAssigned"
  }

  role_based_access_control_enabled = true

  network_profile {
    network_plugin = "azure"
  }

  private_cluster_enabled = true
}

# Step 4: create an ACR registry 
resource "azurerm_container_registry" "acr" {
  name                = "hopsworksaiacr"
  resource_group_name = data.azurerm_resource_group.rg.name
  location            = data.azurerm_resource_group.rg.location
  sku                 = "Premium"
  admin_enabled       = false
  retention_policy {
    enabled = true
    days    = 7
  }
}

# Step 5: allow AKS nodes to pull images from ACR
resource "azurerm_role_assignment" "aks_to_acr" {
  scope                = azurerm_container_registry.acr.id
  role_definition_name = "AcrPull"
  principal_id         = azurerm_kubernetes_cluster.aks.kubelet_identity.0.object_id
}


# Step 6: create an ssh key
resource "azurerm_ssh_public_key" "key" {
  name                = "hopsworksai-key"
  resource_group_name = data.azurerm_resource_group.rg.name
  location            = data.azurerm_resource_group.rg.location
  public_key          = file("~/.ssh/id_rsa.pub")
}

# Step 7: create a Hopsworks cluster with 1 worker
data "hopsworksai_instance_type" "smallest_worker" {
  cloud_provider = "AZURE"
  node_type      = "worker"
  region         = data.azurerm_resource_group.rg.location
  min_cpus       = 8
}

resource "hopsworksai_cluster" "cluster" {
  name    = "hopsworks-cluster"
  ssh_key = module.azure.ssh_key_pair_name

  head {
  }

  workers {
    instance_type = data.hopsworksai_instance_type.smallest_worker.id
    disk_size     = 256
    count         = 1
  }

  azure_attributes {
    location                       = data.azurerm_resource_group.rg.location
    resource_group                 = data.azurerm_resource_group.rg.name
    user_assigned_managed_identity = module.azure.user_assigned_identity_name
    container {
      storage_account = module.azure.storage_account_name
    }
    network {
      virtual_network_name = azurerm_virtual_network.vnet.name
      subnet_name          = azurerm_subnet.hopsworksai_subnet.name
    }
    aks_cluster_name  = azurerm_kubernetes_cluster.aks.name
    acr_registry_name = azurerm_container_registry.acr.name
  }

  rondb {

  }

  open_ports {
    ssh = true
  }

  tags = {
    Purpose = "testing"
  }
}
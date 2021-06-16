terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "2.63.0"
    }
    hopsworksai = {
      source = "logicalclocks/hopsworksai"
    }
  }
}

variable "azure_resource_group" {
  description = "The resource group where we will create an AKS cluster, an ACR registry, and a Hopsworks cluster."
  type        = string
}

provider "azurerm" {
  features {}
  skip_provider_registration = true
}

provider "hopsworksai" {
}

data "azurerm_resource_group" "rg" {
  name = var.azure_resource_group
}

# Step 1: create storage account
resource "azurerm_storage_account" "storage" {
  name                     = "hopsworksaistorage"
  resource_group_name      = data.azurerm_resource_group.rg.name
  location                 = data.azurerm_resource_group.rg.location
  account_tier             = "Standard"
  account_replication_type = "RAGRS"
}

# Step 2: create user assigned identity with hopsworks ai permissions
resource "azurerm_user_assigned_identity" "identity" {
  resource_group_name = data.azurerm_resource_group.rg.name
  location            = data.azurerm_resource_group.rg.location
  name                = "hopsworksai-identity"
}

# Storage permissions
data "hopsworksai_azure_user_assigned_identity_permissions" "storage_policy" {
  enable_storage = true
  enable_backup  = true
}

resource "azurerm_role_definition" "storage_role" {
  name  = "hopsworksai-identity-role"
  scope = azurerm_storage_account.storage.id
  permissions {
    actions          = data.hopsworksai_azure_user_assigned_identity_permissions.storage_policy.actions
    not_actions      = data.hopsworksai_azure_user_assigned_identity_permissions.storage_policy.not_actions
    data_actions     = data.hopsworksai_azure_user_assigned_identity_permissions.storage_policy.data_actions
    not_data_actions = data.hopsworksai_azure_user_assigned_identity_permissions.storage_policy.not_data_actions
  }
}

resource "azurerm_role_assignment" "storage_role_assignment" {
  scope              = azurerm_storage_account.storage.id
  role_definition_id = azurerm_role_definition.storage_role.role_definition_resource_id
  principal_id       = azurerm_user_assigned_identity.identity.principal_id
}

# AKS and ACR permissions 
data "hopsworksai_azure_user_assigned_identity_permissions" "aks_acr_policy" {
  enable_storage     = false
  enable_backup      = false
  enable_aks_and_acr = true
}

resource "azurerm_role_definition" "aks_acr_role" {
  name  = "hopsworksai-identity-aks-role"
  scope = data.azurerm_resource_group.rg.id
  permissions {
    actions          = data.hopsworksai_azure_user_assigned_identity_permissions.aks_acr_policy.actions
    not_actions      = data.hopsworksai_azure_user_assigned_identity_permissions.aks_acr_policy.not_actions
    data_actions     = data.hopsworksai_azure_user_assigned_identity_permissions.aks_acr_policy.data_actions
    not_data_actions = data.hopsworksai_azure_user_assigned_identity_permissions.aks_acr_policy.not_data_actions
  }
}

resource "azurerm_role_assignment" "aks_acr_role_assignment" {
  scope              = data.azurerm_resource_group.rg.id
  role_definition_id = azurerm_role_definition.aks_acr_role.role_definition_resource_id
  principal_id       = azurerm_user_assigned_identity.identity.principal_id
}


# Step 3: create virtual network and two subnets, one for AKS and one for Hopsworks
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

# Step 4: create an AKS cluster
resource "azurerm_kubernetes_cluster" "aks" {
  name                = "hopsworksai-aks"
  location            = data.azurerm_resource_group.rg.location
  resource_group_name = data.azurerm_resource_group.rg.name
  dns_prefix          = "hopsworksai-aks-dns"

  default_node_pool {
    name               = "default"
    node_count         = 2
    vm_size            = "Standard_DS2_v2"
    os_disk_size_gb    = 100
    vnet_subnet_id     = azurerm_subnet.aks_subnet.id
    availability_zones = [1, 2, 3]
  }

  identity {
    type = "SystemAssigned"
  }

  role_based_access_control {
    enabled = true
  }

  network_profile {
    network_plugin = "azure"
  }

  private_cluster_enabled = true
}

# Step 5: create an ACR registry 
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

# Step 6: allow AKS nodes to pull images from ACR
resource "azurerm_role_assignment" "aks_to_acr" {
  scope                = azurerm_container_registry.acr.id
  role_definition_name = "AcrPull"
  principal_id         = azurerm_kubernetes_cluster.aks.kubelet_identity.0.object_id
}


# Step 7: create an ssh key
resource "azurerm_ssh_public_key" "key" {
  name                = "hopsworksai-key"
  resource_group_name = data.azurerm_resource_group.rg.name
  location            = data.azurerm_resource_group.rg.location
  public_key          = file("~/.ssh/id_rsa.pub")
}

# Step 8: create a Hopsworks cluster with 1 worker
data "hopsworksai_instance_type" "smallest_worker" {
  cloud_provider = "AZURE"
  node_type      = "worker"
}

resource "hopsworksai_cluster" "cluster" {
  name    = "hopsworks-cluster"
  ssh_key = azurerm_ssh_public_key.key.name

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
    storage_account                = azurerm_storage_account.storage.name
    user_assigned_managed_identity = azurerm_user_assigned_identity.identity.name
    network {
      virtual_network_name = azurerm_virtual_network.vnet.name
      subnet_name          = azurerm_subnet.hopsworksai_subnet.name
    }
    aks_cluster_name  = azurerm_kubernetes_cluster.aks.name
    acr_registry_name = azurerm_container_registry.acr.name
  }

  open_ports {
    ssh = true
  }

  tags = {
    Purpose = "testing"
  }
}
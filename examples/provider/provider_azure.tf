terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "2.60.0"
    }
    hopsworksai = {
      source = "logicalclocks/hopsworksai"
    }
  }
}

locals {
  resource_group = "YOUR AZURE RESOURCE GROUP"
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
  name = local.resource_group
}

# Step 1: create storage account
resource "azurerm_storage_account" "storage" {
  name                     = "tfhopsworksstorage"
  resource_group_name      = data.azurerm_resource_group.rg.name
  location                 = data.azurerm_resource_group.rg.location
  account_tier             = "Standard"
  account_replication_type = "RAGRS"
}

# Step 2: create user assigned identity with hopsworks ai permissions
data "hopsworksai_azure_user_assigned_identity_permissions" "policy" {

}

resource "azurerm_user_assigned_identity" "identity" {
  resource_group_name = data.azurerm_resource_group.rg.name
  location            = data.azurerm_resource_group.rg.location
  name                = "tf-hopsworksai-identity"
}

resource "azurerm_role_definition" "storage_role" {
  name  = "tf-hopsworksai-identity-role"
  scope = azurerm_storage_account.storage.id
  permissions {
    actions          = data.hopsworksai_azure_user_assigned_identity_permissions.policy.actions
    not_actions      = data.hopsworksai_azure_user_assigned_identity_permissions.policy.not_actions
    data_actions     = data.hopsworksai_azure_user_assigned_identity_permissions.policy.data_actions
    not_data_actions = data.hopsworksai_azure_user_assigned_identity_permissions.policy.not_data_actions
  }
}

resource "azurerm_role_assignment" "storage_role_assignment" {
  scope              = azurerm_storage_account.storage.id
  role_definition_id = azurerm_role_definition.storage_role.role_definition_resource_id
  principal_id       = azurerm_user_assigned_identity.identity.principal_id
}

# Step 3: create an ssh key
resource "azurerm_ssh_public_key" "key" {
  name                = "tf-hopsworksai-key"
  resource_group_name = data.azurerm_resource_group.rg.name
  location            = data.azurerm_resource_group.rg.location
  public_key          = file("~/.ssh/id_rsa.pub")
}

# Step 4: create the cluster
resource "hopsworksai_cluster" "cluster" {
  name    = "tfhopsworkscluster"
  ssh_key = azurerm_ssh_public_key.key.name

  head {
  }

  workers {
    count = 1
  }

  azure_attributes {
    location                       = data.azurerm_resource_group.rg.location
    resource_group                 = data.azurerm_resource_group.rg.name
    storage_account                = azurerm_storage_account.storage.name
    user_assigned_managed_identity = azurerm_user_assigned_identity.identity.name
  }

  open_ports {
    ssh = true
  }

  tags = {
    Purpose = "testing"
  }
}

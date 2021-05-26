data "hopsworksai_azure_user_assigned_identity_permissions" "policy" {
  enable_upgrade = false
}

data "azurerm_resource_group" "rg" {
  name = var.resource_group
}

resource "azurerm_storage_account" "storage" {
  name                     = var.storage_account_name
  resource_group_name      = data.azurerm_resource_group.rg.name
  location                 = data.azurerm_resource_group.rg.location
  account_tier             = "Standard"
  account_replication_type = "RAGRS"

  tags = {
    Creator = "Terraform"
    Purpose = "Hopsworks.ai"
  }
}

resource "azurerm_user_assigned_identity" "identity" {
  resource_group_name = data.azurerm_resource_group.rg.name
  location            = data.azurerm_resource_group.rg.location
  name                = var.user_assigned_identity_name

  tags = {
    Creator = "Terraform"
    Purpose = "Hopsworks.ai"
  }
}

resource "azurerm_role_definition" "storage_role" {
  name        = "${var.user_assigned_identity_name}-role"
  scope       = azurerm_storage_account.storage.id
  description = "This is a custom role created via Terraform"

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

resource "azurerm_ssh_public_key" "key" {
  name                = var.ssh_key_name
  resource_group_name = data.azurerm_resource_group.rg.name
  location            = data.azurerm_resource_group.rg.location
  public_key          = var.ssh_public_key
}
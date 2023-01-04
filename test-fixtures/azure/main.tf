data "hopsworksai_azure_user_assigned_identity_permissions" "policy" {
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

resource "azurerm_virtual_network" "vnet" {
  name                = var.virtual_network_name
  location            = data.azurerm_resource_group.rg.location
  resource_group_name = data.azurerm_resource_group.rg.name
  address_space       = ["10.240.0.0/16"]
}

resource "azurerm_subnet" "subnet" {
  name                 = "${var.virtual_network_name}-subnet"
  resource_group_name  = data.azurerm_resource_group.rg.name
  virtual_network_name = azurerm_virtual_network.vnet.name
  address_prefixes     = ["10.240.1.0/24"]
}

resource "azurerm_network_security_group" "security_group" {
  name                = "${var.virtual_network_name}-security-group"
  location            = data.azurerm_resource_group.rg.location
  resource_group_name = data.azurerm_resource_group.rg.name

  security_rule {
    name                       = "Http"
    priority                   = 100
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "80"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }

  security_rule {
    name                       = "Https"
    priority                   = 110
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "443"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }
}

# add aks/acr permissions on the resource group
data "hopsworksai_azure_user_assigned_identity_permissions" "resource_group_policy" {
  enable_storage = false
  enable_backup  = false
  enable_aks     = true
  enable_acr     = true
}

resource "azurerm_role_definition" "rg_role" {
  name  = "${var.user_assigned_identity_name}-rg-role"
  scope = data.azurerm_resource_group.rg.id
  permissions {
    actions          = data.hopsworksai_azure_user_assigned_identity_permissions.resource_group_policy.actions
    not_actions      = data.hopsworksai_azure_user_assigned_identity_permissions.resource_group_policy.not_actions
    data_actions     = data.hopsworksai_azure_user_assigned_identity_permissions.resource_group_policy.data_actions
    not_data_actions = data.hopsworksai_azure_user_assigned_identity_permissions.resource_group_policy.not_data_actions
  }
}

resource "azurerm_role_assignment" "rg_role_assignment" {
  scope              = data.azurerm_resource_group.rg.id
  role_definition_id = azurerm_role_definition.rg_role.role_definition_resource_id
  principal_id       = azurerm_user_assigned_identity.identity.principal_id
}

resource "azurerm_container_registry" "acr" {
  name                = var.acr_registry_name
  resource_group_name = data.azurerm_resource_group.rg.name
  location            = data.azurerm_resource_group.rg.location
  sku                 = "Premium"
  admin_enabled       = false
  retention_policy {
    enabled = true
    days    = 7
  }
}
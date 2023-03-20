// In the following locals block, we define all ACR and AKS references that are needed to correctly create a Hopsworks cluster with AKS integration
locals {
  ext_aks_resource_group_name  = data.azurerm_resource_group.aks_resource_group.name
  ext_aks_cluster_name         = azurerm_kubernetes_cluster.aks.name
  ext_aks_cluster_id           = azurerm_kubernetes_cluster.aks.id
  ext_acr_registry_name        = azurerm_container_registry.acr.name
  ext_acr_registry_id          = azurerm_container_registry.acr.id
  ext_aks_virtual_network_id   = azurerm_virtual_network.aks_vnet.id
  ext_aks_virtual_network_name = azurerm_virtual_network.aks_vnet.name

  ext_aks_generated_resource_group = azurerm_kubernetes_cluster.aks.node_resource_group
  ext_aks_private_dns_zone_name    = data.azurerm_private_dns_zone.aks_private_dns_zone.name
}

data "azurerm_resource_group" "hopsworks_resource_group" {
  name = var.hopsworks_resource_group
}

# Create a Hopsworks virtual network, subnet and security group with port 80 and 443 open
resource "azurerm_virtual_network" "hopsworks_vnet" {
  name                = "hopsworks-vnet-${local.suffix_string}"
  location            = data.azurerm_resource_group.hopsworks_resource_group.location
  resource_group_name = data.azurerm_resource_group.hopsworks_resource_group.name
  // Notice this address space shouldn't overlap with the address space used in your AKS virtual network 
  address_space = ["172.18.0.0/16"]
}

resource "azurerm_subnet" "hopsworks_subnet" {
  name                 = "hopsworks-subnet-${local.suffix_string}"
  resource_group_name  = data.azurerm_resource_group.hopsworks_resource_group.name
  virtual_network_name = azurerm_virtual_network.hopsworks_vnet.name
  address_prefixes     = ["172.18.0.0/24"]
}

resource "azurerm_network_security_group" "hopsworks_security_group" {
  name                = "hopsworks-security-group-${local.suffix_string}"
  location            = data.azurerm_resource_group.hopsworks_resource_group.location
  resource_group_name = data.azurerm_resource_group.hopsworks_resource_group.name

  security_rule {
    name                       = "HTTP"
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
    name                       = "HTTPS"
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

resource "azurerm_subnet_network_security_group_association" "hopsworks_security_group_association" {
  subnet_id                 = azurerm_subnet.hopsworks_subnet.id
  network_security_group_id = azurerm_network_security_group.hopsworks_security_group.id
}

// Setup peering between the AKS virtual network and Hopsworks virtual network

resource "azurerm_virtual_network_peering" "hopsworks_to_aks_peering" {
  name                      = "aks-to-hopsworks-peering-${local.suffix_string}"
  resource_group_name       = data.azurerm_resource_group.hopsworks_resource_group.name
  virtual_network_name      = azurerm_virtual_network.hopsworks_vnet.name
  remote_virtual_network_id = local.ext_aks_virtual_network_id

  allow_virtual_network_access = true
  allow_forwarded_traffic      = false
}

resource "azurerm_virtual_network_peering" "aks_to_hopsworks_peering" {
  name                      = "aks-to-hopsworks-peering-${local.suffix_string}"
  resource_group_name       = local.ext_aks_resource_group_name
  virtual_network_name      = local.ext_aks_virtual_network_name
  remote_virtual_network_id = azurerm_virtual_network.hopsworks_vnet.id

  allow_virtual_network_access = true
  allow_forwarded_traffic      = false
}

// set up a DNS private link to Hopsworks virtual network

resource "azurerm_private_dns_zone_virtual_network_link" "hopsworks_aks_private_dns_link" {
  name                  = "hopsworks-aks-dns-${local.suffix_string}"
  resource_group_name   = local.ext_aks_generated_resource_group
  private_dns_zone_name = local.ext_aks_private_dns_zone_name
  virtual_network_id    = azurerm_virtual_network.hopsworks_vnet.id
}

# Create a new Storage account for Hopsworks 
resource "azurerm_storage_account" "hopsworks_storage" {
  name                     = "hopsworks${local.suffix_string}"
  resource_group_name      = data.azurerm_resource_group.hopsworks_resource_group.name
  location                 = data.azurerm_resource_group.hopsworks_resource_group.location
  account_tier             = "Standard"
  account_replication_type = "RAGRS"
}


# Create a user assigned identity with hopsworks.ai permissions
resource "azurerm_user_assigned_identity" "hopsworks_identity" {
  name                = "hopsworks-identity-${local.suffix_string}"
  resource_group_name = data.azurerm_resource_group.hopsworks_resource_group.name
  location            = data.azurerm_resource_group.hopsworks_resource_group.location
}

# Add permissions to the user assigned identity

# add storage permissions on the storage account 
data "hopsworksai_azure_user_assigned_identity_permissions" "hopsworks_storage_policy" {
  enable_storage = true
  enable_backup  = true
  enable_aks     = false
  enable_acr     = false
}

resource "azurerm_role_definition" "hopsworks_storage_role" {
  name  = "hopsworks-identity-storage-role-${local.suffix_string}"
  scope = azurerm_storage_account.hopsworks_storage.id
  permissions {
    actions          = data.hopsworksai_azure_user_assigned_identity_permissions.hopsworks_storage_policy.actions
    not_actions      = data.hopsworksai_azure_user_assigned_identity_permissions.hopsworks_storage_policy.not_actions
    data_actions     = data.hopsworksai_azure_user_assigned_identity_permissions.hopsworks_storage_policy.data_actions
    not_data_actions = data.hopsworksai_azure_user_assigned_identity_permissions.hopsworks_storage_policy.not_data_actions
  }
}

resource "azurerm_role_assignment" "hopsworks_storage_role_assignment" {
  scope              = azurerm_storage_account.hopsworks_storage.id
  role_definition_id = azurerm_role_definition.hopsworks_storage_role.role_definition_resource_id
  principal_id       = azurerm_user_assigned_identity.hopsworks_identity.principal_id
}


# add acr permissions on the ext acr registry 
data "hopsworksai_azure_user_assigned_identity_permissions" "hopsworks_acr_policy" {
  enable_storage = false
  enable_backup  = false
  enable_acr     = true
  enable_aks     = false
}

resource "azurerm_role_definition" "hopsworks_acr_role" {
  name  = "hopsworks-identity-acr-role-${local.suffix_string}"
  scope = local.ext_acr_registry_id
  permissions {
    actions          = data.hopsworksai_azure_user_assigned_identity_permissions.hopsworks_acr_policy.actions
    not_actions      = data.hopsworksai_azure_user_assigned_identity_permissions.hopsworks_acr_policy.not_actions
    data_actions     = data.hopsworksai_azure_user_assigned_identity_permissions.hopsworks_acr_policy.data_actions
    not_data_actions = data.hopsworksai_azure_user_assigned_identity_permissions.hopsworks_acr_policy.not_data_actions
  }
}

resource "azurerm_role_assignment" "hopsworks_acr_role_assignment" {
  scope              = local.ext_acr_registry_id
  role_definition_id = azurerm_role_definition.hopsworks_acr_role.role_definition_resource_id
  principal_id       = azurerm_user_assigned_identity.hopsworks_identity.principal_id
}


# add aks permissions on the ext aks cluster 
data "hopsworksai_azure_user_assigned_identity_permissions" "hopsworks_aks_policy" {
  enable_storage = false
  enable_backup  = false
  enable_acr     = false
  enable_aks     = true
}

resource "azurerm_role_definition" "hopsworks_aks_role" {
  name  = "hopsworks-identity-aks-role-${local.suffix_string}"
  scope = local.ext_aks_cluster_id
  permissions {
    actions          = data.hopsworksai_azure_user_assigned_identity_permissions.hopsworks_aks_policy.actions
    not_actions      = data.hopsworksai_azure_user_assigned_identity_permissions.hopsworks_aks_policy.not_actions
    data_actions     = data.hopsworksai_azure_user_assigned_identity_permissions.hopsworks_aks_policy.data_actions
    not_data_actions = data.hopsworksai_azure_user_assigned_identity_permissions.hopsworks_aks_policy.not_data_actions
  }
}

resource "azurerm_role_assignment" "hopsworks_aks_role_assignment" {
  scope              = local.ext_aks_cluster_id
  role_definition_id = azurerm_role_definition.hopsworks_aks_role.role_definition_resource_id
  principal_id       = azurerm_user_assigned_identity.hopsworks_identity.principal_id
}

# Create an ssh key pair 
resource "azurerm_ssh_public_key" "hopsworks_ssh_key" {
  name                = "hopsworks-key-${local.suffix_string}"
  resource_group_name = data.azurerm_resource_group.hopsworks_resource_group.name
  location            = data.azurerm_resource_group.hopsworks_resource_group.location
  public_key          = file("~/.ssh/id_rsa.pub")
}

# Step 7: create a Hopsworks cluster with 1 worker

data "hopsworksai_instance_type" "head" {
  cloud_provider = "AZURE"
  node_type      = "head"
  region         = data.azurerm_resource_group.hopsworks_resource_group.location
}

data "hopsworksai_instance_type" "rondb_mgm" {
  cloud_provider = "AZURE"
  node_type      = "rondb_management"
  region         = data.azurerm_resource_group.hopsworks_resource_group.location
}

data "hopsworksai_instance_type" "rondb_data" {
  cloud_provider = "AZURE"
  node_type      = "rondb_data"
  region         = data.azurerm_resource_group.hopsworks_resource_group.location
  min_cpus       = 4
  min_memory_gb  = 32
}

data "hopsworksai_instance_type" "rondb_mysql" {
  cloud_provider = "AZURE"
  node_type      = "rondb_mysql"
  region         = data.azurerm_resource_group.hopsworks_resource_group.location
}

data "hopsworksai_instance_type" "smallest_worker" {
  cloud_provider = "AZURE"
  node_type      = "worker"
  region         = data.azurerm_resource_group.hopsworks_resource_group.location
  min_cpus       = 8
}

resource "hopsworksai_cluster" "cluster" {
  #count = 0
  name    = "cluster-${local.suffix_string}"
  version = "3.1.0"
  ssh_key = azurerm_ssh_public_key.hopsworks_ssh_key.name

  head {
    instance_type = data.hopsworksai_instance_type.head.id
    disk_size     = 512
  }

  azure_attributes {
    location                       = data.azurerm_resource_group.hopsworks_resource_group.location
    resource_group                 = data.azurerm_resource_group.hopsworks_resource_group.name
    user_assigned_managed_identity = azurerm_user_assigned_identity.hopsworks_identity.name
    container {
      storage_account = azurerm_storage_account.hopsworks_storage.name
    }
    network {
      virtual_network_name = azurerm_virtual_network.hopsworks_vnet.name
      subnet_name          = azurerm_subnet.hopsworks_subnet.name
      security_group_name  = azurerm_network_security_group.hopsworks_security_group.name
    }
    aks_cluster_name  = local.ext_aks_cluster_name
    acr_registry_name = local.ext_acr_registry_name
  }

  rondb {
    configuration {
      ndbd_default {
        replication_factor = 2
      }
    }
    management_nodes {
      instance_type = data.hopsworksai_instance_type.rondb_mgm.id
      disk_size     = 30
    }
    data_nodes {
      instance_type = data.hopsworksai_instance_type.rondb_data.id
      disk_size     = 512
      count         = 2
    }
    mysql_nodes {
      instance_type = data.hopsworksai_instance_type.rondb_mysql.id
      disk_size     = 256
      count         = 1
    }
  }

  autoscale {
    non_gpu_workers {
      instance_type       = data.hopsworksai_instance_type.smallest_worker.id
      disk_size           = 512
      min_workers         = 0
      max_workers         = 5
      downscale_wait_time = 300
      standby_workers     = 0.5
    }
  }

  tags = {
    Purpose = "testing"
  }
}

data "azurerm_resource_group" "aks_resource_group" {
  name = var.aks_resource_group
}

# Create an AKS vnet 
resource "azurerm_virtual_network" "aks_vnet" {
  name                = "aks-vnet-${local.suffix_string}"
  location            = data.azurerm_resource_group.aks_resource_group.location
  resource_group_name = data.azurerm_resource_group.aks_resource_group.name
  address_space       = ["10.240.0.0/16"]
}

resource "azurerm_subnet" "aks_subnet" {
  name                                           = "aks-subnet-${local.suffix_string}"
  resource_group_name                            = data.azurerm_resource_group.aks_resource_group.name
  virtual_network_name                           = azurerm_virtual_network.aks_vnet.name
  address_prefixes                               = ["10.240.1.0/24"]
  enforce_private_link_endpoint_network_policies = true
}


// Create a private AKS cluster - In Hopsworks, we currently only support System-assigned managed identity based authentication 
resource "azurerm_kubernetes_cluster" "aks" {
  name                = "aks${local.suffix_string}"
  location            = data.azurerm_resource_group.aks_resource_group.location
  resource_group_name = data.azurerm_resource_group.aks_resource_group.name
  dns_prefix          = "aks-dns"

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

// Create an ACR registry to be used by the AKS cluster
resource "azurerm_container_registry" "acr" {
  name                = "acr${local.suffix_string}"
  resource_group_name = data.azurerm_resource_group.aks_resource_group.name
  location            = data.azurerm_resource_group.aks_resource_group.location
  sku                 = "Premium"
  admin_enabled       = false
  retention_policy {
    enabled = true
    days    = 7
  }
}

// Allow the AKS cluster to pull images from ACR
resource "azurerm_role_assignment" "aks_to_acr" {
  scope                = azurerm_container_registry.acr.id
  role_definition_name = "AcrPull"
  principal_id         = azurerm_kubernetes_cluster.aks.kubelet_identity.0.object_id
}

data "azurerm_private_dns_zone" "aks_private_dns_zone" {
  name                = join(".", slice(split(".", azurerm_kubernetes_cluster.aks.private_fqdn), 1, length(split(".", azurerm_kubernetes_cluster.aks.private_fqdn))))
  resource_group_name = azurerm_kubernetes_cluster.aks.node_resource_group
}


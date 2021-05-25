output "resource_group" {
  value = data.azurerm_resource_group.rg.name
}

output "location" {
  value = data.azurerm_resource_group.rg.location
}

output "storage_account_name" {
  value = azurerm_storage_account.storage.name
}

output "user_assigned_identity_name" {
  value = azurerm_user_assigned_identity.identity.name
}

output "ssh_key_name" {
  value = azurerm_ssh_public_key.key.name
}

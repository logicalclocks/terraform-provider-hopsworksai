variable "resource_group" {
  description = "Azure resource group"
  type        = string
}

variable "storage_account_name" {
  description = "Storage account name"
  type        = string
}

variable "user_assigned_identity_name" {
  description = "User assigned identity name"
  type        = string
}

variable "ssh_key_name" {
  description = "SSH Key pair name"
  type        = string
}

variable "ssh_public_key" {
  description = "Public key used with this ssh key pair"
  type        = string
}

variable "virtual_network_name" {
  description = "Virtual network name"
  type        = string
}
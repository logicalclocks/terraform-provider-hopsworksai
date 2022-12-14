provider "aws" {
  region  = var.aws_region
  profile = var.aws_profile
}

provider "azurerm" {
  features {}
  skip_provider_registration = true
}

resource "random_string" "suffix" {
  length  = 8
  special = false
  lower   = true
  upper   = false
}

module "aws" {
  count              = var.skip_aws ? 0 : 1
  source             = "./aws"
  region             = var.aws_region
  bucket_name_prefix = "tf-bucket-${random_string.suffix.result}"
  # This is the number of buckets to be created for testing
  # Important that this number have be incremented for each new AWS test case that require creating a cluster
  num_buckets           = 13
  instance_profile_name = "tf-instance-profile-${random_string.suffix.result}"
  ssh_key_name          = "tf-key-${random_string.suffix.result}"
  ssh_public_key        = file("${path.module}/.keys/tf.pub")
}

module "azure" {
  count                       = var.skip_azure || var.azure_resource_group == null ? 0 : 1
  source                      = "./azure"
  resource_group              = var.azure_resource_group
  storage_account_name        = "tfstorage${random_string.suffix.result}"
  user_assigned_identity_name = "tf-identity-${random_string.suffix.result}"
  ssh_key_name                = "tf-key-${random_string.suffix.result}"
  ssh_public_key              = file("${path.module}/.keys/tf.pub")
  virtual_network_name        = "tf-vnet-${random_string.suffix.result}"
  acr_registry_name           = "tfacr${random_string.suffix.result}"
}

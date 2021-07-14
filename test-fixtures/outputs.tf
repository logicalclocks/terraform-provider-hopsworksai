output "aws_instance_profile_arn" {
  value = var.skip_aws ? null : module.aws[0].instance_profile_arn
}

output "aws_bucket_names" {
  value = var.skip_aws ? null : module.aws[0].bucket_names
}

output "aws_ssh_key_name" {
  value = var.skip_aws ? null : module.aws[0].ssh_key_name
}

output "aws_region" {
  value = var.skip_aws ? null : module.aws[0].region
}

output "aws_vpc_id" {
  value = var.skip_aws ? null : module.aws[0].vpc_id
}

output "aws_subnet_id" {
  value = var.skip_aws ? null : module.aws[0].subnet_id
}

output "aws_security_group_id" {
  value = var.skip_aws ? null : module.aws[0].security_group_id
}

output "azure_resource_group" {
  value = var.skip_azure ? null : module.azure[0].resource_group
}

output "azure_location" {
  value = var.skip_azure ? null : module.azure[0].location
}

output "azure_storage_account_name" {
  value = var.skip_azure ? null : module.azure[0].storage_account_name
}

output "azure_user_assigned_identity_name" {
  value = var.skip_azure ? null : module.azure[0].user_assigned_identity_name
}

output "azure_ssh_key_name" {
  value = var.skip_azure ? null : module.azure[0].ssh_key_name
}

output "azure_virtual_network_name" {
  value = var.skip_azure ? null : module.azure[0].virtual_network_name
}

output "azure_subnet_name" {
  value = var.skip_azure ? null : module.azure[0].subnet_name
}

output "azure_security_group_name" {
  value = var.skip_azure ? null : module.azure[0].security_group_name
}
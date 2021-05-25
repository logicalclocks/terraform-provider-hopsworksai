output "aws_instance_profile_arn" {
  value = var.skip_aws ? null : module.aws[0].instance_profile_arn
}

output "aws_bucket_name" {
  value = var.skip_aws ? null : module.aws[0].bucket_name
}

output "aws_ssh_key_name" {
  value = var.skip_aws ? null : module.aws[0].ssh_key_name
}

output "aws_region" {
  value = var.skip_aws ? null : module.aws[0].region
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

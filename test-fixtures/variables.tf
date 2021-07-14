variable "azure_resource_group" {
  description = "Azure resource group where we create the needed test infra"
  type        = string
  default     = null
}

variable "aws_region" {
  description = "Default AWS region where we create the needed test infra"
  type        = string
  default     = "us-east-2"
}

variable "aws_profile" {
  description = "AWS profile to use."
  type        = string
  default     = "default"
}

variable "skip_aws" {
  description = "Skip creating resources for AWS tests"
  type        = bool
  default     = false
}

variable "skip_azure" {
  description = "Skip creating resources for Azure tests"
  type        = bool
  default     = false
}
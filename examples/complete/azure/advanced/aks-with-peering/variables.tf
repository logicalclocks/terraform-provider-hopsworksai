variable "aks_resource_group" {
  description = "The resource group where we will create an AKS cluster, an ACR registry"
  type        = string
}

variable "hopsworks_resource_group" {
  description = "The resource group where we will create a Hopsworks cluster."
  type        = string
}
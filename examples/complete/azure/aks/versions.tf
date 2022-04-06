terraform {
  required_version = ">= 0.14.0"

  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "2.63.0"
    }
    hopsworksai = {
      source = "logicalclocks/hopsworksai"
    }
  }
}
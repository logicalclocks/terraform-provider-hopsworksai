terraform {
  required_version = ">= 0.14.0"

  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "3.46.0"
    }
    hopsworksai = {
      source = "logicalclocks/hopsworksai"
    }
    random = {
      source  = "hashicorp/random"
      version = "3.2.0"
    }
  }
}
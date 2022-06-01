terraform {
  required_version = ">= 0.14.0"

  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "3.8.0"
    }
    hopsworksai = {
      source = "logicalclocks/hopsworksai"
    }
  }
}

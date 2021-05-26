terraform {
  required_version = "~> 0.14"

  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "2.60.0"
    }

    hopsworksai = {
      source  = "logicalclocks/hopsworksai"
      version = "0.1.0"
    }
  }
}
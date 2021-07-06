terraform {
  required_version = ">= 0.14"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "3.42.0"
    }

    azurerm = {
      source  = "hashicorp/azurerm"
      version = "2.60.0"
    }

    hopsworksai = {
      source = "logicalclocks/hopsworksai"
    }

    random = {
      source  = "hashicorp/random"
      version = "3.1.0"
    }
  }
}
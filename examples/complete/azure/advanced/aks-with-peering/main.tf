provider "hopsworksai" {
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

locals {
  suffix_string = random_string.suffix.result
}
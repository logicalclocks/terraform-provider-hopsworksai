terraform {
  required_version = ">= 0.14.0"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "5.13.0"
    }
    hopsworksai = {
      source = "logicalclocks/hopsworksai"
    }
  }
}

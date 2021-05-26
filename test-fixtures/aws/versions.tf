terraform {
  required_version = "~> 0.14"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "3.42.0"
    }

    hopsworksai = {
      source  = "logicalclocks/hopsworksai"
      version = "0.1.0"
    }
  }
}
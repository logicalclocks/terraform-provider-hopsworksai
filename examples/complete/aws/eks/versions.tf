terraform {
  required_version = ">= 0.14.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "4.16.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "2.3.0"
    }
    hopsworksai = {
      source = "logicalclocks/hopsworksai"
    }
  }
}

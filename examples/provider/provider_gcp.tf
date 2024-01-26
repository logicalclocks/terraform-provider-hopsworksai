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


variable "region" {
  type    = string
  default = "us-east1"
}

variable "project" {
  type = string
}

provider "google" {
  region  = var.region
  project = var.project
}

provider "hopsworksai" {
  # Highly recommended to use the HOPSWORKSAI_API_KEY environment variable instead
  api_key = "YOUR HOPSWORKS API KEY"
}


# Step 1: Create required google resources, a storage bucket and an service account with the required hopsworks permissions
data "hopsworksai_gcp_service_account_custom_role_permissions" "service_account" {

}

resource "google_project_iam_custom_role" "service_account_role" {
  role_id     = "tf.HopsworksAIInstances"
  title       = "Hopsworks AI Instances"
  description = "Role that allows Hopsworks AI Instances to access resources"
  permissions = data.hopsworksai_gcp_service_account_custom_role_permissions.service_account.permissions
}

resource "google_service_account" "service_account" {
  account_id   = "tf-hopsworks-ai-instances"
  display_name = "Hopsworks AI instances"
  description  = "Service account for Hopsworks AI instances"
}

resource "google_project_iam_binding" "service_account_role_binding" {
  project = var.project
  role    = google_project_iam_custom_role.service_account_role.id

  members = [
    google_service_account.service_account.member
  ]
}

resource "google_storage_bucket" "bucket" {
  name          = "tf-hopsworks-bucket"
  location      = var.region
  force_destroy = true
}

# Step 2: create a cluster with 1 worker

data "google_compute_zones" "available" {
  region = var.region
}

locals {
  zone = data.google_compute_zones.available.names.0
}

data "hopsworksai_instance_type" "head" {
  cloud_provider = "GCP"
  node_type      = "head"
  region         = local.zone
}

data "hopsworksai_instance_type" "rondb_data" {
  cloud_provider = "GCP"
  node_type      = "rondb_data"
  region         = local.zone
}

data "hopsworksai_instance_type" "small_worker" {
  cloud_provider = "GCP"
  node_type      = "worker"
  region         = local.zone
  min_memory_gb  = 16
  min_cpus       = 4
}

resource "hopsworksai_cluster" "cluster" {
  name = "tf-cluster"

  head {
    instance_type = data.hopsworksai_instance_type.head.id
  }

  workers {
    instance_type = data.hopsworksai_instance_type.smallest_worker.id
    count         = 1
  }

  gcp_attributes {
    project_id            = var.project
    region                = var.region
    zone                  = local.zone
    service_account_email = google_service_account.service_account.email
    bucket {
      name = google_storage_bucket.bucket.name
    }
  }

  rondb {
    single_node {
      instance_type = data.hopsworksai_instance_type.rondb_data.id
    }
  }

  open_ports {
    ssh = true
  }
}

# Outputs the url of the newly created cluster 
output "hopsworks_cluster_url" {
  value = hopsworksai_cluster.cluster.url
}
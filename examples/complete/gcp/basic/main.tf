provider "google" {
  region  = var.region
  project = var.project
}

provider "hopsworksai" {
}

# Create required google resources, a storage bucket and an service account with the required hopsworks permissions
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

data "google_compute_zones" "available" {
  region = var.region
}

locals {
  zone = data.google_compute_zones.available.names.0
}

# Create a simple cluster with two workers with two different configuration
data "hopsworksai_instance_type" "head" {
  cloud_provider = "GCP"
  node_type      = "head"
  region         = local.zone
}

data "hopsworksai_instance_type" "rondb_mgm" {
  cloud_provider = "GCP"
  node_type      = "rondb_management"
  region         = local.zone
}

data "hopsworksai_instance_type" "rondb_data" {
  cloud_provider = "GCP"
  node_type      = "rondb_data"
  region         = local.zone
  min_memory_gb  = 32
}

data "hopsworksai_instance_type" "rondb_mysql" {
  cloud_provider = "GCP"
  node_type      = "rondb_mysql"
  region         = local.zone
}

data "hopsworksai_instance_type" "small_worker" {
  cloud_provider = "GCP"
  node_type      = "worker"
  region         = local.zone
  min_memory_gb  = 16
  min_cpus       = 4
}

data "hopsworksai_instance_type" "large_worker" {
  cloud_provider = "GCP"
  node_type      = "worker"
  region         = local.zone
  min_memory_gb  = 32
  min_cpus       = 4
}

resource "hopsworksai_cluster" "cluster" {
  name = "tf-cluster"

  head {
    instance_type = data.hopsworksai_instance_type.head.id
  }

  workers {
    instance_type = data.hopsworksai_instance_type.small_worker.id
    disk_size     = 256
    count         = 1
  }

  workers {
    instance_type = data.hopsworksai_instance_type.large_worker.id
    disk_size     = 512
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
    management_nodes {
      instance_type = data.hopsworksai_instance_type.rondb_mgm.id
    }
    data_nodes {
      instance_type = data.hopsworksai_instance_type.rondb_data.id
    }
    mysql_nodes {
      instance_type = data.hopsworksai_instance_type.rondb_mysql.id
    }
  }

  open_ports {
    ssh = true
  }
}

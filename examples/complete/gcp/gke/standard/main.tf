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

# Attach Kubernetes developer role to the service account for cluster instance

resource "google_project_iam_binding" "service_account_k8s_role_binding" {
  project = var.project
  role    = "roles/container.developer"

  members = [
    google_service_account.service_account.member
  ]
}

# Create a network 
data "google_compute_zones" "available" {
  region = var.region
}

locals {
  zone = data.google_compute_zones.available.names.0
}

resource "google_compute_network" "network" {
  name                    = "tf-hopsworks"
  auto_create_subnetworks = false
  mtu                     = 1460
}

resource "google_compute_subnetwork" "subnetwork" {
  name          = "tf-hopsworks-subnetwork"
  ip_cidr_range = "10.1.0.0/24"
  region        = var.region
  network       = google_compute_network.network.id
}

resource "google_compute_firewall" "nodetonode" {
  name    = "tf-hopsworks-nodetonode"
  network = google_compute_network.network.name
  allow {
    protocol = "all"
  }
  direction               = "INGRESS"
  source_service_accounts = [google_service_account.service_account.email]
  target_service_accounts = [google_service_account.service_account.email]
}

resource "google_compute_firewall" "inbound" {
  name    = "tf-hopsworks-inbound"
  network = google_compute_network.network.name
  allow {
    protocol = "tcp"
    ports    = ["80", "443"]
  }

  direction               = "INGRESS"
  target_service_accounts = [google_service_account.service_account.email]
  source_ranges           = ["0.0.0.0/0"]
}

# Create a standard GKE cluster 
resource "google_container_cluster" "cluster" {
  name       = "tf-gke-cluster"
  location   = local.zone
  network    = google_compute_network.network.name
  subnetwork = google_compute_subnetwork.subnetwork.name

  ip_allocation_policy {
    cluster_ipv4_cidr_block = "10.124.0.0/14"
  }

  deletion_protection = false
  # We can't create a cluster with no node pool defined, but we want to only use
  # separately managed node pools. So we create the smallest possible default
  # node pool and immediately delete it.
  remove_default_node_pool = true
  initial_node_count       = 1
}

resource "google_container_node_pool" "node_pool" {
  name       = "tf-hopsworks-node-pool"
  location   = local.zone
  cluster    = google_container_cluster.cluster.name
  node_count = 1
  node_config {
    machine_type = "e2-standard-8"
  }
}

resource "google_compute_firewall" "gke_traffic" {
  name    = "tf-hopsworks-gke-traffic"
  network = google_compute_network.network.name
  allow {
    protocol = "all"
  }

  direction               = "INGRESS"
  target_service_accounts = [google_service_account.service_account.email]
  source_ranges           = [google_container_cluster.cluster.ip_allocation_policy.0.cluster_ipv4_cidr_block]
}

# Create a simple cluster with autoscale and GKE integration
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
}

data "hopsworksai_instance_type" "rondb_mysql" {
  cloud_provider = "GCP"
  node_type      = "rondb_mysql"
  region         = local.zone
}

data "hopsworksai_instance_type" "worker" {
  cloud_provider = "GCP"
  node_type      = "worker"
  region         = local.zone
  min_memory_gb  = 16
  min_cpus       = 8
}

resource "hopsworksai_cluster" "cluster" {
  name = "tf-cluster"

  head {
    instance_type = data.hopsworksai_instance_type.head.id
  }

  autoscale {
    non_gpu_workers {
      instance_type       = data.hopsworksai_instance_type.worker.id
      disk_size           = 256
      min_workers         = 1
      max_workers         = 5
      standby_workers     = 0.5
      downscale_wait_time = 300
    }
  }

  gcp_attributes {
    project_id            = var.project
    region                = var.region
    zone                  = local.zone
    service_account_email = google_service_account.service_account.email
    bucket {
      name = google_storage_bucket.bucket.name
    }
    network {
      network_name    = google_compute_network.network.name
      subnetwork_name = google_compute_subnetwork.subnetwork.name
    }
    gke_cluster_name = google_container_cluster.cluster.name
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
}

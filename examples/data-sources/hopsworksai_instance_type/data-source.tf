# retrieve the smallest supported instance type for head node
data "hopsworksai_instance_type" "supported_type" {
  node_type      = "head"
  cloud_provider = "AWS"
  region         = "us-east-2"
}

# retrieve the smallest supported instance type for head node with at least 32 GB memory and 16 vCPUs
data "hopsworksai_instance_type" "supported_type" {
  node_type      = "head"
  cloud_provider = "AWS"
  region         = "us-east-2"
  min_memory_gb  = 32
  min_cpus       = 16
}
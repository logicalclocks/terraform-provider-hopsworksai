# retrieve the smallest supported instance type for head node
data "hopsworksai_instance_type" "supported_type" {
  node_type      = "head"
  cloud_provider = "AWS"
}

# retrieve the smallest supported instance type for head node with at least 32 GB memory and 16 vCPUs
data "hopsworksai_instance_type" "supported_type" {
  node_type      = "head"
  cloud_provider = "AWS"
  min_memory_gb  = 32
  min_cpus       = 16
}
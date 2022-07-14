# retrieve all supported instance types for workers in Hopsworks.ai 
data "hopsworksai_instance_types" "supported_worker_types" {
  node_type      = "head"
  cloud_provider = "AWS"
  region         = "us-east-2"
}
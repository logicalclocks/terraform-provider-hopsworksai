
# default policy
data "hopsworksai_aws_instance_profile_policy" "policy" {

}

# limit permissions to a single S3 bucket
data "hopsworksai_aws_instance_profile_policy" "policy" {
  bucket_name = "my-bucket"
}

# remove eks and ecr permissions
data "hopsworksai_aws_instance_profile_policy" "policy" {
  enable_eks = false
  enable_ecr = false
}
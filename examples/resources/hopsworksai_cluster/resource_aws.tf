resource "hopsworksai_cluster" "cluster" {
  name    = "my-cluster-name"
  ssh_key = "my-ssh-key"

  head {
  }

  aws_attributes {
    region               = "us-east-2"
    instance_profile_arn = "arn:aws:iam::0000000000:instance-profile/my-instance-profile"
    bucket_name          = "my-bucket"
  }

  open_ports {
    ssh = true
  }

  tags = {
    "Purpose" = "testing"
  }
}
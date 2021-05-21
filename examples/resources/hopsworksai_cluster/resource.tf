resource "hopsworksai_cluster" "myCluster" {
  name    = "my-cluster-name"
  ssh_key = "my-ssh-key" # your AWS SSH key
  version = "2.2.0"

  head {
  }

  workers {
    count = 1
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
resource "hopsworksai_cluster" "cluster" {
  name    = "my-cluster-name"
  ssh_key = "my-ssh-key"

  head {
    instance_type = ""
  }

  aws_attributes {
    region               = "us-east-2"
    instance_profile_arn = "arn:aws:iam::0000000000:instance-profile/my-instance-profile"
    bucket {
      name = "my-bucket"
    }
  }

  rondb {
    management_nodes {
      instance_type = ""
    }
    data_nodes {
      instance_type = ""
    }
    mysql_nodes {
      instance_type = ""
    }
  }

  open_ports {
    ssh = true
  }

  tags = {
    "Purpose" = "testing"
  }
}
resource "hopsworksai_cluster" "cluster" {
  name = "my-cluster-name"

  head {
    instance_type = ""
  }


  gcp_attributes {
    project_id            = "my-project"
    region                = "us-east1"
    zone                  = "us-east1-b"
    service_account_email = "hopsworks-ai-instances@my-project.iam.gserviceaccount.com"
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
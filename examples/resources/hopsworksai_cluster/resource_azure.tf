resource "hopsworksai_cluster" "cluster" {
  name    = "myclustername"
  ssh_key = "my-ssh-key"

  head {
    instance_type = ""
  }

  azure_attributes {
    location                       = "northeurope"
    resource_group                 = "mygroup"
    user_assigned_managed_identity = "my-identity"
    container {
      storage_account = "mystorage"
    }
    acr_registry_name = "registry-name"
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
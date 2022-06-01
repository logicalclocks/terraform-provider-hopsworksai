resource "hopsworksai_cluster" "cluster" {
  name    = "myclustername"
  ssh_key = "my-ssh-key"

  head {
  }

  azure_attributes {
    location                       = "northeurope"
    resource_group                 = "mygroup"
    user_assigned_managed_identity = "my-identity"
    container {
      storage_account = "mystorage"
    }
  }

  rondb {

  }

  open_ports {
    ssh = true
  }

  tags = {
    "Purpose" = "testing"
  }
}
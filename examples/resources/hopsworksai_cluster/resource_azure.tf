resource "hopsworksai_cluster" "cluster" {
  name    = "myclustername"
  ssh_key = "my-ssh-key"

  head {
  }

  workers {
    count = 1
  }

  azure_attributes {
    location                       = "northeurope"
    resource_group                 = "mygroup"
    storage_account                = "mystorage"
    user_assigned_managed_identity = "my-identity"
  }

  open_ports {
    ssh = true
  }

  tags = {
    "Purpose" = "testing"
  }
}
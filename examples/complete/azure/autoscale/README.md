# Hopsworks cluster with autoscale 

In this example, we create a Hopsworks cluster with autoscale support. We configure autoscale to use an instance type with at least 4 vCPUs and 16 GB of memory, a min number of workers set to 0, a max number of workers set to 10, at least 50% of workers should be running as standby, and only remove workers after 300 seconds of inactivity.

## Configure RonDB

You can configure RonDB nodes instead of relying on the default configurations, for instance in the following example, we increased the number of data nodes to 4 and we used an instance type with at least 8 CPUs and 16 GB of memory.

```hcl
data "hopsworksai_instance_type" "smallest_rondb_datanode" {
  cloud_provider = "AZURE"
  node_type      = "rondb_data"
  min_memory_gb  = 16
  min_cpus       = 8
}

resource "hopsworksai_cluster" "cluster" {
  # all the other configurations are omitted for clarity 

  rondb {
    data_nodes {
      instance_type = data.hopsworksai_instance_type.smallest_rondb_datanode.id
      disk_size     = 512
      count         = 4
    }
  }
}
```

## How to run the example 
First ensure that your azure credentials are setup correctly by running the following command

```bash
az login 
```

Then, run the following commands. Replace the placeholders with your Hopsworks API Key and your Azure resource group

```bash
export HOPSWORKSAI_API_KEY=<YOUR_HOPSWORKSAI_API_KEY>
terraform init
terraform apply -var="resource_group=<YOUR_RESOURCE_GROUP>"
```

## Update Autoscale 

You can update the autoscale configuration after creations, by changing the `autoscale` configuration block. For example, you can configure your own worker as follows:

> **Notice** that you need to run `terraform apply` after updating your configuration for your changes to take place.

```hcl
data "hopsworksai_instance_type" "my_worker" {
  cloud_provider = "AZURE"
  node_type      = "worker"
  min_cpus       = 16
}

resource "hopsworksai_cluster" "cluster" {
  # all the other configurations are omitted for clarity 

  autoscale {
    non_gpu_workers {
      instance_type = data.hopsworksai_instance_type.my_worker.id
      disk_size = 256
      min_workers = 0
      max_workers = 10
      standby_workers = 0.5
      downscale_wait_time = 300
    }
  }

}
```

You can also remove the autoscale configuration by removing the `autoscale` configuration block. Notice that after removing `autoscale` block, workers created during autoscale will not be destroyed automatically, instead, on the next `terraform apply` you will be prompted to destroy these workers as shown below. 

```diff
# hopsworksai_cluster.cluster will be updated in-place
  ~ resource "hopsworksai_cluster" "cluster" {
        id                             = "3b653b00-d016-11eb-9b2f-1345035d566e"
        name                           = "tf-hopsworks-cluster"
        tags                           = {}
        # (13 unchanged attributes hidden)




      - workers {
          - count         = 2 -> null
          - disk_size     = 256 -> null
          - instance_type = "m5.xlarge" -> null
        }
        # (3 unchanged blocks hidden)
    }

Plan: 0 to add, 1 to change, 0 to destroy.
```

If you do wish to keep these workers as part of your configuration, you should add their corresponding `workers` blocks to the terraform main file.


## Terminate the cluster

You can run `terraform destroy` to delete the cluster and all the other required cloud resources created in this example.
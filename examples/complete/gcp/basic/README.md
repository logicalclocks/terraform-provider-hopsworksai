# Hopsworks cluster with 2 workers

In this example, we create a Hopsworks cluster with 2 workers. We configure one of the workers to use an instance type with at least 4 vCPUs and 16 GB of memory while using at least 32 GB of memory for the other worker.

## Configure RonDB

You can configure RonDB nodes instead of relying on the default configurations, for instance in the following example, we increased the number of data nodes to 4 and we used an instance type with at least 8 CPUs and 16 GB of memory.

```hcl
data "hopsworksai_instance_type" "smallest_rondb_datanode" {
  cloud_provider = "GCP"
  node_type      = "rondb_data"
  region         = local.zone
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
First ensure that your aws credentials are setup correctly by running the following command 

```bash
aws configure 
```

Then, run the following commands. Replace the placeholder with your Hopsworks API Key. The cluster will be created in us-east-2 region by default, however, you can configure which region to use by setting the variable region when applying the changes `-var="region=YOUR_REGION"`

```bash
export HOPSWORKSAI_API_KEY=<YOUR_HOPSWORKSAI_API_KEY>
terraform init
terraform apply
```

## Update workers 

You can always update the worker configurations after creation, for example you can increase the number of small workers to use 2 instead of 1 as follows:

> **Notice** that you need to run `terraform apply` after updating your configuration for your changes to take place.

```hcl
resource "hopsworksai_cluster" "cluster" {
  # all the other configurations are omitted for clarity 

  workers {
    instance_type = data.hopsworksai_instance_type.small_worker.id
    disk_size = 256
    count = 2
  }

  workers {
    instance_type = data.hopsworksai_instance_type.large_worker.id
    disk_size = 512
    count = 1
  }

}
```

Also, you can remove the large worker if you want by removing the large workers block as follows:

```hcl
resource "hopsworksai_cluster" "cluster" {
  # all the other configurations are omitted for clarity 

  workers {
    instance_type = data.hopsworksai_instance_type.small_worker.id
    disk_size = 256
    count = 2
  }

}
```

You can add a new different worker type for example another worker with at least 16 cpu cores as follows:

```hcl
data "hopsworksai_instance_type" "my_worker" {
  cloud_provider = "GCP"
  node_type      = "worker"
  region         = local.zone
  min_cpus       = 16
}

resource "hopsworksai_cluster" "cluster" {
  # all the other configurations are omitted for clarity 

  workers {
    instance_type = data.hopsworksai_instance_type.small_worker.id
    disk_size = 256
    count = 2
  }

  workers {
    instance_type = data.hopsworksai_instance_type.my_worker.id
    disk_size = 512
    count = 1
  }

}
```

## Terminate the cluster

You can run `terraform destroy` to delete the cluster and all the other required cloud resources created in this example.
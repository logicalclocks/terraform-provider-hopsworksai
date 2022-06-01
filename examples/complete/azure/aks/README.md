# Integrate Hopsworks cluster with Azure AKS and ACR

In this example, we create an AKS cluster, an ACR registry, and a Hopsworks cluster that is integrated with both AKS and ACR. We also create a virtual network where both AKS and Hopsworks reside ensuring that they can communicate with each other given that the AKS cluster is private with no public access allowed.

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
terraform apply  -var="resource_group=<YOUR_RESOURCE_GROUP>"
```

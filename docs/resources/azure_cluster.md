---
page_title: "hopsworksai_azure_cluster Resource - terraform-provider-hopsworksai"
subcategory: ""
description: |-
  Sample resource in the Terraform provider scaffolding.
---

# Resource `hopsworksai_azure_cluster`

Sample resource in the Terraform provider scaffolding.



## Schema

### Required

- **cluster_name** (String) Name of the hopsworks cluster, must be unique
- **location** (String) Azure location the Hopsworks cluster will reside in
- **managed_identity** (String) Azure managed identity the Hopsworks instances will be started with
- **resource_group** (String) Resource group the Hopsworks cluster will reside in
- **ssh_key** (String) SSH key resource for the instances
- **storage_account** (String) Azure storage account the Hopsworks cluster will use to store data in

### Optional

- **head_node_instance_type** (String) Instance type of the head node, default Standard_D8_v3
- **head_node_local_storage** (Number) Disk size of the head node in units of GB, default 512 GB
- **id** (String) The ID of this resource.
- **storage_container_name** (String) Azure storage container the Hopsworks cluster will use to store data in, automatically generated if not set.
- **version** (String) Hopsworks version, default 2.1.0
- **worker_node_count** (Number) Number of worker nodes
- **worker_node_instance_type** (String) Instance type of worker nodes, default Standard_D8_v3
- **worker_node_local_storage** (Number) Disk size of worker node(s) in units of GB, default 512 GB

### Read-only

- **cluster_id** (String) Unique identifier of the hopsworks cluster, automatically generated



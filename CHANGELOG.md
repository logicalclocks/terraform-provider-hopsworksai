## 0.3.0 (June 24, 2021)

BUG FIXES:
* resource/hopsworksai_cluster: Set Required to true for `azure_attributes/network/subnet_name` to ensure setting the subnet_name if setting up your own network configuration

ENHANCEMENTS:
* resource/hopsworksai_cluster: Add a new attribute `azure_attributes/network/resource_group` to allow setting up the network on a different resource group


## 0.2.1 (June 18, 2021)

ENHANCEMENTS:
* Update documentations 
* Add more examples

## 0.2.0 (June 16, 2021)

BUG FIXES:
* resource/hopsworksai_cluster: check if `ecr_registry_account_id`, `eks_cluster_name`, and `aks_cluster_name` is not an empty string before setting it
* resource/hopsworksai_cluster: Skip name validation and relay on backend for validation

ENHANCEMENTS:
* more unit tests
* resource/hopsworksai_cluster: Add a new attribute `rondb`
* resource/hopsworksai_cluster: Add a new attribute `autoscale`
* resource/hopsworksai_cluster: Set Computed to true for `head/instance_type` and `azure_attributes/storage_container_name` 
* resource/hopsworksai_cluster: Set Required to true for `workers/instance_type`
* resource/hopsworksai_cluster: Set Required to true for `azure_attributes/network/virtual_network_name` and `aws_attributes/network/vpc_id`
* resource/hopsworksai_cluster: Set Computed to true for `azure_attributes/network/subnet_name`,  `azure_attributes/network/security_group_name`, `aws_attributes/network/subnet_id` and `aws_attributes/network/security_group_id`
* datasource/azure_user_assigned_identity_permissions: Add a new attribute `enable_aks_and_acr`
* resource/hopsworksai_cluster: Set Computed to true for `aws_attributes/ecr_registry_account_id` 
* complete example to create Hopsworks clusters with AKS/ACR and EKS/ECR

FEATURES:
* **New Data Source**: `hopsworksai_instance_type`
* **New Data Source**: `hopsworksai_instance_types`


## 0.1.0 (June 2, 2021)

FEATURES:
* **New Resource**: `hopsworksai_cluster`
* **New Data Source**: `hopsworksai_clusters`
* **New Data Source**: `hopsworksai_aws_instance_profile_policy`
* **New Data Source**: `hopsworksai_azure_user_assigned_identity_permissions`
* **New Data Source**: `hopsworksai_cluster`

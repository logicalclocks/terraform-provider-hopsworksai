## 0.9.0 (Unreleased)

NOTES:

BREAKING CHANGES:

BUG FIXES:

ENHANCEMENTS:

FEATURES:


## 0.8.0 (February 25, 2022)

NOTES:

* datasource/aws_instance_profile_policy: Deprecate `enable_upgrade` attribute since these permissions are not required anymore from version 2.4.0 onwards
* datasource/azure_user_assigned_identity_permissions: Deprecate `enable_upgrade` attribute since these permissions are not required anymore from version 2.4.0 onwards

BUG FIXES:
* resource/hopsworksai_cluster: fix incorrect conversion error
* datasource/aws_instance_profile_policy: Add missing backup permissions
* resource/hopsworksai_cluster: fix upgrade and rollback to work with version 2.4.0 and onwards

ENHANCEMENTS:
* datasource/aws_instance_profile_policy: Set Default to false for `enable_upgrade` attribute 
* resource/hopsworksai_cluster: Set Default `version` to 2.5.0
* Extend resource testing framework to allow multiplexing HTTP requests with the same method and path
* resource/hopsworksai_cluster: Add support for changing `instance_type` of head node and RonDB nodes

## 0.7.0 (December 14, 2021)

ENHANCEMENTS:
* examples/complete/aws: add aws profile in variables
* datasource/dataSourceAWSInstanceProfilePolicy: Add possibility to limit permissions to eks cluster name.
* datasource/dataSourceAWSInstanceProfilePolicy: Add possibility to limit permissions to cluster id.

## 0.6.0 (October 27, 2021)

BREAKING CHANGES:

BUG FIXES:
* resource/hopsworksai_cluster: The `version` updates should always run with no other updates.
* Fix TestExpandTags to avoid inconsistent results due to different ordering when iterating the tags map.

ENHANCEMENTS:
* resource/hopsworksai_cluster: Add support for updating `version` attribute to allow upgrade and rollback
* resource/hopsworksai_cluster: Add a new computed attribute `upgrade_in_progress`
* resource/hopsworksai_cluster: Rename attribute `azure_attributes/search_domain` to `azure_attributes/network/search_domain` and deprecate `azure_attributes/search_domain`
* resource/hopsworksai_cluster: Update default `version` to 2.4.0

FEATURES:
* **New Data Source**: `hopsworksai_version`


## 0.5.0 (July 30, 2021)

ENHANCEMENTS:
* resource/hopsworksai_cluster: Add a new attribute `search_domain`


## 0.4.0 (July 14, 2021)

BUG FIXES:
* resource/hopsworksai_cluster: Fix validation condition for `backup_retention_period`
* datasource/hopsworksai_cluster: Check if cluster is not nil before updating state
* resource/hopsworksai_cluster: Set Required to true for `aws_attributes/network/subnet_id` to ensure setting the subnet_name if setting up your own network configuration
* Add command-failed to cluster error states

ENHANCEMENTS:
* resource/hopsworksai_cluster: Add a new attribute `init_script`
* resource/hopsworksai_cluster: Add new attributes `workers/spot_config`, `autoscale/non_gpu_workers/spot_config`, and `autoscale/gpu_workers/spot_config`
* resource/hopsworksai_cluster: Add a new attribute `os`
* resource/hopsworksai_cluster: Add a new attribute `run_init_script_first`
* Allow setting aws_profile and aws_region when running acceptance tests

FEATURES:
* **New Resource**: `hopsworksai_backup`
* **New Resource**: `hopsworksai_cluster_from_backup`
* **New Data Source**: `hopsworksai_backup`
* **New Data Source**: `hopsworksai_backups`


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

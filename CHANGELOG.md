## 1.3.1 (January 26, 2023)

ENHANCEMENTS:
* resource/hopsworksai_cluster: Enforce setting `ecr_registry_account_id` and `acr_registry_name` during upgrade from 3.0 to 3.1

## 1.3.0 (January 4, 2023)

ENHANCEMENTS:
* datasource/aws_instance_profile_policy: Deprecate `enable_eks_and_ecr` attribute to use `enable_eks` and `enable_ecr` instead
* datasource/azure_user_assigned_identity_permissions: Deprecate `enable_aks_and_acr` attribute  to use `enable_aks` and `enable_acr` instead
* Update acceptance tests and examples for new release

## 1.2.2 (December 6, 2022)

ENHANCEMENTS:
* add a jenkins pipline for acceptance tests
* dependencies: Bump hashicorp/terraform-plugin-sdk/v2 from 2.24.0 to 2.24.1
* collect ec2init log in aws acceptance tests
* datasource/aws_instance_profile_policy: Add permissions for internal service images (onlinefs, airflow, git)

BUG FIXES:
* Fix TestAccClusterAWS_RonDB and TestAccClusterAzure_RonDB

## 1.2.1 (November 10, 2022)

ENHANCEMENTS:
* datasource/aws_instance_profile_policy: Add ecr:TagResource permission

## 1.2.0 (November 8, 2022)

ENHANCEMENTS:
* Bump minimum Go version to 1.18
* Bump golangci-lint version to 1.50.1
* dependencies: Bump hashicorp/terraform-plugin-sdk/v2 from 2.20.0 to 2.24.0
* resource/hopsworksai_cluster: Use managed docker containers by default from version 3.1.0

FEATURES:
* Add support for head node dedicated instance profile.
* Filter instance types based on nvme drives

## 1.1.0 (October 14, 2022)

ENHANCEMENTS:
* return nodes private ips

## 1.0.1 (August 1, 2022)

BUG FIXES:
docs: Update documentation links 

## 1.0.0 (August 1, 2022)

NOTES:
* The `instance_type` attribue(s) are not optional anymore
* resource/hopsworksai_cluster: Change default values for RonDB cluster. New default number of replicas is `2` and new default number of Datanodes is `2`

BREAKING CHANGES:
* Creating a Hopsworks cluster will require a seperate RonDB node. RonDB attribute is required by default.
* Remove deprecated `aws_attributes/bucket_name` attribute.
* Set `aws_attributes/bucket/name` attribute to be required.
* Remove deprecated `azure_attributes/storage_account` and `azure_attributes/storage_container_name` attributes.
* Set `azure_attributes/container/storage_account` attribute to be required.
* Remove deprecated `azure_attributes/search_domain` attribute.

ENHANCEMENTS:
* dependencies: Bump hashicorp/terraform-plugin-docs from 0.10.1 to 0.13.0
* dependencies: Bump hashicorp/terraform-plugin-sdk/v2 from 2.17.0 to 2.20.0
* dependencies: Bump hashicorp/terraform-plugin-log from 0.4.1 to 0.7.0
* devtools: Bump golangci-lint from 1.45.2 to 1.46.2
* datasource/instance_type(s): Add required `region` attribute to filter supported instance types
* resource/hopsworksai_cluster: Update `update_state` attribute description
* resource/hopsworksai_cluster: Do not use default `instance_type` values and make the attribute(s) required
* resource/hopsworksai_cluster: Set Default `version` to 3.0.0
* resource/hopsworksai_cluster: Add attribute `rondb/single_node` to use a RonDB single node cluster.

BUG FIXES:
* resource/hopsworksai_backup: Update acceptance tests to not stop cluster before taking backups
* resource/hopsworksai_cluster_from_backup: Update acceptance tests to not stop cluster before taking backups
* datasource/hopsworksai_azure_user_assigned_identity_permissions: Add the missing listkeys action permission for backup


## 0.11.0 (June 16, 2022)

NOTES:
* resource/hopsworksai_cluster: Change default values for RonDB cluster. New default number of replicas is `1` and new default number of Datanodes is `1`

BREAKING CHANGES:
* Default values for RonDB cluster changed. New default number of replicas is `1` and new default number of Datanodes is `1`

BUG FIXES:

ENHANCEMENTS:
* Bump minimum Go version to 1.17
* Bump golangci-lint version to 1.45.2
* dependencies: Bump hashicorp/terraform-plugin-sdk/v2 from 2.13.0 to 2.17.0
* dependencies: Bump hashicorp/terraform-plugin-docs from 0.7.0 to 0.10.1
* dependencies: Bump hashicorp/terraform-plugin-log from 0.3.0 to 0.4.1
* examples: Update versions and remove deprecated attributes
* resource/hopsworksai_cluster: Add `ha_enabled` experimental attribute to allow using multi head node setup for high availability.
* resource/hopsworksai_cluster: Add `cluster_domain_prefix` attribute to override the default UUID name of a Cluster. This option is available **only** to users with special privileges.
* resource/hopsworksai_cluster: Add `custom_hosted_zone` attribute to override the default Hosted Zone of a cluster's public domain name (cloud.hopsworks.ai). This option is available **only** to users with special privileges.
* resource/hopsworksai_cluster: Add `aws_attributes/ebs_encryption` attribute to configure EBS encryption for disks on AWS clusters.
* resource/hopsworksai_cluster: Change `ssh_key` attribute to be optional for AWS.

BUG FIXES:
* datasource/aws_instance_profile_policy: The `ecr:CreateRepository` permission has no resource level condition for private registries

## 0.10.1 (April 19, 2022)

BUG FIXES:
* resource/hopsworksai_backup: Handle get backup if backup not found and backupPipeline is InProgress
* resource/hopsworksai_backup: Wait for cluster start during backup pipeline
* resource/hopsworksai_backup: Fix interface conversion error 
* resource/hopsworksai_backup: Return empty backup object when pending to avoid not found checks 

ENHANCEMENTS:
* provider: Add `api_gateway` optional parameter to set a development API gateway. If not specified it defaults to `https://api.hopsworks.ai`

## 0.10.0 (April 11, 2022)

NOTES:
* resource/hopsworksai_cluster: Deprecate `aws_attributes/bucket_name` attribute to use `aws_attributes/bucket/name` instead
* resource/hopsworksai_cluster: Deprecate `azure_attributes/storage_account` attribute to use `azure_attributes/container/storage_account` instead
* resource/hopsworksai_cluster: Deprecate `azure_attributes/storage_container_name` attribute to use `azure_attributes/container/name` instead

ENHANCEMENTS:
* resource/hopsworksai_cluster: Add `aws_attributes/bucket` block to contain all bucket related configurations
* resource/hopsworksai_cluster: Add `aws_attributes/bucket/encryption` and `aws_attributes/bucket/acl` attributes to configure the bucket encryption and ACL properties
* resource/hopsworksai_cluster: Add `azure_attributes/container` block to contain all container related configurations
* resource/hopsworksai_cluster: Add `azure_attributes/container/encryption` attributes to configure the container encryption

## 0.9.1 (April 6, 2022)

ENHANCEMENTS:
* acceptance_tests: Tag resources with their respective test name
* acceptance_tests: Providers is deprecated use ProviderFactories instead
* acceptance_tests: Use r5 and c5 instance types in RonDB upscale tests
* logging: use tflog instead of log.Printf
* dependencies: Bump hashicorp/terraform-plugin-sdk/v2 from 2.11.0 to 2.13.0  
* examples: pin aws/azure provider versions to avoid breaking changes

## 0.9.0 (March 16, 2022)

BUG FIXES:
* resource/cluster: Add suppression check on `disk_size` attribute to avoid forced replacement during rollback

ENHANCEMENTS:
* resource/hopsworksai_cluster: Add `deactivate_hopsworksai_log_collection` attribute to deactivate or activate Hopsworks.ai log collection.
* resource/hopsworksai_cluster: Add `collect_logs` attribute to enable pushing services' logs to AWS CloudWatch.
* resource/hopsworksai_cluster: Add `head/node_id` readonly attribute to retrieve the corresponding aws/azure instance id of the head node.
* dependencies: Bump hashicorp/terraform-plugin-sdk/v2 from 2.10.1 to 2.11.0  
* dependencies: Bump hashicorp/terraform-plugin-docs from 0.5.1 to 0.7.0

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

## 0.2.0 (Unreleased)

BUG FIXES:
* resource/hopsworksai_cluster: check if `ecr_registry_account_id`, `eks_cluster_name`, and `aks_cluster_name` is not an empty string before setting it
* resource/hopsworksai_cluster: Skip name validation and relay on backend for validation

ENHANCEMENTS:
* more unit tests
* resource/hopsworksai_cluster: Add a new attribute `rondb`
* resource/hopsworksai_cluster: Set computed to true for `head/instance_type`, `workers/instance_type`, and `azure_attributes/storage_container_name` 
* resource/hopsworksai_cluster: Add a new attribute `autoscale`

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

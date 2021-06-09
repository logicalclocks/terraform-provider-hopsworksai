## 0.2.0 (Unreleased)

NOTES:

BUG FIXES:
* resource/hopsworksai_cluster: check if `ecr_registry_account_id` is not an empty string before setting it

ENHANCEMENTS:
* resource/hopsworksai_cluster: Add a new attribute `rondb`
* resource/hopsworksai_cluster: Set computed to true for `head/instance_type`, `workers/instance_type`, and `azure_attributes/storage_container_name` 
* more unit tests

FEATURES:

## 0.1.0 (June 2, 2021)

FEATURES:
* **New Resource**: `hopsworksai_cluster`
* **New Data Source**: `hopsworksai_clusters`
* **New Data Source**: `hopsworksai_aws_instance_profile_policy`
* **New Data Source**: `hopsworksai_azure_user_assigned_identity_permissions`
* **New Data Source**: `hopsworksai_cluster`

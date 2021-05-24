# default permissions
data "hopsworksai_azure_user_assigned_identity_permissions" "permissions" {

}

# disable backup permissions
data "hopsworksai_azure_user_assigned_identity_permissions" "permissions" {
  enable_backup = false
}
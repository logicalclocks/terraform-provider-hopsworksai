
# get all the user backups
data "hopsworksai_backups" "backups" {

}

# get all the user backups for a specific cluster
data "hopsworksai_backups" "backups" {
  cluster_id = "<CLUSTER ID>"
}
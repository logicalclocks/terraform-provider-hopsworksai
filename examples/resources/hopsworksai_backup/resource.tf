
resource "hopsworksai_backup" "backup" {
  cluster_id  = "<CLUSTER_ID>"
  backup_name = "my-test-backup"
}
# retrieve all clusters
data "hopsworksai_clusters" "all" {

}

# retrieve only AWS clusters
data "hopsworksai_clusters" "awsClusters" {
  filter {
    cloud = "AWS"
  }
}
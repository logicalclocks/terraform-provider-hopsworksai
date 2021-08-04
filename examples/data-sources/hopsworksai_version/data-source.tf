# retrieve latest supported version on AWS
data "hopsworksai_version" "latest" {
  cloud_provider = "AWS"
}

# retrieve latest supported version with centos and on region us-east-2 on AWS
data "hopsworksai_version" "latest" {
  cloud_provider = "AWS"
  os             = "centos"
  region         = "us-east-2"
}

# retrieve latest supported version that is upgradeable from 2.1.0 on AWS
data "hopsworksai_version" "latest" {
  cloud_provider           = "AWS"
  upgradeable_from_version = "2.1.0"
}
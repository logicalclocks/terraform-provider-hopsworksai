variable "region" {
  type    = string
  default = "us-east-2"
}

variable "profile" {
  type    = string
  default = "default"
}

variable "eks_cluster_name" {
  type    = string
  default = "tf-hopsworks-eks-cluster"
}
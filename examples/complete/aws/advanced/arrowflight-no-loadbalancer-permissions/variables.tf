variable "region" {
  type    = string
  default = "us-east-2"
}

variable "profile" {
  type    = string
  default = "default"
}

variable "cluster_name" {
  type    = string
  default = "hopsworks-arrow"
}

variable "num_mysql_servers" {
  type    = number
  default = 1
}

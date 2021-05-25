variable "region" {
  description = "AWS Region"
  type        = string
  default     = "us-east-2"
}

variable "bucket_name" {
  description = "S3 bucket name"
  type        = string
}

variable "instance_profile_name" {
  description = "Instance profile name"
  type        = string
}

variable "ssh_key_name" {
  description = "SSH Key pair name"
  type        = string
}

variable "ssh_public_key" {
  description = "Public key used with this ssh key pair"
  type        = string
}

data "hopsworksai_aws_instance_profile_policy" "policy" {
  enable_eks_and_ecr = false
  enable_upgrade     = false
}

resource "aws_iam_role" "role" {
  name        = "${var.instance_profile_name}-role"
  description = "This is a custom role created via Terraform for Hopsworks.ai"
  assume_role_policy = jsonencode(
    {
      Version = "2012-10-17"
      Statement = [
        {
          Action = "sts:AssumeRole"
          Effect = "Allow"
          Principal = {
            Service = "ec2.amazonaws.com"
          }
        },
      ]
    }
  )

  inline_policy {
    name   = "hopsworksai"
    policy = data.hopsworksai_aws_instance_profile_policy.policy.json
  }

  tags = {
    Creator = "Terraform"
    Purpose = "Hopsworks.ai"
  }
}

resource "aws_iam_instance_profile" "profile" {
  name = var.instance_profile_name
  role = aws_iam_role.role.name

  tags = {
    Creator = "Terraform"
    Purpose = "Hopsworks.ai"
  }
}

resource "aws_s3_bucket" "bucket" {
  count         = var.num_buckets
  bucket        = "${var.bucket_name_prefix}-${count.index}"
  acl           = "private"
  force_destroy = true
  tags = {
    Creator = "Terraform"
    Purpose = "Hopsworks.ai"
  }
}

resource "aws_s3_bucket_public_access_block" "block_bucket" {
  count                   = var.num_buckets
  bucket                  = aws_s3_bucket.bucket[count.index].id
  block_public_acls       = true
  block_public_policy     = true
  restrict_public_buckets = true
  ignore_public_acls      = true
}

resource "aws_key_pair" "key" {
  key_name   = var.ssh_key_name
  public_key = var.ssh_public_key
  tags = {
    Creator = "Terraform"
    Purpose = "Hopsworks.ai"
  }
}

data "aws_availability_zones" "available" {
}

module "vpc" {
  source         = "terraform-aws-modules/vpc/aws"
  version        = "3.2.0"
  name           = var.vpc_name
  cidr           = "10.0.0.0/16"
  azs            = data.aws_availability_zones.available.names
  public_subnets = ["10.0.101.0/24"]

  enable_dns_support   = true
  enable_dns_hostnames = true

  manage_default_security_group = true
  default_security_group_ingress = [
    {
      self     = true
      protocol = "-1"
    },
    {
      from_port   = 80
      to_port     = 80
      protocol    = "tcp"
      cidr_blocks = "0.0.0.0/0"
    },
    {
      from_port   = 443
      to_port     = 443
      protocol    = "tcp"
      cidr_blocks = "0.0.0.0/0"
    },
  ]
  default_security_group_egress = [
    {
      protocol    = "-1"
      cidr_blocks = "0.0.0.0/0"
    },
  ]
}
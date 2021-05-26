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
  bucket                  = "${var.bucket_name_prefix}-${count.index}"
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
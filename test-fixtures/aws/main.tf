data "hopsworksai_aws_instance_profile_policy" "policy" {
  enable_eks_and_ecr = false
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

resource "aws_vpc" "vpc" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true
}

resource "aws_subnet" "subnet" {
  vpc_id                  = aws_vpc.vpc.id
  cidr_block              = "10.0.1.0/24"
  map_public_ip_on_launch = true
}

resource "aws_internet_gateway" "internet_gateway" {
  vpc_id = aws_vpc.vpc.id
}

resource "aws_route_table" "route_table" {
  vpc_id = aws_vpc.vpc.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.internet_gateway.id
  }
}

resource "aws_route_table_association" "route_to_subnet" {
  subnet_id      = aws_subnet.subnet.id
  route_table_id = aws_route_table.route_table.id
}

resource "aws_security_group" "security_group" {
  vpc_id = aws_vpc.vpc.id

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = -1
    cidr_blocks = ["0.0.0.0/0"]
  }
  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port = 0
    to_port   = 0
    self      = true
    protocol  = "-1"
  }
}
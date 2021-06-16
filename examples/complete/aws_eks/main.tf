terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">=3.42.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = ">= 2.3.0"
    }
    hopsworksai = {
      source = "logicalclocks/hopsworksai"
    }
  }
}

variable "region" {
  type    = string
  default = "us-east-2"
}

variable "bucket_name" {
  type    = string
  default = "tf-hopsworks-bucket"
}

variable "eks_cluster_name" {
  type    = string
  default = "tf-hopsworks-eks-cluster"
}

provider "aws" {
  region = var.region
}

provider "hopsworksai" {
}

# Step 1: create an instance profile to allow hopsworks cluster 
data "hopsworksai_aws_instance_profile_policy" "policy" {
  bucket_name = var.bucket_name
}

resource "aws_iam_role" "role" {
  name = "tf-hopsworksai-instance-profile-role"
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
}

resource "aws_iam_instance_profile" "profile" {
  name = "hopsworksai-instance-profile"
  role = aws_iam_role.role.name
}

# Step 2: create s3 bucket to be used by your hopsworks cluster to store your data
resource "aws_s3_bucket" "bucket" {
  bucket        = var.bucket_name
  acl           = "private"
  force_destroy = true
}

# Step 3: create an ssh key pair 
resource "aws_key_pair" "key" {
  key_name   = "tf-hopsworksai-key"
  public_key = file("~/.ssh/id_rsa.pub")
}

# Step 4: create vpc 
data "aws_availability_zones" "available" {
}

module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "3.1.0"

  name                 = "${var.eks_cluster_name}-vpc"
  cidr                 = "172.16.0.0/16"
  azs                  = data.aws_availability_zones.available.names
  private_subnets      = ["172.16.1.0/24", "172.16.2.0/24", "172.16.3.0/24"]
  public_subnets       = ["172.16.4.0/24", "172.16.5.0/24", "172.16.6.0/24"]
  enable_nat_gateway   = true
  single_nat_gateway   = true
  enable_dns_hostnames = true

  public_subnet_tags = {
    "kubernetes.io/cluster/${var.eks_cluster_name}" = "shared"
    "kubernetes.io/role/elb"                        = "1"
  }

  private_subnet_tags = {
    "kubernetes.io/cluster/${var.eks_cluster_name}" = "shared"
    "kubernetes.io/role/internal-elb"               = "1"
  }
}

# Step 5: create EKS cluster 
data "aws_iam_instance_profile" "profile" {
  name = aws_iam_instance_profile.profile.name
}

module "eks" {
  source          = "terraform-aws-modules/eks/aws"
  version         = "17.1.0"
  cluster_name    = var.eks_cluster_name
  cluster_version = "1.19"
  subnets         = concat(module.vpc.private_subnets, module.vpc.public_subnets)

  tags = {
    Environment = "test"
  }

  vpc_id = module.vpc.vpc_id

  node_groups_defaults = {
    ami_type  = "AL2_x86_64"
    disk_size = 100
  }

  node_groups = {
    test = {
      desired_capacity = 2
      max_capacity     = 5
      min_capacity     = 1
      instance_types   = ["m5.xlarge"]
    }
  }

  map_roles = [
    {
      rolearn  = data.aws_iam_instance_profile.profile.role_arn
      username = "hopsworks"
      groups   = ["system:masters"]
    },
  ]
}

resource "aws_security_group_rule" "http" {
  type              = "ingress"
  from_port         = 80
  to_port           = 80
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = module.eks.cluster_primary_security_group_id
}

resource "aws_security_group_rule" "https" {
  type              = "ingress"
  from_port         = 443
  to_port           = 443
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = module.eks.cluster_primary_security_group_id
}

data "aws_eks_cluster" "cluster" {
  name = module.eks.cluster_id
}

data "aws_eks_cluster_auth" "cluster" {
  name = module.eks.cluster_id
}

# Step 6: needed for configure aws_auth to succeed 
provider "kubernetes" {
  host                   = data.aws_eks_cluster.cluster.endpoint
  cluster_ca_certificate = base64decode(data.aws_eks_cluster.cluster.certificate_authority.0.data)
  exec {
    api_version = "client.authentication.k8s.io/v1alpha1"
    command     = "aws"
    args = [
      "eks",
      "get-token",
      "--cluster-name",
      data.aws_eks_cluster.cluster.name
    ]
  }
}

# Step 7: create a cluster with 1 worker 
data "hopsworksai_instance_type" "smallest_worker" {
  cloud_provider = "AWS"
  node_type      = "worker"
}

resource "hopsworksai_cluster" "cluster" {
  name    = "tf-hopsworks-cluster"
  ssh_key = aws_key_pair.key.key_name

  head {
  }

  workers {
    instance_type = data.hopsworksai_instance_type.smallest_worker.id
    count         = 1
  }

  aws_attributes {
    region               = var.region
    bucket_name          = var.bucket_name
    instance_profile_arn = aws_iam_instance_profile.profile.arn
    network {
      vpc_id            = module.vpc.vpc_id
      subnet_id         = module.vpc.public_subnets[0]
      security_group_id = module.eks.cluster_primary_security_group_id
    }
    eks_cluster_name = module.eks.cluster_id
  }

  open_ports {
    ssh = true
  }

  tags = {
    Purpose = "testing"
  }
}

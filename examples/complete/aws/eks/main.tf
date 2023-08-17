provider "aws" {
  region  = var.region
  profile = var.profile
}

provider "hopsworksai" {
}

# Step 1: Create required aws resources, an ssh key, an s3 bucket, and an instance profile with the required hopsworks permissions
module "aws" {
  source  = "logicalclocks/helpers/hopsworksai//modules/aws"
  region  = var.region
  version = "2.3.0"
}

# Step 2: create vpc 
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

# Step 3: create EKS cluster 
module "eks" {
  source          = "terraform-aws-modules/eks/aws"
  version         = "17.1.0"
  cluster_name    = var.eks_cluster_name
  cluster_version = "1.26"
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
      rolearn  = module.aws.instance_profile_role_arn
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

# Step 4: needed for configure aws_auth to succeed 
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
      data.aws_eks_cluster.cluster.name,
      "--profile",
      var.profile
    ]
  }
}

# Step 5: create a cluster with 1 worker 
data "hopsworksai_instance_type" "head" {
  cloud_provider = "AWS"
  node_type      = "head"
  region         = var.region
}

data "hopsworksai_instance_type" "rondb_mgm" {
  cloud_provider = "AWS"
  node_type      = "rondb_management"
  region         = var.region
}

data "hopsworksai_instance_type" "rondb_data" {
  cloud_provider = "AWS"
  node_type      = "rondb_data"
  region         = var.region
}

data "hopsworksai_instance_type" "rondb_mysql" {
  cloud_provider = "AWS"
  node_type      = "rondb_mysql"
  region         = var.region
}

data "hopsworksai_instance_type" "smallest_worker" {
  cloud_provider = "AWS"
  node_type      = "worker"
  region         = var.region
  min_cpus       = 8
}

resource "hopsworksai_cluster" "cluster" {
  name    = "tf-hopsworks-cluster"
  ssh_key = module.aws.ssh_key_pair_name

  head {
    instance_type = data.hopsworksai_instance_type.head.id
  }

  workers {
    instance_type = data.hopsworksai_instance_type.smallest_worker.id
    count         = 1
  }

  aws_attributes {
    region = var.region
    bucket {
      name = module.aws.bucket_name
    }
    instance_profile_arn = module.aws.instance_profile_arn
    network {
      vpc_id            = module.vpc.vpc_id
      subnet_id         = module.vpc.public_subnets[0]
      security_group_id = module.eks.cluster_primary_security_group_id
    }
    eks_cluster_name = module.eks.cluster_id
  }

  rondb {
    management_nodes {
      instance_type = data.hopsworksai_instance_type.rondb_mgm.id
    }
    data_nodes {
      instance_type = data.hopsworksai_instance_type.rondb_data.id
    }
    mysql_nodes {
      instance_type = data.hopsworksai_instance_type.rondb_mysql.id
    }
  }

  open_ports {
    ssh = true
  }

  tags = {
    Purpose = "testing"
  }
}

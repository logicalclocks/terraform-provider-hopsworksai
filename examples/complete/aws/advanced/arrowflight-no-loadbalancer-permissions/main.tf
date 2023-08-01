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


# Step 2: Create a VPC 
data "aws_availability_zones" "available" {
}

module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "3.1.0"

  name                 = "${var.cluster_name}-vpc"
  cidr                 = "172.16.0.0/16"
  azs                  = data.aws_availability_zones.available.names
  public_subnets       = ["172.16.4.0/24"]
  enable_dns_hostnames = true
}

# Step 3: Create a security group and open required ports
resource "aws_security_group" "security_group" {
  name        = "${var.cluster_name}-security-group"
  description = "Allow access for Hopsworks cluster"
  vpc_id      = module.vpc.vpc_id

  ingress {
    description = "HTTPS"
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "HTTP"
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "MYSQL"
    from_port   = 3306
    to_port     = 3306
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "ArrowFlight"
    from_port   = 5005
    to_port     = 5005
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "HiveServer"
    from_port   = 9085
    to_port     = 9085
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "HiveMetastore"
    from_port   = 9083
    to_port     = 9083
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "Kafka"
    from_port   = 9092
    to_port     = 9092
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port = 0
    to_port   = 0
    protocol  = -1
    self      = true
  }

  egress {
    from_port        = 0
    to_port          = 0
    protocol         = "-1"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }
}


# Step 3: Create a network load balancer 
resource "aws_lb" "lb" {
  name               = "${var.cluster_name}-lb"
  internal           = false
  load_balancer_type = "network"
  subnets            = module.vpc.public_subnets
}

# Step 4: create a cluster with 1 worker 
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
  min_cpus       = 8
  min_memory_gb  = 16
}

data "hopsworksai_instance_type" "smallest_worker" {
  cloud_provider = "AWS"
  node_type      = "worker"
  region         = var.region
  min_cpus       = 8
}

resource "hopsworksai_cluster" "cluster" {
  name    = var.cluster_name
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
      security_group_id = aws_security_group.security_group.id
    }
  }

  rondb {
    configuration {
      ndbd_default {
        replication_factor = 2
      }
    }

    management_nodes {
      instance_type = data.hopsworksai_instance_type.rondb_mgm.id
      disk_size     = 30
    }
    data_nodes {
      instance_type = data.hopsworksai_instance_type.rondb_data.id
      count         = 2
      disk_size     = 512
    }
    mysql_nodes {
      instance_type            = data.hopsworksai_instance_type.rondb_mysql.id
      count                    = var.num_mysql_servers
      disk_size                = 256
      arrow_flight_with_duckdb = true
    }
  }

  init_script = <<EOF
  #!/usr/bin/env bash
  set -e
  IS_MASTER=`grep -c "INSTANCE_TYPE=master" /var/lib/cloud/instance/scripts/part-001`
  if [[ $IS_MASTER != 2 ]];
  then
    exit 0
  fi
  /srv/hops/mysql-cluster/ndb/scripts/mysql-client.sh hopsworks -e  "UPDATE variables SET value='${aws_lb.lb.dns_name}' WHERE id='loadbalancer_external_domain';"
  EOF

}

# Step 5: Create target groups and register them with the load balancer 

data "aws_instances" "mysqld" {
  instance_tags = {
    Name = "${hopsworksai_cluster.cluster.name}-mysqld*"
  }
}

resource "aws_lb_target_group" "mysqld_target_group" {
  name     = "${hopsworksai_cluster.cluster.name}-mysqld"
  port     = 3306
  protocol = "TCP"
  vpc_id   = hopsworksai_cluster.cluster.aws_attributes[0].network[0].vpc_id
  health_check {
    enabled  = true
    protocol = "TCP"
  }
}

resource "aws_lb_target_group_attachment" "mysqld_target_group" {
  count            = var.num_mysql_servers
  target_group_arn = aws_lb_target_group.mysqld_target_group.arn
  target_id        = data.aws_instances.mysqld.ids[count.index]
  port             = 3306
}

resource "aws_lb_target_group" "arrow_flight_target_group" {
  name     = "${hopsworksai_cluster.cluster.name}-arrowflight"
  port     = 5005
  protocol = "TCP"
  vpc_id   = hopsworksai_cluster.cluster.aws_attributes[0].network[0].vpc_id
  health_check {
    enabled  = true
    protocol = "TCP"
  }
}

resource "aws_lb_target_group_attachment" "arrow_flight_target_group" {
  count            = var.num_mysql_servers
  target_group_arn = aws_lb_target_group.arrow_flight_target_group.arn
  target_id        = data.aws_instances.mysqld.ids[count.index]
  port             = 5005
}

resource "aws_lb_target_group" "hiveserver_target_group" {
  name     = "${hopsworksai_cluster.cluster.name}-hiveserver"
  port     = 9085
  protocol = "TCP"
  vpc_id   = hopsworksai_cluster.cluster.aws_attributes[0].network[0].vpc_id
  health_check {
    enabled  = true
    protocol = "TCP"
  }
}

resource "aws_lb_target_group_attachment" "hiveserver_target_group" {
  target_group_arn = aws_lb_target_group.hiveserver_target_group.arn
  target_id        = hopsworksai_cluster.cluster.head.0.node_id
  port             = 9085
}

resource "aws_lb_target_group" "hivemetastore_target_group" {
  name     = "${hopsworksai_cluster.cluster.name}-hivemetastore"
  port     = 9083
  protocol = "TCP"
  vpc_id   = hopsworksai_cluster.cluster.aws_attributes[0].network[0].vpc_id
  health_check {
    enabled  = true
    protocol = "TCP"
  }
}

resource "aws_lb_target_group_attachment" "hivemetastore_target_group" {
  target_group_arn = aws_lb_target_group.hivemetastore_target_group.arn
  target_id        = hopsworksai_cluster.cluster.head.0.node_id
  port             = 9083
}


resource "aws_lb_listener" "mysqld" {
  load_balancer_arn = aws_lb.lb.arn
  protocol          = "TCP"
  port              = 3306
  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.mysqld_target_group.arn
  }
}

resource "aws_lb_listener" "arrowflight" {
  load_balancer_arn = aws_lb.lb.arn
  protocol          = "TCP"
  port              = 5005
  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.arrow_flight_target_group.arn
  }
}

resource "aws_lb_listener" "hiveserver" {
  load_balancer_arn = aws_lb.lb.arn
  protocol          = "TCP"
  port              = 9085
  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.hiveserver_target_group.arn
  }
}

resource "aws_lb_listener" "hivemetastore" {
  load_balancer_arn = aws_lb.lb.arn
  protocol          = "TCP"
  port              = 9083
  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.hivemetastore_target_group.arn
  }
}
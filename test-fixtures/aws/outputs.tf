output "region" {
  value = var.region
}

output "instance_profile_arn" {
  value = aws_iam_instance_profile.profile.arn
}

output "bucket_names" {
  value = join(",", aws_s3_bucket.bucket.*.id)
}

output "ssh_key_name" {
  value = aws_key_pair.key.key_name
}

output "vpc_id" {
  value = module.vpc.vpc_id
}

output "subnet_id" {
  value = module.vpc.public_subnets[0]
}

output "security_group_id" {
  value = module.vpc.default_security_group_id
}
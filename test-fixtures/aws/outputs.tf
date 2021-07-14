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
  value = aws_vpc.vpc.id
}

output "subnet_id" {
  value = aws_subnet.subnet.id
}

output "security_group_id" {
  value = aws_security_group.security_group.id
}
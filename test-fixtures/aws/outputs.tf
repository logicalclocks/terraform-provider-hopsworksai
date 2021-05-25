output "region" {
  value = var.region
}

output "instance_profile_arn" {
  value = aws_iam_instance_profile.profile.arn
}

output "bucket_name" {
  value = aws_s3_bucket.bucket.id
}

output "ssh_key_name" {
  value = aws_key_pair.key.key_name
}

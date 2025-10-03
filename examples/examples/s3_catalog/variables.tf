variable "TESTING_AWS_ACCESS_KEY" {
  type        = string
  sensitive   = true
  description = "AWS S3 access key"
}

variable "TESTING_AWS_SECRET_KEY" {
  type        = string
  sensitive   = true
  description = "AWS S3 secret key"
}

terraform {
  required_providers {
    galaxy = {
      source = "hashicorp.com/starburstdata/galaxy"
    }
  }
}

provider "galaxy" {
  # Credentials from environment variables
}

# Create a service account first
resource "galaxy_service_account" "example" {
  username              = "exampleserviceaccount"
  with_initial_password = false # Don't create initial password
  additional_role_ids   = []
}

# Create a password for the service account
resource "galaxy_service_account_password" "example_password" {
  service_account_id = galaxy_service_account.example.id
  description        = "Production API access password"
}

# Create another password for rotation
resource "galaxy_service_account_password" "rotation_password" {
  service_account_id = galaxy_service_account.example.id
  description        = "Rotation password for zero-downtime updates"
}

# Data source to read service account (includes password info)
data "galaxy_service_account" "example" {
  depends_on = [
    galaxy_service_account_password.example_password,
    galaxy_service_account_password.rotation_password
  ]
  id = galaxy_service_account.example.id
}

output "service_account_id" {
  value = galaxy_service_account.example.id
}

output "password_id" {
  value = galaxy_service_account_password.example_password.service_account_password_id
}

output "password_prefix" {
  value     = galaxy_service_account_password.example_password.password_prefix
  sensitive = false
}

output "password" {
  value     = galaxy_service_account_password.example_password.password
  sensitive = true
}

output "rotation_password_id" {
  value = galaxy_service_account_password.rotation_password.service_account_password_id
}

output "all_passwords" {
  value     = data.galaxy_service_account.example.passwords
  sensitive = true
}
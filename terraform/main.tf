# Specify the GCP Provider
provider "google" {
  project = var.project_id
  region  = var.location
}

# Cloud Run for provision-service
module "cloud-run-service" {
  source  = "GoogleCloudPlatform/cloud-run/google"
  version = "~> 0.5.0"

  service_name = var.service_name
  project_id   = var.project_id
  location     = var.location
  image        = local.service_image

  # Service parameters
  service_account_email = module.gsa-service.email
  
  # Environment variables
  env_vars = local.env_vars

  # Allow invoke
  members = [
    "serviceAccount:${module.gsa-service.email}"
  ]
}

locals {
  # Cloud Run locals
  service_image              = "${var.location}-docker.pkg.dev/${var.project_id}/${var.storage_artifact_registry_id}/${var.service_name}:latest"
  # client_image               = "${var.location}-docker.pkg.dev/${var.project_id}/${var.storage_artifact_registry_id}/${var.service_name}-client:latest"
  # service_image = "gcr.io/cloudrun/hello"
  client_image  = "${var.location}-docker.pkg.dev/${var.project_id}/${var.storage_artifact_registry_id}/${var.service_name}-client:latest"
  env_vars = [
    {
      name  = "RCB_PORT",
      value = "8080"
    },
    {
      name  = "RCB_HOST",
      value = "0.0.0.0"
    },
    {
      name  = "RCB_SERVICE_PROJECTID",
      value = var.project_id
    },
    {
      name  = "RCB_SERVICE_REGION",
      value = var.location
    },
    {
      name  = "RCB_SERVICE_NAME",
      value = var.service_name
    },
    {
      name  = "RCB_SERVICE_IDENTITY",
      value = module.gsa-service.email
    },
  ]
}

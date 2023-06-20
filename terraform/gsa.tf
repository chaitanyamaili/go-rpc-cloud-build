# Identity service account for cloud run.
module "gsa-service" {
  source  = "terraform-google-modules/service-accounts/google"
  version = "4.2.0"

  project_id = var.project_id
  names      = ["${var.service_name}"]

  display_name = "${var.service_name} identity"
  description  = "Service identity for ${var.service_name} Cloud Run instance"

  project_roles = [
    # Create builds and triggers.
    "${var.project_id}=>roles/cloudbuild.builds.builder",
  ]
}

# resource "google_artifact_registry_repository_iam_member" "iam-artifact-registry" {
#   project  = var.project_id
#   location = var.location

#   repository = var.storage_artifact_registry_id

#   role   = "roles/artifactregistry.reader"
#   member = "serviceAccount:${module.gsa-service.email}"

#   depends_on = [
#     module.gsa-service,
#   ]
# }

resource "google_service_account_iam_member" "iam-service-user" {
  service_account_id = "projects/${var.project_id}/serviceAccounts/${module.gsa-service.email}"
  role               = "roles/iam.serviceAccountUser"
  member             = "serviceAccount:${module.gsa-service.email}"

  depends_on = [
    module.gsa-service,
  ]
}

# Allow the default GSA Cloud Run to read the Artifact Registry
resource "google_project_iam_member" "project-iam-bindings" {
  project = var.project_id

  role   = "roles/artifactregistry.reader"
  member = "serviceAccount:service-${var.project_number}@serverless-robot-prod.iam.gserviceaccount.com"
}


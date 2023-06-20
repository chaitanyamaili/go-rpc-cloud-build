resource "google_artifact_registry_repository" "decoupled_artifact_registry" {
  project  = var.project_id
  location = var.location

  repository_id = var.storage_artifact_registry_id

  description = "Docker Repository"
  format      = "DOCKER"
}

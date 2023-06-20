variable "project_id" {
  description = "Google project ID."
  type        = string
}

variable "project_number" {
  description = "Google project number."
  type        = string
}

variable "storage_artifact_registry_id" {
  description = "The artifact registory name."
  type        = string
  default     = "artifact-registry"
}

variable "location" {
  description = "Google cloud region."
  type        = string
  default     = "us-central1"
}

variable "service_name" {
  description = "Service name."
  type        = string
}
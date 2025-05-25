variable "container_image" {
  description = "Container image"
  type        = string
  default     = "ghcr.io/lemigu/terraform-gcp-teams-notification-channel:latest"
}

variable "webhook_url" {
  description = "MS Teams Webhook URL"
  type        = string
}

variable "environment" {
  description = "Deployment Environment"
  type        = string
  default     = "prod"
}

variable "cloudrun_region" {
  description = "Cloudrun Deployment Region"
  type        = string
  default     = "europe-west1"
}

variable "project_id" {
  description = "GCP Project ID"
  type        = string
}

resource "google_cloud_run_service" "teams_notifier" {
  project  = var.project_id
  name     = "cr-teams-notification-${var.environment}"
  location = var.cloudrun_region

  template {
    spec {
      service_account_name = google_service_account.notifier.email

      containers {
        image = var.container_image

        env {
          name  = "WEBHOOK_URL"
          value = var.webhook_url
        }
      }
    }
  }

  traffic {
    percent         = 100
    latest_revision = true
  }
}

resource "google_pubsub_topic" "teams_notification_topic" {
  project = var.project_id
  name    = "teams-nofitication-topic-${var.environment}"
}

resource "google_pubsub_topic_iam_member" "monitoring_alerts_publisher" {
  project = var.project_id
  topic   = google_pubsub_topic.teams_notification_topic.name

  role   = "roles/pubsub.publisher"
  member = "serviceAccount:service-${data.google_project.project.number}@gcp-sa-monitoring-notification.iam.gserviceaccount.com"
}

resource "google_monitoring_notification_channel" "teams_notification_channel" {
  project      = var.project_id
  display_name = "Teams Notification Channel (${var.environment})"
  type         = "pubsub"
  labels = {
    "topic" = google_pubsub_topic.teams_notification_topic.id
  }
}

resource "google_service_account" "teams_notifier_invoker" {
  project      = var.project_id
  account_id   = "notifier-invoker-${var.environment}"
  display_name = "Teams Notifier Invoker"
}

resource "google_service_account" "notifier" {
  project      = var.project_id
  account_id   = "teams-notifier-${var.environment}"
  display_name = "Teams Notifier"
}

resource "google_cloud_run_service_iam_member" "teams_notifier_invoker_member" {
  project  = var.project_id
  service  = google_cloud_run_service.teams_notifier.name
  location = google_cloud_run_service.teams_notifier.location
  role     = "roles/run.invoker"
  member   = "serviceAccount:${google_service_account.teams_notifier_invoker.email}"
}

resource "google_project_iam_member" "notifier_logwriter" {
  project = var.project_id
  role    = "roles/logging.logWriter"
  member  = "serviceAccount:${google_service_account.notifier.email}"
}

resource "google_pubsub_subscription" "teams_notification_subscription" {
  project = var.project_id
  name    = "teams-notification-subscription-${var.environment}"
  topic   = google_pubsub_topic.teams_notification_topic.name

  push_config {
    push_endpoint = google_cloud_run_service.teams_notifier.status[0].url
    oidc_token {
      service_account_email = google_service_account.teams_notifier_invoker.email
    }
  }
}

output "channel_id" {
  description = "Teams notification channel"
  value       = google_monitoring_notification_channel.teams_notification_channel.id
}

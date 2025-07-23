output "app_url" {
  value       = aws_apprunner_service.service.service_url
  description = "Public URL of the shortening service"
}

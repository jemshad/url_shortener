variable "app_name" {
  default     = "url-shortener"
  type        = string
  description = "Name of the Application"
}

variable "aws_region" {
  default     = "eu-central-1"
  type        = string
  description = "AWS Region"
}

variable "container_port" {
  description = "Port the container listens on"
  type        = string
  default     = "8080"
}

variable "auth_user" {
  description = "Basic auth username"
  type        = string
  default     = "admin"
}

variable "auth_pass" {
  description = "Basic auth password"
  type        = string
  sensitive   = true # hide from terraform plan/apply outputs
}
